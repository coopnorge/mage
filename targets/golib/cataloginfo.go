package golib

import (
	"context"

	catalogInfoTargets "github.com/coopnorge/mage/internal/targets/cataloginfo"
	"github.com/magefile/mage/mg"
)

// CatalogInfo is the magefile namespace to group CatalogInfo commands
type CatalogInfo mg.Namespace

// Validate validates all terraform projects
func (CatalogInfo) Validate(ctx context.Context) error {
	mg.CtxDeps(ctx, mg.F(catalogInfoTargets.Validate))
	return nil
}

// Changes returns the string true or false depending on the fact that
// the current branch contains changes compared to the main branch.
func (CatalogInfo) Changes(ctx context.Context) error {
	mg.CtxDeps(ctx, catalogInfoTargets.HasChanges)
	return nil
}
