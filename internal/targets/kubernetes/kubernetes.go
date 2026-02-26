package terraform

import (
	"context"
	_ "embed"

	"github.com/coopnorge/mage/internal/kubernetes"
	"github.com/magefile/mage/mg"
)

// Validate runs kubeconform, kube-score and render templates
func Validate(ctx context.Context) error {
	charts, err := kubernetes.FindHelmCharts(".")
	if err != nil {
		return err
	}

	var renders []any
	var kubeconforms []any
	var kubescores []any
	for _, chart := range charts {
		renders = append(renders, mg.F(render, chart))
		kubeconforms = append(kubeconforms, mg.F(kubernetes.ValidateWithKubeConform, chart))
		kubescores = append(kubescores, mg.F(kubernetes.ValidateWithKubeScore, chart))
	}

	mg.CtxDeps(ctx, renders...)
	mg.CtxDeps(ctx, append(kubeconforms, kubescores)...)
	return nil
}

func render(_ context.Context, chart kubernetes.HelmChart) error {
	_, cleanup, err := kubernetes.RenderTemplates(chart, false)
	defer cleanup()
	return err
}

// Diff runs a diff for all the helm charts compared to the manin brdnch
func Diff(ctx context.Context) error {
	charts, err := kubernetes.FindHelmCharts(".")
	if err != nil {
		return err
	}
	var diffs []any
	for _, chart := range charts {
		diffs = append(diffs, mg.F(kubernetes.DiffTemplates, chart))
	}

	mg.SerialCtxDeps(ctx, diffs...)
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
