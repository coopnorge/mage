package goapp

import (
	"context"

	"github.com/coopnorge/mage/internal/core"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// Generate files
func Generate(ctx context.Context) error {
	mg.CtxDeps(ctx, Go.Generate)
	return nil
}

// Build OCI image
func Build(ctx context.Context) error {
	mg.SerialCtxDeps(ctx, Validate, Go.Build)
	return nil
}

// Validate all code
func Validate(ctx context.Context) error {
	mg.CtxDeps(ctx, Go.Validate)
	return nil
}

// Fix files
func Fix(ctx context.Context) error {
	mg.CtxDeps(ctx, Go.Fix)
	return nil
}

// Clean validate and build output
func Clean(_ context.Context) error {
	return sh.Rm(core.OutputDir)
}
