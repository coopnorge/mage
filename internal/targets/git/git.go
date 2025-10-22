package git

import (
	"context"
	"fmt"

	"github.com/coopnorge/mage/internal/git"
)

// ListChanges list all changes compared to origin/main
func ListChanges(_ context.Context) error {
	changes, err := git.DiffToMain()
	if err != nil {
		return err
	}
	for _, change := range changes {
		fmt.Println(change)
	}
	return nil
}
