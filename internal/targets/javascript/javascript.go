package javascript

import (
	"log"
	"io"
	"fmt"
	"os"
	"strings"

	"github.com/magefile/mage/sh"
	"github.com/coopnorge/mage/internal/core"
)

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
     log.Println("biome not setup in your project. install @coopnorge/web-devtools")
	}

	return nil
}

// Checks if package.json file exists or not, checks if distribution/build-output folder
// exists or not, checks if .npmrc file exits or not
func PublishLib() error {
	githubToken := os.Getenv("GITHUB_TOKEN")
	isPrivate := os.Getenv("PRIVATE")
	distDir := os.Getenv("DIST_DIR")
	githubTagname := os.Getenv("GITHUB_TAGNAME")
	newVersion := strings.TrimPrefix(githubTagname, "v")

	if distDir == "" {
		distDir = "dist"
	}

	access := "public"

	if isPrivate == "" {
		access = "private"
	}

	if newVersion == "" {
		log.Fatal("No new package version set. Set PACKAGE_VERSION env variable.")
	}

	isDistDirEmpty, errOnCheckDistDir := IsDirectoryEmpty(distDir)

	if isDistDirEmpty && errOnCheckDistDir != nil {
		log.Fatal(errOnCheckDistDir)
	}

	if isDistDirEmpty && errOnCheckDistDir == nil {
		log.Fatal("No build files to publish")
	}


	if core.FileExists(".npmrc") == false || isNpmrcValidForPublish() == false {
		WriteFile(".npmrc", fmt.Sprintf("@coopnorge:registry=https://npm.pkg.github.com//npm.pkg.github.com/:_authToken=GITHUB_TOKEN"))
	}

	if core.FileExists("package.json") && isDistDirEmpty == false && errOnCheckDistDir == nil  {
		return sh.RunV(
			"docker", "run", "--rm",
			"-e", fmt.Sprintf("GITHUB_TOKEN=%s", githubToken),
			"-v", "./:/app",
			"node:slim",
			"sh",
			"-c",
			fmt.Sprintf("cd /app && npm version %s && npm publish --access %s", newVersion, access),
		)
	}

	return nil
}

func WriteFile(path string, content string) error {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	defer file.Close()

	if _, err := io.WriteString(file, content); err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	return nil
}

// checks if the .npmrc file is configured for GitHub
// Packages.
func isNpmrcValidForPublish() bool {
	registryURL := "npm.pkg.github.com"
	scope := "@coopnorge"
	tokenIndicator := "_authToken="

	npmrcContent, err := os.ReadFile(".npmrc")

	if err != nil {
		return false
	}

	contentStr := string(npmrcContent)

	if !strings.Contains(contentStr, registryURL) && !strings.Contains(contentStr, scope) && !strings.Contains(contentStr, tokenIndicator) {
		return false
	}

	return true
}

func IsDirectoryEmpty(dirPath string) (bool, error) {
	entries, err := os.ReadDir(dirPath)

	if err != nil {
		return true, err
	}

	return len(entries) == 0, nil
}
