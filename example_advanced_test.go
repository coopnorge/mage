package mage_test

import (
	"context"

	//mage:import
	"github.com/coopnorge/mage/targets/goapp"
	"github.com/magefile/mage/mg"
)

// When declaring advanced use cases Build is the only required target.
func Example_advanced() {
	// Do not declare the main function in magefiles
}

// Generate files
func Generate(ctx context.Context) error {
	mg.CtxDeps(ctx, goapp.Generate)
	return nil
}

// Build creates deployable artifacts from the Go source code in the
// repository.
func Build(ctx context.Context) error {
	mg.SerialCtxDeps(ctx, goapp.Build)
	return nil
}

// Validate all code
func Validate(ctx context.Context) error {
	mg.CtxDeps(ctx, goapp.Validate)
	return nil
}

// Fix files
func Fix(ctx context.Context) error {
	mg.CtxDeps(ctx, goapp.Fix)
	return nil
}

// Clean validate and build output
func Clean(ctx context.Context) error {
	mg.CtxDeps(ctx, goapp.Clean)
	return nil
}
