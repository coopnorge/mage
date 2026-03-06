package kubernetes

import (
	"testing"

	"github.com/coopnorge/mage/internal/core"
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
				{
					path:       "infrastructure/kubernetes/helm/charts/charta",
					env:        "fail",
					valueFiles: []string{"values.yaml", "values-production-fail.yaml"},
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
				path:       "internal/kubernetes/testdata/repo/infrastructure/kubernetes/helm/charts/charta",
				valueFiles: []string{"values.yaml", "values-staging.yaml"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, cleanup, err := core.MkdirTemp()
			assert.NoError(t, err, "failed to create temp dir %s", err)
			assert.NoError(t, RenderTemplates(tt.chart, dir, false), "failed to render template")
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
				path:       "internal/kubernetes/testdata/repo/infrastructure/kubernetes/helm/charts/charta",
				valueFiles: []string{"values.yaml", "values-staging.yaml"},
			},
			wantErr: false,
		},
		{
			name: "KubeConform should fail",
			chart: HelmChart{
				env:        "production",
				path:       "internal/kubernetes/testdata/repo/infrastructure/kubernetes/helm/charts/charta",
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
				path:       "internal/kubernetes/testdata/repo/infrastructure/kubernetes/helm/charts/chartc",
				valueFiles: []string{"values.yaml"},
			},
			wantErr: false,
		},
		{
			name: "KubeScore should fail",
			chart: HelmChart{
				env:        "production",
				path:       "internal/kubernetes/testdata/repo/infrastructure/kubernetes/helm/charts/chartc",
				valueFiles: []string{"values.yaml", "inject-fail.yaml"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateWithKubeScore(tt.chart)
			if tt.wantErr {
				assert.Error(t, err, tt.name)
			} else {
				assert.NoError(t, err, tt.name)
			}
		})
	}
}

func TestTemplateRender(t *testing.T) {
	tests := []struct {
		name    string
		title   string
		summary string
		diff    string
		limit   int
		want    string
	}{
		{
			name:    "Template should render",
			title:   "Diff for testing",
			summary: "Some stuff changed",
			diff: `@@ spec.hosts.0 @@
# networking.istio.io/v1beta1/ServiceEntry/coop
! ± value change
- api.staging.coopa
+ api.staging.coop`,
			limit: 64000,
			want:  "### Kubernetes templates for Diff for testing\n\n<details><summary>Some stuff changed</summary>\n\n```diff\n@@ spec.hosts.0 @@\n# networking.istio.io/v1beta1/ServiceEntry/coop\n! ± value change\n- api.staging.coopa\n+ api.staging.coop\n```\n</details>\n",
		},
		{
			name:    "Template should cutoff",
			title:   "Diff for testing",
			summary: "Some stuff changed",
			diff: `@@ spec.hosts.0 @@
# networking.istio.io/v1beta1/ServiceEntry/coop
! ± value change
- api.staging.coopa
+ api.staging.coop`,
			limit: 60,
			want:  "### Kubernetes templates for Diff for testing\n\n<details><summary>Some stuff changed</summary>\n# !!NOTE diff has been cut of because it is longer than 60. Full diff is in action log.\n```diff\n@@ spec.hosts.0 @@\n# networking.istio.io/v1beta1/ServiceEntr\n```\n</details>\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diff, err := diffMarkdownTemplate(tt.title, tt.summary, tt.diff, tt.limit)
			assert.NoError(t, err, tt.name)
			assert.Equal(t, tt.want, diff)
		})
	}
}
