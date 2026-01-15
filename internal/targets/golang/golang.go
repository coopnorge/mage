// Package golang contains targets related to golang
package golang

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/coopnorge/mage/internal/golang"
	"github.com/coopnorge/mage/internal/utils"
	"github.com/magefile/mage/mg"
)

//go:embed golangci-lint.yml
var golangCILintCfg string

// golangciLintFile is the name of the configuration
const golangciLintFile = ".golangci-lint.yaml"

// Generate runs commands described by directives within existing files with
// the intent to generate Go code. Those commands can run any process but the
// intent is to create or update Go source files
func Generate(ctx context.Context) error {
	directories, err := golang.FindGoModules(".")
	if err != nil {
		return err
	}

	generateDirs := []any{}
	for _, workDir := range directories {
		generateDirs = append(generateDirs, mg.F(generate, workDir))
	}

	mg.SerialCtxDeps(ctx, generateDirs...)
	return nil
}

func generate(_ context.Context, workingDirectory string) error {
	return golang.Generate(workingDirectory)
}

// Test automates testing the packages named by the import paths, see also: go
// test.
func Test(ctx context.Context) error {
	directories, err := golang.FindGoModules(".")
	if err != nil {
		return err
	}

	testDirs := []any{}
	for _, workDir := range directories {
		testDirs = append(testDirs, mg.F(test, workDir))
	}

	mg.SerialCtxDeps(ctx, testDirs...)
	return nil
}

func test(_ context.Context, workingDirectory string) error {
	return golang.Test(workingDirectory)
}

// Lint runs the linters
func Lint(ctx context.Context) error {
	directories, err := golang.FindGoModules(".")
	if err != nil {
		return err
	}

	lintDirs := []any{}
	for _, workDir := range directories {
		lintDirs = append(lintDirs, mg.F(lint, workDir))
	}

	mg.SerialCtxDeps(ctx, lintDirs...)
	return nil
}

func lint(_ context.Context, workingDirectory string) error {
	return golang.Lint(workingDirectory, golangCILintCfg)
}

// LintFix fixes found issues (if it's supported by the linters)
func LintFix(ctx context.Context) error {
	directories, err := golang.FindGoModules(".")
	if err != nil {
		return err
	}

	lintDirs := []any{}
	for _, workDir := range directories {
		lintDirs = append(lintDirs, mg.F(lintFix, workDir))
	}

	mg.SerialCtxDeps(ctx, lintDirs...)

	return nil
}

func lintFix(_ context.Context, workingDirectory string) error {
	return golang.LintFix(workingDirectory, golangCILintCfg)
}

// DownloadModules downloads Go modules locally
func DownloadModules(ctx context.Context) error {
	directories, err := golang.FindGoModules(".")
	if err != nil {
		return err
	}
	modules := []any{}
	for _, workDir := range directories {
		modules = append(modules, mg.F(downloadModules, workDir))
	}

	mg.SerialCtxDeps(ctx, modules...)
	return nil
}

func downloadModules(_ context.Context, directory string) error {
	return golang.DownloadModules(directory)
}

// Changes implements a target that check if the current branch has changes
// related to main branch
func Changes(_ context.Context) error {
	directories, err := golang.FindGoSourceCodeFolders(".")
	if err != nil {
		return err
	}

	changes, err := golang.HasChanges(directories)
	if err != nil {
		return err
	}

	if changes {
		fmt.Println("true")
		return nil
	}
	fmt.Println("false")
	return nil
}

// FetchGolangCIConfig fetches and writes the golangci-lint configuration file
// to the repository root if it doesn't already exist.
// TODO(alf): parametrize so that the user can
func FetchGolangCIConfig(where string) error {
	// Get the repository root directory
	repoRoot, err := utils.GetRepoRoot()
	if err != nil {
		return fmt.Errorf("failed to get repository root: %w", err)
	}

	dirs := path.Join(repoRoot, where)
	filePath := path.Join(dirs, golangciLintFile)
	if utils.FileExists(filePath) {
		log.Printf("Config file already exists at %s", filePath)
		return nil
	}

	log.Printf("Writing golangci-lint config to %s", filePath)
	err = os.MkdirAll(dirs, 0777)
	if err != nil {
		fmt.Printf("unable to ensure all directories %s\n", dirs)
		return err
	}

	return os.WriteFile(filePath, []byte(golangCILintCfg), 0644)
}
