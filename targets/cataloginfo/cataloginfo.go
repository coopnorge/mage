// Package cataloginfo exposes the CatalogInfo targets as a standalone mage namespace.
//
// Import it in magefiles/magefile.go, without declaring a main function:
//
//	//mage:import
//	_ "github.com/coopnorge/mage/targets/cataloginfo"
//
// Importing this package together with a bundle such as targets/goapp or
// targets/golib results in duplicate target errors, because the bundles define
// the same targets. Alias one of the imports to avoid this:
//
//	//mage:import cataloginfo
//	_ "github.com/coopnorge/mage/targets/cataloginfo"
package cataloginfo

import (
	"context"

	catalogInfoTargets "github.com/coopnorge/mage/internal/targets/cataloginfo"
	"github.com/magefile/mage/mg"
)

// CatalogInfo is the magefile namespace to group CatalogInfo commands
type CatalogInfo mg.Namespace

// Validate validates the CatalogInfo configuration
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
