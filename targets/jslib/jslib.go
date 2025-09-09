package jslib

import (
	"context"
	"github.com/coopnorge/mage/internal/targets/javascript"
	"github.com/magefile/mage/mg"
)

// JS is the magefile namespace to group Javascript language specific commands
type JS mg.Namespace

// Lint checks all javascript/typescript codd for code standards and formats
//
// See [javascript.Lint] for details.
func (JS) Lint(ctx context.Context) error {
	mg.CtxDeps(ctx, javascript.Lint)
	return nil
}


// Publish npm package to the github package
//
// See [javascript.PublishLib] for details.
func (JS) PublishLib(ctx context.Context) error {
	mg.CtxDeps(ctx, javascript.PublishLib)
	return nil
}
