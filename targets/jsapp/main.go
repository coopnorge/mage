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

// JSAPP is the magefile namespace to group JSAPP commands
type JSAPP mg.Namespace

// BuildApp creates deployable artifacts from the source code in the repository,
// to push the resulting images set the environmental variable PUSH_IMAGE to
// true. Setting PUSH_IMAGE to true will disable the latest image tag.
func (JSAPP) BuildApp(ctx context.Context) error {
	shouldPush, err := docker.ShouldPush()
	if err != nil {
		return err
	}

	mg.SerialCtxDeps(ctx, JSAPP.Validate, mg.F(buildAndPush, shouldPush))
	return writeImageMetadata()
}

// Lint checks all javascript/typescript codd for code standards and formats
//
// See [javascript.Lint] for details.
func (JSAPP) Lint(ctx context.Context) error {
	mg.CtxDeps(ctx, javascript.Lint)
	return nil
}

// Validate Dockerfiles
func (JSAPP) Validate(_ context.Context) error {
	return docker.Validate(dockerfile)
}

func buildAndPush(shouldPush bool) error {
	app := git.RepoNameFromURL()
	imageName := docker.FullyQualifiedlImageName(app, "")
	imagePath := imagePath(app)
	metadataPath := metadataPath(app)

	return docker.BuildAndPush(dockerfile, platforms, imageName, ".", imagePath, metadataPath, app, "nodejsapp", shouldPush)
}

func imageDir(app string) string {
	return path.Join(core.OutputDir, app, "oci")
}

func imagePath(app string) string {
	return path.Join(imageDir(app), "image.tar")
}

func metadataPath(app string) string {
	return path.Join(imageDir(app), "metadata.json")
}

func writeImageMetadata() error {
	images, err := getImageMetadata(core.OutputDir)
	if err != nil {
		return err
	}

	jsonString, err := json.Marshal(images)
	if err != nil {
		return err
	}
	return os.WriteFile(path.Join(core.OutputDir, "oci-images.json"), jsonString, 0644)
}

type binaryImage = map[string]string
type binaryImages = map[string]binaryImage

func getImageMetadata(imageDir string) (binaryImages, error) {
	metadataFiles, err := docker.FindMetadataFiles(imageDir)
	if err != nil {
		return nil, err
	}

	result := make(binaryImages)
	for _, file := range metadataFiles {
		metadata, err := docker.ParseMetadata(file)
		if err != nil {
			return nil, err
		}
		if _, ok := result[metadata.App]; !ok {
			result[metadata.App] = make(binaryImage)
		}
		result[metadata.App]["tag"] = metadata.Tag
		result[metadata.App]["image"] = metadata.ImageName
	}

	return result, nil
}
