package pallets

import (
	"context"
	"fmt"

	"github.com/coopnorge/mage/internal/pallets"
)

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
