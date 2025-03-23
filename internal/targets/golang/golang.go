package golang

import (
	"context"
	_ "embed"

	"github.com/coopnorge/mage/internal/golang"
	"github.com/coopnorge/mage/internal/targets/devtool"
	"github.com/magefile/mage/mg"
)

var (
	//go:embed golangci-lint.yml
	golangCILintCfg string
	//go:embed tools.Dockerfile
	// GolangToolsDockerfile the content of tools.Dockerfile
	GolangToolsDockerfile string
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

func generate(ctx context.Context, workingDirectory string) error {
	mg.CtxDeps(ctx, mg.F(devtool.Build, "golang", GolangToolsDockerfile))
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

func test(ctx context.Context, workingDirectory string) error {
	mg.CtxDeps(ctx, mg.F(devtool.Build, "golang", GolangToolsDockerfile))
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

func lint(ctx context.Context, workingDirectory string) error {
	mg.CtxDeps(ctx, mg.F(devtool.Build, "golangci-lint", GolangToolsDockerfile))
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

func lintFix(ctx context.Context, workingDirectory string) error {
	mg.CtxDeps(ctx, mg.F(devtool.Build, "golangci-lint", GolangToolsDockerfile))
	return golang.LintFix(workingDirectory, golangCILintCfg)
}
