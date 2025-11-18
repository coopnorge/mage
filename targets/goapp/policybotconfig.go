package goapp

import (
	"context"

	policyBotTargets "github.com/coopnorge/mage/internal/targets/policybot"
	"github.com/magefile/mage/mg"
)

// PolicyBotConfig is the magefile namespace to group PolicyBotConfig commands
type PolicyBotConfig mg.Namespace

// Validate validates all terraform projects
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
