package goapp

import (
	"context"
	_ "embed"
	"encoding/json"
	"os"
	"path"
	"strconv"

	"github.com/coopnorge/mage/internal/core"
	"github.com/coopnorge/mage/internal/docker"
	"github.com/coopnorge/mage/internal/golang"

	"github.com/magefile/mage/mg"
)

//go:embed app.Dockerfile
var dockerfile string

// Docker is the magefile namespace to group Docker commands
type Docker mg.Namespace

// BuildAndPush OCI image. Setting the PUSH_IMAGE environmental variable to true will push the images to the
// registries. When push is true images are not tagged with latest.
//
// Given the input:
//
//	./var
//	├── app1
//	│   └── bin
//	│       ├── darwin
//	│       │   └── arm64
//	│       │       ├── dataloader
//	│       │       └── server
//	│       └── linux
//	│           ├── amd64
//	│           │   ├── dataloader
//	│           │   └── server
//	│           └── arm64
//	│               ├── dataloader
//	│               └── server
//	└── app2
//	    └── bin
//	        ├── darwin
//	        │   └── arm64
//	        │       ├── dataloader
//	        │       └── server
//	        └── linux
//	            ├── amd64
//	            │   ├── dataloader
//	            │   └── server
//	            └── arm64
//	                ├── dataloader
//	                └── server
//	                └── server
//
// [Docker.BuildAndPush] will create:
//
//	./var
//	├── oci-images.json
//	├── app1
//	│   └── oci
//	│       ├── dataloader
//	│       │   ├── image.tar
//	│       │   └── metadata.json
//	│       └── server
//	│           ├── image.tar
//	│           └── metadata.json
//	└── app2
//	    └── oci
//	        ├── dataloader
//	        │   ├── image.tar
//	        │   └── metadata.json
//	        └── server
//	            ├── image.tar
//	            └── metadata.json
//
// oci-images.json will contain a map over the images and tags per app and
// binary.
//
//	{
//	  "app1": {
//	    "dataloader": {
//	      "image": "ocreg.invalid/coopnorge/app1/dataloader:v2025.03.11135857",
//	      "tag": "v2025.03.11135857"
//	    },
//	    "server": {
//	      "image": "ocreg.invalid/coopnorge/app1/server:v2025.03.11135857",
//	      "tag": "v2025.03.11135857"
//	    }
//	  }
//	  "app2": {
//	    "dataloader": {
//	      "image": "ocreg.invalid/coopnorge/app2/dataloader:v2025.03.11135857",
//	      "tag": "v2025.03.11135857"
//	    },
//	    "server": {
//	      "image": "ocreg.invalid/coopnorge/app2/server:v2025.03.11135857",
//	      "tag": "v2025.03.11135857"
//	    }
//	  }
//	}
func (Docker) BuildAndPush(ctx context.Context) error {
	mg.SerialCtxDeps(ctx, Go.Build, Docker.BuildImages)
	return nil
}

// BuildImages just builds docker images. It expects the
// binaries to present in the ./var/bin/ directories.
// Setting the PUSH_IMAGE environmental variable to true will push the images to the
// registries.
func (Docker) BuildImages(ctx context.Context) error {
	shouldPush, err := shouldPush()
	if err != nil {
		return err
	}
	goModules, err := golang.FindGoModules(".")
	if err != nil {
		return err
	}

	cmds, err := findCommands(goModules)
	if err != nil {
		return err
	}

	deps := []any{}
	for _, cmd := range cmds {
		for _, binary := range cmd.binaries {
			deps = append(deps, mg.F(buildAndPush, cmd.goModule, binary, shouldPush))
		}
	}
	mg.CtxDeps(ctx, deps...)

	return writeImageMetadata()
}

func buildAndPush(_ context.Context, app, binary string, shouldPush bool) error {
	imageName := docker.FullyQualifiedlImageName(app, binary)
	imagePath := imagePath(app, binary)
	metadataPath := metadataPath(app, binary)

	return docker.BuildAndPush(dockerfile, golang.DockerPlatforms(), imageName, ".", imagePath, metadataPath, app, binary, shouldPush)
}

func writeImageMetadata() error {
	images, err := docker.Images(core.OutputDir)
	if err != nil {
		return err
	}

	jsonString, err := json.Marshal(images)
	if err != nil {
		return err
	}
	return os.WriteFile(path.Join(core.OutputDir, "oci-images.json"), jsonString, 0o644)
}

// Validate Dockerfiles
func (Docker) Validate(_ context.Context) error {
	return docker.Validate(dockerfile)
}

func imageDir(app, binary string) string {
	return path.Join(core.OutputDir, app, "oci", binary)
}

func imagePath(app, binary string) string {
	return path.Join(imageDir(app, binary), "image.tar")
}

func metadataPath(app, binary string) string {
	return path.Join(imageDir(app, binary), "metadata.json")
}

func shouldPush() (bool, error) {
	val, ok := os.LookupEnv(PushEnv)
	if !ok || val == "" {
		return false, nil
	}
	boolValue, err := strconv.ParseBool(val)
	if err != nil {
		return false, err
	}
	return boolValue, nil
}
