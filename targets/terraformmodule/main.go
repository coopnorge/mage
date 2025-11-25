// Package goapp implements the [mage targets] for working with Go
// applications.
//
// # Setup
//
// In the root of the repository initialize a new Go modules and import this
// module and mage as as tool.
//
//	go mod init
//	go get github.com/coopnorge/mage@latest
//	go get -tool github.com/magefile/mage
//
// Configure mage in magefiles/magefile.go (goapp is used for the example).
//
//	import (
//		//mage:import
//		_ "github.com/coopnorge/mage/targets/terraformmodule"
//	)
//
// In the repository add a GitHub Actions Workflow
//
//	on:
//	  pull_request: {}
//	  push:
//	    branches:
//	      - main
//	jobs:
//	  cicd:
//	    uses: coopnorge/mage/.github/workflows/mage.yaml@v0 # Use the latest available version
//	    permissions:
//	      contents: read
//	      id-token: write
//	      packages: read
//	    secrets: inherit
//
// [mage targets]: https://magefile.org/targets/
//
// [import]: https://magefile.org/importing/
package terraformmodule

import (
	"context"

	"github.com/magefile/mage/mg"
)

// Build runs all validation steps
func Build(ctx context.Context) error {
	mg.SerialCtxDeps(ctx, Validate)
	return nil
}

// Validate runs validation check on the source code in the repository.
//
// For details see [Terraform.Validate]
func Validate(ctx context.Context) error {
	mg.CtxDeps(ctx, Terraform.Validate, PolicyBotConfig.Validate, CatalogInfo.Validate)
	return nil
}

// Fix fixes found issues (if it's supported by the linters)
//
// For details see and [Terraform.Fix].
func Fix(ctx context.Context) error {
	mg.CtxDeps(ctx, Terraform.Fix)
	return nil
}
