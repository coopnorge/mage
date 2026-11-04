// Package pallets exposes the Pallets targets as a standalone mage namespace.
//
// Import it in magefiles/magefile.go, without declaring a main function:
//
//	//mage:import
//	_ "github.com/coopnorge/mage/targets/pallets"
//
// Importing this package together with a bundle such as targets/goapp or
// targets/golib results in duplicate target errors, because the bundles define
// the same targets. Alias one of the imports to avoid this:
//
//	//mage:import pallets
//	_ "github.com/coopnorge/mage/targets/pallets"
package pallets

import (
	"context"

	palletsTargets "github.com/coopnorge/mage/internal/targets/pallets"
	"github.com/magefile/mage/mg"
)

// Pallets is the magefile namespace to group Pallets commands
type Pallets mg.Namespace

// Validate validates the Pallets configuration
func (Pallets) Validate(ctx context.Context) error {
	mg.CtxDeps(ctx, mg.F(palletsTargets.Validate))
	return nil
}

// Changes returns the string true or false depending on the fact that
// the current branch contains changes compared to the main branch.
func (Pallets) Changes(ctx context.Context) error {
	mg.CtxDeps(ctx, palletsTargets.Changes)
	return nil
}
