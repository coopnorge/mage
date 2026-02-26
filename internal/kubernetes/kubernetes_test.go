package kubernetes

import (
	_ "embed"
	"testing"

	"github.com/magefile/mage/sh"
	"github.com/stretchr/testify/assert"
)

func TestFindHelmCharts(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		workdir string
		want    []HelmChart
		wantErr bool
	}{
		{
			name:    "Should find all relevant charts with envs",
			workdir: "testdata/repo",
			want: []HelmChart{
				{
					path:       "infrastructure/kubernetes/helm/charts/charta",
					env:        "production",
					valueFiles: []string{"values.yaml", "values-production.yaml"},
				},
				{
					path:       "infrastructure/kubernetes/helm/charts/charta",
					env:        "staging",
					valueFiles: []string{"values.yaml", "values-staging.yaml"},
				},
				{
					path:       "infrastructure/kubernetes/helm/charts/chartb",
					env:        "dev",
					valueFiles: []string{"values.yaml", "values-this-dev.yaml"},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Chdir(tt.workdir)
			got, gotErr := FindHelmCharts(".")
			assert.NoError(t, gotErr)
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestRenderHelmChart(t *testing.T) {
	tests := []struct {
		name  string
		chart HelmChart
	}{
		{
			name: "simple chart should render",
			chart: HelmChart{
				env:        "staging",
				path:       "testdata/repo/infrastructure/kubernetes/helm/charts/charta",
				valueFiles: []string{"values.yaml", "values-staging.yaml"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, cleanup, err := RenderTemplates(tt.chart, false)
			assert.NoError(t, err)
			assert.NoError(t, sh.RunV("git", "--no-pager", "diff", "--no-index", dir, "testdata/ref-data/chart-a-staging/"))
			t.Cleanup(cleanup)
		})
	}
}

func TestKubeConform(t *testing.T) {
	tests := []struct {
		name    string
		chart   HelmChart
		wantErr bool
	}{
		{
			name: "KubeConform should pass",
			chart: HelmChart{
				env:        "staging",
				path:       "testdata/repo/infrastructure/kubernetes/helm/charts/charta",
				valueFiles: []string{"values.yaml", "values-staging.yaml"},
			},
			wantErr: false,
		},
		{
			name: "KubeConform should fail",
			chart: HelmChart{
				env:        "production",
				path:       "testdata/repo/infrastructure/kubernetes/helm/charts/charta",
				valueFiles: []string{"values.yaml", "values-production-fail.yaml"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateWithKubeConform(tt.chart)
			if tt.wantErr {
				assert.Error(t, err, tt.name)
			} else {
				assert.NoError(t, err, tt.name)
			}
		})
	}
}

func TestKubeScore(t *testing.T) {
	tests := []struct {
		name    string
		chart   HelmChart
		wantErr bool
	}{
		{
			name: "KubeScore should pass",
			chart: HelmChart{
				env:        "staging",
				path:       "testdata/repo/infrastructure/kubernetes/helm/charts/chartc",
				valueFiles: []string{"values.yaml"},
			},
			wantErr: false,
		},
		{
			name: "KubeScore should fail",
			chart: HelmChart{
				env:        "production",
				path:       "testdata/repo/infrastructure/kubernetes/helm/charts/chartc",
				valueFiles: []string{"values.yaml", "inject-fail.yaml"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateWithKubeScore(tt.chart)
			if tt.wantErr {
				if assert.Error(t, err, tt.name) { //&& tt.errMsg != "" {
					// assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err, tt.name)
			}
		})
	}
}
