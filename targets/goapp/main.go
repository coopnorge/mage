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
//		_ "github.com/coopnorge/mage/targets/goapp"
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
//	    with:
//	      oci-image-base: europe-docker.pkg.dev/helloworld-shared-0918
//	      push-oci-image: ${{ github.ref == 'refs/heads/main' }}
//	      workload-identity-provider: projects/889992792607/locations/global/workloadIdentityPools/github-actions/providers/github-actions-provider
//	      service-account: helloworld-github-actions@helloworld-shared-0918.iam.gserviceaccount.com
//
// [mage targets]: https://magefile.org/targets/
//
// [import]: https://magefile.org/importing/
package goapp

import (
	"context"
	"os"
	"strconv"

	"github.com/coopnorge/mage/internal/core"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

const (
	// PushEnv is the name of the environmental variable used to trigger
	// pushing of OCI images. Set PUSH_IMAGE to true to push images.
	PushEnv = "PUSH_IMAGE"
)

// Generate runs commands described by directives within existing files with
// the intent to generate Go code. Those commands can run any process but the
// intent is to create or update Go source files
//
// For details see [Go.Generate].
func Generate(ctx context.Context) error {
	mg.CtxDeps(ctx, Go.Generate)
	return nil
}

// Build creates deployable artifacts from the source code in the repository,
// to push the resulting images set the environmental variable PUSH_IMAGE to
// true. Setting PUSH_IMAGE to true will disable the latest image tag.
//
// For details see [Go.Build] and [Docker.BuildAndPush].
func Build(ctx context.Context) error {
	shouldPush, err := shouldPush()
	if err != nil {
		return err
	}
	mg.SerialCtxDeps(ctx, Validate, Go.Build, mg.F(Docker.BuildAndPush, shouldPush))
	return nil
}

// Validate runs validation check on the source code in the repository.
//
// For details see [Go.Validate], [Terraform.Validate] and [Docker.Validate].
func Validate(ctx context.Context) error {
	mg.CtxDeps(ctx, Go.Validate, Docker.Validate, Terraform.Validate)
	return nil
}

// Fix fixes found issues (if it's supported by the linters)
//
// For details see [Go.Fix] and [Terraform.Fix].
func Fix(ctx context.Context) error {
	mg.CtxDeps(ctx, Go.Fix, Terraform.Fix)
	return nil
}

// Clean removes validate and build output.
//
// Deletes the [core.OutputDir].
func Clean(_ context.Context) error {
	return sh.Rm(core.OutputDir)
}

func shouldPush() (bool, error) {
	val, ok := os.LookupEnv(PushEnv)
	if !ok || val == "" {
		return false, nil
	}
	boolValue, err := strconv.ParseBool(val)
	if err != nil {
		return false, err
	}
	return boolValue, nil
}
