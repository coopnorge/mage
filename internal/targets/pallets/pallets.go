package pallets

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/coopnorge/mage/internal/devtool"
	"github.com/coopnorge/mage/internal/pallets"
)

// KubeConformDocker the content of tools.Dockerfile
//
//go:embed tools.Dockerfile
var KubeConformDocker string

// Validate validates policybot config file
func Validate(_ context.Context) error {
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

// DownloadDevTool downloads devtools related to pallet validation
func DownloadDevTool(_ context.Context, tool string) error {
	return devtool.Build(tool, KubeConformDocker)
}
