package javascript

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/coopnorge/mage/internal/core"
	"github.com/magefile/mage/sh"
)

const (
	// PushEnv is the name of the environmental variable used to trigger
	// pushing of OCI images. Set PUSH_IMAGE to true to push images.
	PushEnv = "PUSH_IMAGE"
)

// Lint checks for the biome config file and runs the linting in a docker container
// Prints error and exits
func Lint() error {
	if core.FileExists("biome.json") {
		// Get the current working directory to mount it.
		cwd, err := os.Getwd()

		if err == nil {
			return sh.RunV(
				"docker", "run", "--rm",
				"-v", fmt.Sprintf("%s:/app", cwd),
				"ghcr.io/biomejs/biome:1.8.3",
				"lint", "/app",
			)
		}
	} else {
		return errors.New("biome not setup in your project. Install @coopnorge/web-devtools")
	}

	return nil
}

// PublishLib checks if package.json file exists or not, checks if distribution/build-output folder
// exists or not, checks if .npmrc file exits or not
func PublishLib(shouldBuild bool, buildCommand string) error {
	githubToken := os.Getenv("GITHUB_TOKEN")
	isPrivate := os.Getenv("PRIVATE")
	distDir := os.Getenv("DIST_DIR")
	githubTagname := os.Getenv("GITHUB_TAGNAME")
	newVersion := strings.TrimPrefix(githubTagname, "v")

	if distDir == "" {
		distDir = "dist"
	}

	access := "public"

	if isPrivate != "" {
		access = "private"
	}

	if newVersion == "" {
		return errors.New("no new package version set. Set PACKAGE_VERSION env variable")
	}

	isDistDirEmpty, errOnCheckDistDir := core.IsDirectoryEmpty(distDir)

	if isDistDirEmpty {
		return errors.New("no build files to publish")
	}

	if errOnCheckDistDir != nil {
		return errOnCheckDistDir
	}

	if !core.FileExists(".npmrc") {
		return errors.New(".npmrc file missing")
	}

	if !core.IsNpmrcValidForPublish(".") {
		return errors.New(".npmrc has no auth configuration")
	}

	if !core.FileExists("package.json") {
		return errors.New("not a js node project")
	}

	if shouldBuild {
		if buildCommand == "" {
			buildCommand = "build"
		}
		buildCommand = fmt.Sprintf("npm ci && npm run %s", buildCommand)
	}

	return sh.RunV(
		"docker", "run", "--rm",
		"-e", fmt.Sprintf("GITHUB_TOKEN=%s", githubToken),
		"-v", "./:/app",
		"node:slim",
		"sh",
		"-c",
		fmt.Sprintf("cd /app %s && npm version %s && npm publish --access %s", buildCommand, newVersion, access),
	)
}
