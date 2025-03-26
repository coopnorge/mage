package goapp

import (
	"context"
	_ "embed"
	"encoding/json"
	"os"
	"path"

	"github.com/coopnorge/mage/internal/core"
	"github.com/coopnorge/mage/internal/docker"
	"github.com/coopnorge/mage/internal/golang"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

const (
	platforms = "linux/amd64,linux/arm64"
)

var (
	//go:embed app.Dockerfile
	dockerfile string
)

// Docker is the magefile namespace to group Docker commands
type Docker mg.Namespace

// BuildAndPush OCI image. Setting push to true will push the images to the
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
func (Docker) BuildAndPush(ctx context.Context, shouldPush bool) error {
	mg.CtxDeps(ctx, Go.Build)

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
		deps = append(deps, mg.F(buildAndPush, cmd.goModule, cmd.binary, shouldPush))
	}
	mg.CtxDeps(ctx, deps...)

	return writeImageMetadata()
}

func buildAndPush(_ context.Context, app, binary string, shouldPush bool) error {
	imageName := docker.FullyQualifiedlImageName(app, binary)
	imagePath := imagePath(app, binary)
	metadataPath := metadataPath(app, binary)

	return docker.BuildAndPush(dockerfile, platforms, imageName, ".", imagePath, metadataPath, app, binary, shouldPush)
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
	return os.WriteFile(path.Join(core.OutputDir, "oci-images.json"), jsonString, 0644)
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
