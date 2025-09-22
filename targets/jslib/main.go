package jslib

import (
	"context"

	"github.com/coopnorge/mage/internal/targets/javascript"
	"github.com/magefile/mage/mg"
)

// JS is the magefile namespace to group Javascript language specific commands
type JSLIB mg.Namespace

// Lint checks all javascript/typescript codd for code standards and formats
//
// See [javascript.Lint] for details.
func (JSLIB) Lint(ctx context.Context) error {
	mg.CtxDeps(ctx, javascript.Lint)
	return nil
}

// PublishLib publish npm package to the github package
//
// See [javascript.PublishLib] for details.
func (JSLIB) PublishLib(ctx context.Context) error {
	mg.CtxDeps(ctx, javascript.PublishLib)
	return nil
}
