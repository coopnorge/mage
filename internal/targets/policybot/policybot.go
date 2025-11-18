package policybot

import (
	"context"
	"fmt"

	"github.com/coopnorge/mage/internal/policybot"
)

// Validate validates policybot config file
func Validate(ctx context.Context) error {
	return policybot.Validate(ctx)
}

// Changes implements a target that check if the current branch has changes
// related to main branch
func Changes(_ context.Context) error {
	changes, err := policybot.HasChanges()
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
