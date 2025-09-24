package golang

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	doublestar "github.com/bmatcuk/doublestar/v4"

	"github.com/magefile/mage/sh"

	"github.com/coopnorge/mage/internal/core"
	"github.com/coopnorge/mage/internal/devtool"
	"github.com/coopnorge/mage/internal/git"
)

const coverageReport = "coverage.out"

// IsGoModule returns true if a directory contains a go module.
func IsGoModule(p string, d fs.DirEntry) bool {
	if !d.IsDir() {
		return false
	}
	if _, err := os.Stat(path.Join(p, "go.mod")); os.IsNotExist(err) {
		return false
	}
	return true
}

// FindGoModules will search through the base directory to find the all the
// Go modules.
func FindGoModules(base string) ([]string, error) {
	directories := []string{}

	err := filepath.WalkDir(base, func(workDir string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if core.IsDotDirectory(workDir, d) {
			return filepath.SkipDir
		}
		if !IsGoModule(workDir, d) {
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

// IsGoModule returns true if a directory contains a go module. It returns
// true is a directory contains a .go file.
func ContainsGoSourceCode(p string, d fs.DirEntry) (bool, error) {
	if !d.IsDir() {
		return false, nil
	}

	entries, err := os.ReadDir(p)
	if err != nil {
		return false, err
	}

	for _, entry := range entries {
		ext := filepath.Ext(entry.Name())
		if strings.EqualFold(".go", ext) {
			return true, nil
		}
	}
	return false, nil
}

// FindGoSourceCodeFolders will return a list of directories at contain
// golang source code
func FindGoSourceCodeFolders(base string) ([]string, error) {
	directories := []string{}

	err := filepath.WalkDir(base, func(workDir string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if core.IsDotDirectory(workDir, d) {
			return filepath.SkipDir
		}

		sourceCodeDir, err := ContainsGoSourceCode(workDir, d)
		if err != nil {

			return err
		}
		if !sourceCodeDir {
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

// HasChanges checks if the current branch has any terraform changes compared
// to the main branch
func HasChanges(goSourceCodeFolders []string) (bool, error) {
	changedFiles, err := git.DiffToMain()
	if err != nil {
		return false, err
	}
	for _, change := range changedFiles {
		for _, folder := range goSourceCodeFolders {
			match, err := path.Match(fmt.Sprintf("%s/*", folder), change)
			if err != nil {
				return false, err
			}
			if match {
				return true, nil
			}
		}
		for _, pattern := range strings.Split(os.Getenv("ADDITIONAL_GLOBS_GO"), ",") {
			matchAdditional, err := doublestar.Match(pattern, change)
			if err != nil {
				return false, err
			}
			if matchAdditional {
				return true, nil
			}
		}

	}
	return false, nil
}

// Generate runs commands described by directives within existing files with
// the intent to generate Go code. Those commands can run any process but the
// intent is to create or update Go source files
func Generate(directory string) error {
	return DevtoolGo(nil, "go", "-C", directory, "generate", "./...")
}

// Test automates testing the packages named by the import paths, see also: go
// test.
func Test(directory string) error {
	err := os.MkdirAll(path.Join(core.OutputDir, directory), 0700)
	if err != nil {
		return err
	}

	rootPath, err := os.Getwd()
	if err != nil {
		return err
	}
	relativeRootPath, err := core.GetRelativeRootPath(rootPath, directory)
	if err != nil {
		return err
	}

	output := path.Join(relativeRootPath, core.OutputDir, directory, coverageReport)

	return DevtoolGo(nil, "go", "-C", directory, "test", "--cover", fmt.Sprintf("-coverprofile=%s", output), "-covermode=atomic", "-race", "-tags='datadog.no_waf'", "./...")
}

// Lint runs the linters
func Lint(directory, golangCILintCfg string) error {
	lintCfg, cleanup, err := core.WriteTempFile(core.OutputDir, "golangci-lint.yml", golangCILintCfg)
	if err != nil {
		return err
	}
	defer cleanup()

	lintCfgPath, err := filepath.Rel(fmt.Sprintf("./%s", directory), lintCfg)
	if err != nil {
		return err
	}

	return DevtoolGolangCILint(nil, "bash", "-c", fmt.Sprintf("cd %s && golangci-lint run --verbose --timeout 5m --config %s ./...", directory, lintCfgPath))
}

// LintFix fixes found issues (if it's supported by the linters)
func LintFix(directory, golangCILintCfg string) error {
	lintCfg, cleanup, err := core.WriteTempFile(core.OutputDir, "golangci-lint.yml", golangCILintCfg)
	if err != nil {
		return err
	}
	defer cleanup()

	lintCfgPath, err := filepath.Rel(fmt.Sprintf("./%s", directory), lintCfg)
	if err != nil {
		return err
	}
	return DevtoolGolangCILint(nil, "bash", "-c", fmt.Sprintf("cd %s && golangci-lint run --verbose --timeout 5m --fix --config %s ./...", directory, lintCfgPath))
}

// DownloadModules downloads Go modules locally
func DownloadModules(directory string) error {
	log.Printf("Downloading modules for dir %q", directory)
	return DevtoolGo(nil, "go", "-C", directory, "mod", "download", "-x")
}

// DevtoolGo runs the devtool for Go
func DevtoolGo(env map[string]string, cmd string, args ...string) error {
	// This is a bit hacky to use the local go binary instead of the container
	// this is used for running the integration tests on targets.
	if os.Getenv("GO_RUNTIME") == "local" {
		return sh.RunWithV(env, cmd, args...)
	}

	path, err := os.Getwd()
	if err != nil {
		return err
	}

	goModCache, err := sh.Output("go", "env", "GOMODCACHE")
	if err != nil {
		goModCache = "$HOME/go/pkg/mod"
	}

	dockerArgs := []string{
		"--volume", fmt.Sprintf("%s:/go/pkg/mod", goModCache),
		"--volume", "/var/run/docker.sock:/var/run/docker.sock",
		"--volume", "$HOME/.cache:/root/.cache",
		"--volume", "$HOME/.gitconfig:/root/.gitconfig",
		"--volume", "$HOME/.ssh:/root/.ssh",
		"--env", "TESTCONTAINERS_HOST_OVERRIDE=host.docker.internal",
		"--add-host", "host.docker.internal:host-gateway",
		"--volume", fmt.Sprintf("%s:/app", path),
		"--workdir", "/app",
	}

	if env == nil {
		env = map[string]string{}
	}
	for k, v := range env {
		dockerArgs = append(dockerArgs, "--env", fmt.Sprintf("%s=%s", k, v))
	}

	return devtool.Run("golang", dockerArgs, cmd, args...)
}

// DevtoolGolangCILint runs the devtool for Golangci-lint
func DevtoolGolangCILint(env map[string]string, cmd string, args ...string) error {
	path, err := os.Getwd()
	if err != nil {
		return err
	}

	goModCache, err := sh.Output("go", "env", "GOMODCACHE")
	if err != nil {
		goModCache = "$HOME/go/pkg/mod"
	}

	dockerArgs := []string{
		"--volume", fmt.Sprintf("%s:/go/pkg/mod", goModCache),
		"--volume", fmt.Sprintf("%s:/app", path),
		"--workdir", "/app",
	}

	if env == nil {
		env = map[string]string{}
	}
	for k, v := range env {
		dockerArgs = append(dockerArgs, "--env", fmt.Sprintf("%s=%s", k, v))
	}

	return devtool.Run("golangci-lint", dockerArgs, cmd, args...)
}
