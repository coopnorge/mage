package goapp

import (
	"context"

	terraformTargets "github.com/coopnorge/mage/internal/targets/terraform"

	"github.com/magefile/mage/mg"
)

// Go is the magefile namespace to group Go commands
type Terraform mg.Namespace



// For details see [Terraform.Test] and [Terraform.Lint].
func (Terraform) Validate(ctx context.Context) error {
	mg.CtxDeps(ctx, Terraform.Test, Terraform.Lint)
	return nil
}

// Fix runs auto fixes on the Go source code in the repository.
//
// For details see [Terraform.LintFix].
func (Terraform) Fix(ctx context.Context) error {
	mg.CtxDeps(ctx, Terraform.LintFix)
	return nil
}

// Test automates testing the packages named by the import paths, see also: go
// test.
//
// For details see [terraformTargets.Test].
func (Terraform) Test(ctx context.Context) error {
	mg.CtxDeps(ctx, mg.F(terraformTargets.Test))
	return nil
}

// Lint checks all Go source code for issues.
//
// For details see [terraformTargets.Lint].
func (Terraform) Lint(ctx context.Context) error {
	mg.CtxDeps(ctx, terraformTargets.Lint)
	return nil
}

// LintFix fixes found issues (if it's supported by the linters)
//
// For details see [terraformTargets.LintFix].
func (Terraform) LintFix(ctx context.Context) error {
	mg.CtxDeps(ctx, terraformTargets.LintFix)
	return nil
}

// For details see [terraformTargets.LintFix].
func (Terraform) Init(ctx context.Context) error {
	mg.CtxDeps(ctx, terraformTargets.Init)
	return nil
}
