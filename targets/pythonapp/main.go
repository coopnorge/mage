// Package python implements the [mage targets] for working with Python
// for now it only deals with terraform validation
package pythonapp

import (
	"context"

	"github.com/magefile/mage/mg"
)

// Build runs all ci steps for a python application. For now it only
// validates terraform code within a python application.
func Build(ctx context.Context) error {
	mg.CtxDeps(ctx, Validate)
	return nil
}

// Validate runs validation check on the source code in the repository.
func Validate(ctx context.Context) error {
	mg.CtxDeps(ctx, Terraform.Validate)
	return nil
}

// Fix fixes found issues (if it's supported by the linters)
func Fix(ctx context.Context) error {
	mg.CtxDeps(ctx, Terraform.Fix)
	return nil
}
