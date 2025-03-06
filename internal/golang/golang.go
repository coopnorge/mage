package golang

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/coopnorge/mage/internal/core"
	"github.com/coopnorge/mage/internal/devtool"
)

const coverageReport = "coverage.out"

// IsGoModule checks if a directory contains a go module
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
		if isDotDirectory(workDir, d) {
			return nil
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

func isDotDirectory(path string, d fs.DirEntry) bool {
	if !d.IsDir() {
		return false
	}
	if path == "." {
		return false
	}
	return strings.HasPrefix(path, ".")
}

// Generate files
func Generate(directory string) error {
	return devtool.Run("golang", "go", "-C", directory, "generate", "./...")
}

// Test runs go test
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
	return devtool.Run(
		"golang",
		"go",
		"-C", directory,
		"test",
		"--cover",
		fmt.Sprintf("-coverprofile=%s", output),
		"-covermode=atomic",
		"-race",
		"-tags='datadog.no_waf'",
		"./...")
}

// Lint runs linting
func Lint(directory, golangCILintCfg string) error {
	lintCfg, err := core.WriteTempFile(core.OutputDir, "golangci-lint.yml", golangCILintCfg)
	if err != nil {
		return err
	}
	defer os.Remove(lintCfg.Name()) //nolint:errcheck

	lintCfgPath, err := filepath.Rel(fmt.Sprintf("./%s", directory), lintCfg.Name())
	if err != nil {
		return err
	}

	return devtool.Run("golangci-lint", "bash", "-c", fmt.Sprintf("cd %s && golangci-lint run --verbose --timeout 5m --config %s ./...", directory, lintCfgPath))
}

// LintFix runs auto fixes
func LintFix(directory, golangCILintCfg string) error {
	lintCfg, err := core.WriteTempFile(core.OutputDir, "golangci-lint.yml", golangCILintCfg)
	if err != nil {
		return err
	}
	defer os.Remove(lintCfg.Name()) //nolint:errcheck

	lintCfgPath, err := filepath.Rel(fmt.Sprintf("./%s", directory), lintCfg.Name())
	if err != nil {
		return err
	}
	return devtool.Run("golangci-lint", "bash", "-c", fmt.Sprintf("cd %s && golangci-lint run --verbose --timeout 5m --fix --config %s ./...", directory, lintCfgPath))
}
