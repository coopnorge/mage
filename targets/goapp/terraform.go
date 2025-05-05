package goapp

import (
	"context"

	terraformTargets "github.com/coopnorge/mage/internal/targets/terraform"

	"github.com/magefile/mage/mg"
)

// Terraform is the magefile namespace to group Terraform commands
type Terraform mg.Namespace

// Validate validates all terraform projects
func (Terraform) Validate(ctx context.Context) error {
	mg.CtxDeps(ctx, Terraform.Test, Terraform.Lint, Terraform.Security)
	return nil
}

// Fix tries to fix all validation issues where possible
func (Terraform) Fix(ctx context.Context) error {
	mg.CtxDeps(ctx, Terraform.LintFix)
	return nil
}

// Test tests all terraform projects
func (Terraform) Test(ctx context.Context) error {
	mg.CtxDeps(ctx, mg.F(terraformTargets.Test))
	return nil
}

// Lint lints all terraform projects
func (Terraform) Lint(ctx context.Context) error {
	mg.CtxDeps(ctx, terraformTargets.Lint)
	return nil
}

// LintFix tries to fix linting issues
func (Terraform) LintFix(ctx context.Context) error {
	mg.CtxDeps(ctx, terraformTargets.LintFix)
	return nil
}

// Init initializes a terraform projects
func (Terraform) Init(ctx context.Context) error {
	mg.CtxDeps(ctx, terraformTargets.Init)
	return nil
}

// InitUpgrade upgrades the terraform projects within their version
// constraints.
func (Terraform) InitUpgrade(ctx context.Context) error {
	mg.CtxDeps(ctx, terraformTargets.InitUpgrade)
	return nil
}

// LockProviders pdates the locks.terraform.lock.hcl file. Run this when a provider has
// changed.
func (Terraform) LockProviders(ctx context.Context) error {
	mg.CtxDeps(ctx, terraformTargets.LockProviders)
	return nil
}

// Clean the cache directory in the terraform projects
func (Terraform) Clean(ctx context.Context) error {
	mg.CtxDeps(ctx, terraformTargets.Clean)
	return nil
}

// Security scans the security posture of the terraform projects
func (Terraform) Security(ctx context.Context) error {
	mg.CtxDeps(ctx, terraformTargets.Security)
	return nil
}
