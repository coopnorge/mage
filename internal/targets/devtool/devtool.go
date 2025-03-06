package devtool

import (
	"context"

	"github.com/coopnorge/mage/internal/devtool"
)

// Build allow a mage target to depend on a Docker image. This will
// pull the image from a Docker registry.
func Build(_ context.Context, target, dockerfile string) error {
	return devtool.Build(target, dockerfile)
}
