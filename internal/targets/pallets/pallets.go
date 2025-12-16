package pallets

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/coopnorge/mage/internal/devtool"
	"github.com/coopnorge/mage/internal/pallets"
	"github.com/magefile/mage/mg"
)

// KubeConformDocker the content of tools.Dockerfile
//
//go:embed tools.Dockerfile
var KubeConformDocker string

// Validate validates policybot config file
func Validate(ctx context.Context) error {
	mg.CtxDeps(ctx, buildKubeConform)
	err := pallets.Validate()
	if err != nil {
		return err
	}
	return nil
}

// Changes implements a target that check if the current branch has changes
// related to main branch
func Changes(_ context.Context) error {
	changes, err := pallets.HasChanges()
	if err != nil {
		return err
	}

	if changes {
		fmt.Println("true")
		return nil
	}
	fmt.Println("false")
	return nil
}

func buildKubeConform(_ context.Context) error {
	return devtool.Build("kubeconform", KubeConformDocker)
}
