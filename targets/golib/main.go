// Package golib implements the [mage targets] for working with Go libraries.
//
// To enable the targets in a repository [import] them in
// magefiles/magefile.go.
//
// [mage targets]: https://magefile.org/targets/
// [import]: https://magefile.org/importing/
package golib

import (
	"context"

	"github.com/coopnorge/mage/internal/core"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// Generate runs commands described by directives within existing files with
// the intent to generate Go code. Those commands can run any process but the
// intent is to create or update Go source files
//
// For details see [Go.Generate].
func Generate(ctx context.Context) error {
	mg.CtxDeps(ctx, Go.Generate)
	return nil
}

// Build runs validate
//
// For details see [Validate].
func Build(ctx context.Context) error {
	mg.SerialCtxDeps(ctx, Validate)
	return nil
}

// Validate runs validation check on the source code in the repository.
//
// For details see [Go.Validate].
func Validate(ctx context.Context) error {
	mg.CtxDeps(ctx, Go.Validate)
	return nil
}

// Fix fixes found issues (if it's supported by the linters)
//
// For details see [Go.Fix].
func Fix(ctx context.Context) error {
	mg.CtxDeps(ctx, Go.Fix)
	return nil
}

// Clean removes validate and build output.
//
// Deletes the [core.OutputDir].
func Clean(_ context.Context) error {
	return sh.Rm(core.OutputDir)
}
