package cataloginfo

import (
	"context"
	"fmt"

	"github.com/coopnorge/mage/internal/cataloginfo"
)

// HasChanges checks if the current branch has any catalog-info changes compared
// to the main branch
func HasChanges() error {
	changes, err := cataloginfo.HasChanges()
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

// Validate validates catalog-info files
func Validate(_ context.Context) error {
	return cataloginfo.Validate()
}
