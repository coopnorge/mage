package kubernetes

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/coopnorge/mage/internal/core"
	"github.com/coopnorge/mage/internal/kubernetes"
)

// Validate runs kubeconform, kube-score and render templates
func Validate(ctx context.Context) error {
	charts, err := kubernetes.FindHelmCharts(".")
	if err != nil {
		return err
	}
	// we are not using mg.(Serial)CtxDeps here because the input of the
	// functions are not strings, int, bools or time duration.
	// Ref: https://github.com/magefile/mage/blob/master/mg/fn.go#L174-L192
	for _, chart := range charts {
		err = render(ctx, chart)
		if err != nil {
			return err
		}
		err = kubeconform(ctx, chart)
		if err != nil {
			return err
		}
		err = kubescore(ctx, chart)
		if err != nil {
			return err
		}
	}
	return nil
}

func render(_ context.Context, chart kubernetes.HelmChart) error {
	dest, cleanup, err := core.MkdirTemp()
	defer cleanup()
	if err != nil {
		return err
	}
	return kubernetes.RenderTemplates(chart, dest, false)
}

func kubeconform(_ context.Context, chart kubernetes.HelmChart) error {
	return kubernetes.ValidateWithKubeConform(chart)
}

func kubescore(_ context.Context, chart kubernetes.HelmChart) error {
	return kubernetes.ValidateWithKubeScore(chart)
}

// Diff runs a diff for all the helm charts compared to the manin brdnch
func Diff(ctx context.Context) error {
	charts, err := kubernetes.FindHelmCharts(".")
	if err != nil {
		return err
	}
	for _, chart := range charts {
		err = kubernetes.DiffTemplates(chart)
		if err != nil {
			return err
		}
	}
	return nil
}

// List lists the found helm charts
func List(ctx context.Context) error {
	charts, err := kubernetes.FindHelmCharts(".")
	if err != nil {
		return err
	}
	kubernetes.ListHelmCharts(charts)
	return nil
}

// Changes implements a target that check if the current branch has changes
// related to main branch
func Changes(_ context.Context) error {
	changes, err := kubernetes.HasChanges()
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
