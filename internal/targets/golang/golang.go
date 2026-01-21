// Package golang contains targets related to golang
package golang

import (
	"context"
	"fmt"

	"github.com/coopnorge/mage/internal/devtool"
	golangcilint "github.com/coopnorge/mage/internal/devtool/golangci-lint"
	"github.com/coopnorge/mage/internal/golang"
	"github.com/magefile/mage/mg"
)

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
	return golang.Lint(workingDirectory, golangcilint.Cfg())
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
	return golang.LintFix(workingDirectory, golangcilint.Cfg())
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
// to the specified directory relative to the repository root.
// The config file will be named .golangci-lint.yaml.
//
// The where parameter specifies the directory path relative to the repository root.
// Use "." or "" to write to the repository root directory.
func FetchGolangCIConfig(where string) error {
	return devtool.FetchGolangCILintConfig(where)
}
