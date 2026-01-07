package golib

import (
	"context"

	palletsTargets "github.com/coopnorge/mage/internal/targets/pallets"
	"github.com/magefile/mage/mg"
)

// Pallets is the magefile namespace to group Pallets commands
type Pallets mg.Namespace

// Validate validates all terraform projects
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
