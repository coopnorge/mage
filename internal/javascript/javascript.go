package javascript

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/coopnorge/mage/internal/core"
	"github.com/magefile/mage/sh"
)

// Lint checks for the biome config file and runs the linting in a docker container
// Prints error and exits
func Lint() error {
	if !core.FileExists("biome.json") {
		return errors.New("biome not setup in your project. Install @coopnorge/web-devtools")
	}
	return devtoolBiomeLint()
}

// PublishLib checks if package.json file exists or not, checks if distribution/build-output folder
// exists or not, checks if .npmrc file exits or not
func PublishLib() error {
	githubToken := os.Getenv("GITHUB_TOKEN")
	privateEnv := os.Getenv("PRIVATE")
	isPrivate := strings.ToLower(privateEnv) == "true" || privateEnv == "1"
	githubTagname := os.Getenv("GITHUB_TAGNAME")
	skipBuild := os.Getenv("SKIP_BUILD")
	buildCommand := os.Getenv("BUILD_COMMAND")
	newVersion := strings.TrimPrefix(githubTagname, "v")

	distDir := "dist"

	access := "public"

	if isPrivate {
		access = "restricted"
	}

	if newVersion == "" {
		return errors.New("no new package version set. Set GITHUB_TAGNAME env variable")
	}

	if skipBuild == "" {
		isDistDirEmpty, errOnCheckDistDir := core.IsDirectoryEmpty(distDir)

		if isDistDirEmpty {
			return errors.New("no build files to publish")
		}

		if errOnCheckDistDir != nil {
			return errOnCheckDistDir
		}
	}

	if !core.FileExists(".npmrc") {
		return errors.New(".npmrc file missing")
	}

	if !IsNpmrcValidForPublish(".") {
		return errors.New(".npmrc has no auth configuration")
	}

	if !core.FileExists("package.json") {
		return errors.New("not a js node project")
	}

	// Run build if build command is set or skip build is not set.
	// Some pakages won't need to be build
	if buildCommand != "" || skipBuild == "" {
		if buildCommand == "" {
			buildCommand = "build"
		}
		buildCommand = fmt.Sprintf("npm install && npm run %s", buildCommand)
	}

	commands := fmt.Sprintf("%s && npm version %s && npm publish --access %s", buildCommand, newVersion, access)

	return devtoolPublishNpmLib(commands, githubToken)
}

// IsNpmrcValidForPublish checks if the .npmrc file is configured for GitHub
// Packages.
func IsNpmrcValidForPublish(directory string) bool {
	if directory == "" {
		directory = "."
	}

	registryURL := "npm.pkg.github.com"
	scope := "@coopnorge"
	tokenIndicator := "_authToken="

	npmrcContent, err := os.ReadFile(fmt.Sprintf("%s/.npmrc", directory))

	if err != nil {
		return false
	}

	contentStr := string(npmrcContent)

	if !strings.Contains(contentStr, registryURL) && !strings.Contains(contentStr, scope) && !strings.Contains(contentStr, tokenIndicator) {
		return false
	}

	return true
}

func devtoolBiomeLint() error {
	// Get the current working directory to mount it.
	cwd, err := os.Getwd()

	if err != nil {
		return err
	}

	dockerArgs := []string{
		"--volume", fmt.Sprintf("%s:/app", cwd),
		"--workdir", "/app",
	}

	return Run("ghcr.io/biomejs/biome:1.8.3", dockerArgs, "lint")
}

func devtoolPublishNpmLib(commands string, githubToken string) error {
	// Get the current working directory to mount it.
	cwd, err := os.Getwd()

	if err != nil {
		return err
	}

	dockerArgs := []string{
		"-e", fmt.Sprintf("GITHUB_TOKEN=%s", githubToken),
		"--volume", fmt.Sprintf("%s:/app", cwd),
		"--workdir", "/app",
	}

	return Run("node:slim", dockerArgs, "sh", "-c", commands)
}

// Run will run the specified command with arguments in the
// specified Docker image
func Run(image string, dockerRunArgs []string, args ...string) error {
	call := []string{
		"run",
		"--rm",
	}

	call = append(call, dockerRunArgs...)
	call = append(call, image)
	call = append(call, args...)
	return sh.RunV("docker", call...)
}
