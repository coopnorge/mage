package jsapp

import (
	"context"
	_ "embed"
	"encoding/json"
	"os"
	"path"

	"github.com/coopnorge/mage/internal/core"
	"github.com/coopnorge/mage/internal/docker"
	"github.com/coopnorge/mage/internal/git"
	"github.com/coopnorge/mage/internal/javascript"
	"github.com/magefile/mage/mg"
)

var (
	//go:embed app.Dockerfile
	dockerfile string
)

const (
	platforms = "linux/amd64,linux/arm64"
)

// JSApp is the magefile namespace to group JSAPP commands
type JSApp mg.Namespace

// BuildAndPush OCI image. Setting push to true will push the images to the
// registries. When push is true images are not tagged with latest.
//
// [BuildApp] will create:
//
//	./var
//	├── oci-images.json
//	└── app
//		└── oci
//	       ├── production
//	       │   ├── image.tar
//	       │   └── metadata.json
//	       └── testing
//	           ├── image.tar
//	           └── metadata.json
//
// oci-images.json will contain a map over the images and tags for app per
// environment. Use case: We add data-test-id for automating browser testing.
// These are quite a lot of ids and we remove them for production build/env.
//
//	{
//	  "app": {
//	    "testing": {
//	      "image": "ocreg.invalid/coopnorge/app/testing:v2025.03.11135857",
//	      "tag": "v2025.03.11135857"
//	    },
//	    "production": {
//	      "image": "ocreg.invalid/coopnorge/app1/production:v2025.03.11135857",
//	      "tag": "v2025.03.11135857"
//	    }
//	  }
//	}

// BuildApp creates deployable artifacts from the source code in the repository,
// to push the resulting images set the environmental variable PUSH_IMAGE to
// true. Setting PUSH_IMAGE to true will disable the latest image tag.
func (JSApp) BuildApp(ctx context.Context) error {
	shouldPush, err := docker.ShouldPush()
	if err != nil {
		return err
	}

	mg.SerialCtxDeps(ctx, JSApp.Validate, mg.F(buildAndPush, shouldPush))
	return writeImageMetadata()
}

// Lint checks all javascript/typescript codd for code standards and formats
//
// See [javascript.Lint] for details.
func (JSApp) Lint(ctx context.Context) error {
	mg.CtxDeps(ctx, javascript.Lint)
	return nil
}

// Validate Dockerfiles
func (JSApp) Validate(_ context.Context) error {
	return docker.Validate(dockerfile)
}

func buildAndPush(shouldPush bool) error {
	env := os.Getenv("DEPLOY_ENV")

	if env == "" {
		env = "production"
	}

	app, err := git.RepoNameFromURL()

	if err != nil {
		return err
	}

	imageName := docker.FullyQualifiedlImageName(app, env)
	imagePath := imagePath(app, env)
	metadataPath := metadataPath(app, env)

	return docker.BuildAndPush(dockerfile, platforms, imageName, ".", imagePath, metadataPath, app, env, shouldPush)
}

func imageDir(app string, env string) string {
	return path.Join(core.OutputDir, app, "oci", env)
}

func imagePath(app string, env string) string {
	return path.Join(imageDir(app, env), "image.tar")
}

func metadataPath(app string, env string) string {
	return path.Join(imageDir(app, env), "metadata.json")
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
