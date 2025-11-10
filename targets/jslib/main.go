package jslib

import (
	"context"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// BuildAndPublish checks for linting and formatting issue runs for unit test
// if not skipped and builds the project and publish it to the npm repository
func BuildAndPublish(ctx context.Context) error {
	mg.SerialCtxDeps(ctx, Lint, Format, UnitTest, JavaScript.Build, JavaScript.Publish)
	return nil
}


// Install fetches all Node.js dependencies.
func Install(ctx context.Context) error {
	mg.CtxDeps(ctx, JavaScript.Install)
	return nil
}

// Lint runs the standard linting script defined in package.json.
func Lint(ctx context.Context) error {
	mg.CtxDeps(ctx, JavaScript.Lint)
	return nil
}

// Format runs the standard formatting check script defined in package.json.
func Format(ctx context.Context) error {
	mg.CtxDeps(ctx, JavaScript.Format)
	return nil
}

// UnitTest unit tests using the package.json script.
func UnitTest(ctx context.Context) error {

	mg.CtxDeps(ctx, JavaScript.UnitTest)
	return nil
}
