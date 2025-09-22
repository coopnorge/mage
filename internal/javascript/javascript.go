package javascript

import (
	"fmt"
	"log"
	"os"
	"strings"
	"io/fs"
	"path"
	"path/filepath"

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
		log.Fatal("Biome not setup in your project. Install @coopnorge/web-devtools.")
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

	if isPrivate == "" {
		access = "private"
	}

	if newVersion == "" {
		log.Fatal("No new package version set. Set PACKAGE_VERSION env variable.")
	}

	isDistDirEmpty, errOnCheckDistDir := core.IsDirectoryEmpty(distDir)

	if isDistDirEmpty && errOnCheckDistDir != nil {
		log.Fatal(errOnCheckDistDir)
	}

	if isDistDirEmpty && errOnCheckDistDir == nil {
		log.Fatal("No build files to publish")
	}

	if !core.FileExists(".npmrc") {
		log.Fatal(".npmrc file missing.")
		os.Exit(1)
	}

	if !core.IsNpmrcValidForPublish(".") {
		log.Fatal(".npmrc has no auth configuration.")
		os.Exit(1)
	}

	if (shouldBuild == true) {
		if (buildCommand == "") {
			buildCommand = "build"
		}
		buildCommand = fmt.Sprintf("npm ci && npm run %s", buildCommand)
	}

	if core.FileExists("package.json") && !isDistDirEmpty && errOnCheckDistDir == nil {
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

	return nil
}

// func shouldPush() (bool, error) {
// 	val, ok := os.LookupEnv(PushEnv)
// 	if !ok || val == "" {
// 		return false, nil
// 	}
// 	boolValue, err := strconv.ParseBool(val)
// 	if err != nil {
// 		return false, err
// 	}
// 	return boolValue, nil
// }

// IsNodeModule checks if directory is a Node.js project by looking for a
// 'package.json' file.
func IsNodeModule(p string, d os.DirEntry) bool {
	// A Node.js module root must be a directory

	if !d.IsDir() {
		return false
	}

	// Check for the existence of 'package.json' within the directory
	if _, err := os.Stat(path.Join(p, "package.json")); os.IsNotExist(err) {
		return false
	}
	return true
}

// FindNodeModules finds all Node.js projects within a base directory.
// It works similarly to the FindGoModules function by walking the diretory
// tree.
func FindNodeModules(base string) ([]string, error) {
	directories := []string{}

	err := filepath.WalkDir(base, func(workDir string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if core.IsDotDirectory(workDir, d) {
			return filepath.SkipDir
		}
		if !IsNodeModule(workDir, d) {
			return nil
		}

		directories = append(directories, workDir)

		return nil
	})
	if err != nil {
		return nil, err
	}
	return directories, nil
}
