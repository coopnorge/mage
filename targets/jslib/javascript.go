package jslib

import (
	"context"

	"github.com/coopnorge/mage/internal/targets/javascript"
	"github.com/magefile/mage/mg"
)

// JavaScript is the magefile namespace to group javascript/typescript commands
type JavaScript mg.Namespace

// Install fetches all Node.js dependencies.
func (JavaScript) Install(ctx context.Context) error {
	mg.CtxDeps(ctx, javascript.Install)
	return nil
}

// Lint runs the standard linting script defined in package.json.
func (JavaScript) Lint(ctx context.Context) error {
	mg.CtxDeps(ctx, javascript.Lint)
	return nil
}

// Format runs the standard formatting check script defined in package.json.
func (JavaScript) Format(ctx context.Context) error {
	mg.CtxDeps(ctx, javascript.Format)
	return nil
}

// UnitTest runs unit tests using the package.json script.
func (JavaScript) UnitTest(ctx context.Context) error {
	mg.CtxDeps(ctx, javascript.UnitTest)
	return nil
}

// Build compiles the JavaScript/TypeScript into distribution files.
func (JavaScript) Build(ctx context.Context) error {
	mg.CtxDeps(ctx, mg.F(javascript.Build, "build:library"))
	return nil
}

// NpmPublish publishes npm repository into a github packages
func (JavaScript) NpmPublish(ctx context.Context) error {
	mg.CtxDeps(ctx, mg.F(javascript.NpmPublish, "build:library"))
	return nil
}
