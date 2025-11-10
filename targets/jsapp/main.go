package jsapp

import (
	"context"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// BuildAndPublish creates deployable artifacts from the source code in the repository,
// to push the resulting images set the environmental variable PUSH_IMAGE to
// true. Setting PUSH_IMAGE to true will disable the latest image tag.
func BuildAndPublish(ctx context.Context) error {
	mg.SerialCtxDeps(ctx, Install, Lint, Format, UnitTest, E2ETest, JavaScript.BuildAndPushDockerImage)
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

// E2ETest runs browser tests using the package.json script.
func E2ETest(ctx context.Context) error {
	mg.CtxDeps(ctx, JavaScript.E2ETest)
	return nil
}

