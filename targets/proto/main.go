package proto

import (
	"context"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// Build runs Validate
func Build(ctx context.Context) error {
	mg.SerialCtxDeps(ctx, Validate, Proto.Generate)
	return nil
}

// Generate output
func Generate(ctx context.Context) error {
	mg.SerialCtxDeps(ctx, Validate, Proto.Generate)
	return nil
}

// Validate all code
func Validate(ctx context.Context) error {
	mg.CtxDeps(ctx, Proto.Validate)
	return nil
}

// Fix files
func Fix(ctx context.Context) error {
	mg.CtxDeps(ctx, Proto.Fix)
	return nil
}

// Clean validate and build output
func Clean(_ context.Context) error {
	return sh.Rm(outputDir)
}
