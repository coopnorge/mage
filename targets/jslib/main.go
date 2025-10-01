package jslib

import (
	"context"

	"github.com/coopnorge/mage/internal/javascript"
	"github.com/magefile/mage/mg"
)

// JSLib is the magefile namespace to group Javascript language specific commands
type JSLib mg.Namespace

// Lint checks all javascript/typescript codd for code standards and formats
//
// See [javascript.Lint] for details.
func (JSLib) Lint(ctx context.Context) error {
	mg.CtxDeps(ctx, javascript.Lint)
	return nil
}

// Publish publish npm package to the github package
//
// See [javascript.Publish] for details.
func (JSLib) Publish(ctx context.Context) error {
	mg.CtxDeps(ctx, javascript.PublishLib)
	return nil
}
