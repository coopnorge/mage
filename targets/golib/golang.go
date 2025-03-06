package golib

import (
	"context"

	"github.com/coopnorge/mage/internal/targets/golang"

	"github.com/magefile/mage/mg"
)

// Go is the magefile namespace to group Go commands
type Go mg.Namespace

// Generate files
func (Go) Generate(ctx context.Context) error {
	mg.CtxDeps(ctx, golang.Generate)
	return nil
}

// Validate files
func (Go) Validate(ctx context.Context) error {
	mg.CtxDeps(ctx, Go.Test, Go.Lint)
	return nil
}

// Fix files
func (Go) Fix(ctx context.Context) error {
	mg.CtxDeps(ctx, Go.LintFix)
	return nil
}

// Test runs the test suite
func (Go) Test(ctx context.Context) error {
	mg.CtxDeps(ctx, golang.Test)
	return nil
}

// Lint all source code
func (Go) Lint(ctx context.Context) error {
	mg.CtxDeps(ctx, golang.Lint)
	return nil
}

// LintFix linting issues
func (Go) LintFix(ctx context.Context) error {
	mg.CtxDeps(ctx, golang.LintFix)
	return nil
}
