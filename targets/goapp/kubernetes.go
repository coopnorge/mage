package goapp

import (
	"context"

	kubernetesTargets "github.com/coopnorge/mage/internal/targets/kubernetes"
	"github.com/magefile/mage/mg"
)

// K8s is the magefile namespace to group Kubernetes commands
type K8s mg.Namespace

// Validate validates all helm charts
func (K8s) Validate(ctx context.Context) error {
	mg.CtxDeps(ctx, mg.F(kubernetesTargets.Validate))
	return nil
}

// Diff returns the string true or false depending on the fact that
// the current branch contains changes compared to the main branch.
func (K8s) Diff(ctx context.Context) error {
	mg.CtxDeps(ctx, kubernetesTargets.Diff)
	return nil
}

func (K8s) List(ctx context.Context) error {
	mg.CtxDeps(ctx, kubernetesTargets.List)
	return nil
}
