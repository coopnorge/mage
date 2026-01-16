package golib

import (
	"context"

	"github.com/coopnorge/mage/internal/targets/golang"

	"github.com/magefile/mage/mg"
)

// Go is the magefile namespace to group Go commands
type Go mg.Namespace

// Generate runs commands described by directives within existing files with
// the intent to generate Go code. Those commands can run any process but the
// intent is to create or update Go source files
//
// For details see [golang.Generate].
func (Go) Generate(ctx context.Context) error {
	mg.CtxDeps(ctx, golang.Generate)
	return nil
}

// Validate runs validation check on the Go source code in the repository.
//
// See [Go.Test] and [Go.Lint] for details.
func (Go) Validate(ctx context.Context) error {
	mg.CtxDeps(ctx, Go.DownloadModules)
	mg.CtxDeps(ctx, Go.Test, Go.Lint, CatalogInfo.Validate)
	return nil
}

// Fix runs auto fixes on the Go source code in the repository.
//
// For details see [Go.LintFix].
func (Go) Fix(ctx context.Context) error {
	mg.CtxDeps(ctx, Go.LintFix)
	return nil
}

// Test automates testing the packages named by the import paths, see also: go
// test.
//
// For details see [golang.Test].
func (Go) Test(ctx context.Context) error {
	mg.CtxDeps(ctx, golang.Test)
	return nil
}

// Lint checks all Go source code for issues.
//
// See [golang.Lint] for details.
func (Go) Lint(ctx context.Context) error {
	mg.CtxDeps(ctx, golang.Lint)
	return nil
}

// LintFix fixes found issues (if it's supported by the linters)
//
// For details see [golang.LintFix].
func (Go) LintFix(ctx context.Context) error {
	mg.CtxDeps(ctx, golang.LintFix)
	return nil
}

// Changes returns the string true or false depending on the fact that
// the current branch contains changes compared to the main branch.
func (Go) Changes(ctx context.Context) error {
	mg.CtxDeps(ctx, golang.Changes)
	return nil
}

// DownloadModules download the go modules
func (Go) DownloadModules(ctx context.Context) error {
	mg.CtxDeps(ctx, golang.DownloadModules)
	return nil
}

// FetchGolangCILintConfig writes the golangci-lint configuration file provided path relative
// to root if it doesn't already exist.
func (Go) FetchGolangCILintConfig(_ context.Context, where string) error {
	// Leaving context unused which will be when logging package exists
	return golang.FetchGolangCIConfig(where)
}
