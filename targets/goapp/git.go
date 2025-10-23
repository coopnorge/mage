package goapp

import (
	"context"

	gitTargets "github.com/coopnorge/mage/internal/targets/git"

	"github.com/magefile/mage/mg"
)

// Git is the magefile namespace to group Git commands
type Git mg.Namespace

// ListChanges list all changes to origin/main
func (Git) ListChanges(ctx context.Context) error {
	mg.CtxDeps(ctx, gitTargets.ListChanges)
	return nil
}
