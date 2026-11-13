// Package policybotconfig exposes the PolicyBotConfig targets as a standalone mage namespace.
//
// Import it in magefiles/magefile.go, without declaring a main function:
//
//	//mage:import
//	_ "github.com/coopnorge/mage/targets/policybotconfig"
//
// Importing this package together with a bundle such as targets/goapp or
// targets/golib results in duplicate target errors, because the bundles define
// the same targets. Alias one of the imports to avoid this:
//
//	//mage:import policybotconfig
//	_ "github.com/coopnorge/mage/targets/policybotconfig"
package policybotconfig

import (
	"context"

	policyBotTargets "github.com/coopnorge/mage/internal/targets/policybot"
	"github.com/magefile/mage/mg"
)

// PolicyBotConfig is the magefile namespace to group PolicyBotConfig commands
type PolicyBotConfig mg.Namespace

// Validate validates the PolicyBotConfig configuration
func (PolicyBotConfig) Validate(ctx context.Context) error {
	mg.CtxDeps(ctx, mg.F(policyBotTargets.Validate))
	return nil
}

// Changes returns the string true or false depending on the fact that
// the current branch contains changes compared to the main branch.
func (PolicyBotConfig) Changes(ctx context.Context) error {
	mg.CtxDeps(ctx, policyBotTargets.Changes)
	return nil
}
