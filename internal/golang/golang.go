package golang

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/coopnorge/mage/internal/core"
	"github.com/coopnorge/mage/internal/devtool"
	"github.com/coopnorge/mage/internal/git"
)

const coverageReport = "coverage.out"

var (
	toolGo   devtool.Go
	toolLint devtool.GoLangCILint
)

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

// ContainsGoSourceCode returns true if a directory contains a .go file.
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

// FindGoSourceCodeFolders will return a list of directories that contains
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

// HasChanges checks if the current branch has any Go changes compared
// to the main branch
func HasChanges(goSourceCodeFolders []string, pattern string) (bool, error) {
	changedFiles, err := git.DiffToTagPattern(pattern)
	if err != nil {
		return false, err
	}
	// always trigger on go.mod/sum and workflows because of changes in ci.
	additionalGlobs := append([]string{"**/go.mod", "**/go.sum", ".github/workflows/*"}, strings.Split(os.Getenv("ADDITIONAL_GLOBS_GO"), ",")...)
	return core.CompareChangesToPaths(changedFiles, goSourceCodeFolders, additionalGlobs)
}

// Generate runs commands described by directives within existing files with
// the intent to generate Go code. Those commands can run any process but the
// intent is to create or update Go source files
func Generate(directory string) error {
	return toolGo.Run(nil, "-C", directory, "generate", "./...")
}

// Test automates testing the packages named by the import paths, see also: go
// test.
func Test(directory string) error {
	err := os.MkdirAll(path.Join(core.OutputDir, directory), 0o700)
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

	return toolGo.Run(
		nil,
		"-C",
		directory,
		"test",
		"-vet=off",
		"--cover",
		fmt.Sprintf("-coverprofile=%s", output),
		"-covermode=atomic",
		"-race",
		"-tags='datadog.no_waf'",
		"./...")
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
	return toolLint.Run(nil, directory, "run", "--verbose", "--timeout", "10m", "--config", lintCfgPath, "./...")
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
	return toolLint.Run(nil, directory, "run", "--verbose", "--timeout", "10m", "--fix", "--config", lintCfgPath, "./...")
}

// DownloadModules downloads Go modules locally
func DownloadModules(directory string) error {
	log.Printf("Downloading modules for dir %q", directory)
	return toolGo.Run(nil, "-C", directory, "mod", "download", "-x")
}

// OSArch returns a list of os/arch combinations for which to build binaries
// for
func OSArch() []map[string]string {
	if runtime.GOOS == "darwin" {
		return []map[string]string{
			{
				"GOOS": "darwin", "GOARCH": "arm64",
			},
			{
				"GOOS": "linux", "GOARCH": "amd64",
			},
			{
				"GOOS": "linux", "GOARCH": "arm64",
			},
		}
	}
	if runtime.GOOS == "linux" {
		return []map[string]string{
			{"GOOS": runtime.GOOS, "GOARCH": runtime.GOARCH},
		}
	}
	if runtime.GOOS == "windows" {
		return []map[string]string{
			{"GOOS": runtime.GOOS, "GOARCH": runtime.GOARCH},
		}
	}
	return []map[string]string{
		{
			"GOOS": "darwin", "GOARCH": "arm64",
		},
		{
			"GOOS": "linux", "GOARCH": "amd64",
		},
		{
			"GOOS": "linux", "GOARCH": "arm64",
		},
	}
}

// DockerPlatforms returns the docker platforms to build based on the Golang
// archs that have been build
func DockerPlatforms() string {
	platforms := []string{}
	for _, osarch := range OSArch() {
		if osarch["GOOS"] == "darwin" {
			continue
		}
		platforms = append(platforms, fmt.Sprintf("%s/%s", osarch["GOOS"], osarch["GOARCH"]))
	}
	return strings.Join(platforms, ",")
}
