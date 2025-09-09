package javascript

import (
	"context"

	"github.com/magefile/mage/mg"
)

// Lint runs linting on js/ts project
func Lint(ctx context.Context) error {
	mg.CtxDeps(ctx, Lint)
	return nil
}

// PublishLib checks if package.json file exists or not, checks if distribution/build-output folder
// exists or not, checks if .npmrc file exits or not
func PublishLib(ctx context.Context) error {
	mg.CtxDeps(ctx, PublishLib)
	return nil
}
