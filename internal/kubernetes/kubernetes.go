// Package kubernetes has the concern of validating pallets
package kubernetes

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/coopnorge/mage/internal/core"
	"github.com/coopnorge/mage/internal/devtool"
	"github.com/coopnorge/mage/internal/git"
)

var (
	helm        devtool.Helm
	kubeconform devtool.KubeConform
	kubescore   devtool.KubeScore
	dyff        devtool.Dyff
)

type HelmChart struct {
	path       string
	env        string
	valueFiles []string
}

func isHelmChart(p string, d fs.DirEntry) bool {
	if !d.IsDir() {
		return false
	}
	return core.FileExists(filepath.Join(p, "Chart.yaml"))
}

// RenderTemplates renders the templates of a specific helm chart. It will
// return a function for cleanup
// When second argument is set to true it will try to render even if some
// files are not there. This is used when rendering a template which is in
// unkown state
func RenderTemplates(chart HelmChart, try bool) (string, func(), error) {
	outdir, cleanup, err := core.MkdirTemp()
	if err != nil {
		return outdir, nil, err
	}
	if try {
		// if the chart does not exist it will just return an empty dir, which
		// we can diff against
		if !core.FileExists(filepath.Join(chart.path, "Chart.yaml")) {
			return outdir, cleanup, nil
		}
	}

	valueFilesFlags := []string{}
	for _, file := range chart.valueFiles {
		fp := filepath.Join(chart.path, file)
		if try {
			// when in try, continue if file does not exist
			if !core.FileExists(fp) {
				continue
			}
		}
		valueFilesFlags = append(valueFilesFlags, "--values")
		valueFilesFlags = append(valueFilesFlags, fp)
	}
	args := []string{}
	args = append(args, "template")
	args = append(args, chart.path)
	args = append(args, "--output-dir")
	args = append(args, outdir)
	args = append(args, valueFilesFlags...)

	return outdir, cleanup, helm.Run(nil, args...)
}

func DiffTemplates(chart HelmChart) error {
	// dyff between a/helloworld/charts/app/templates/ b/helloworld/charts/app/templates/ -o github
	currentBranch, err := git.CurrentBranch()
	if err != nil {
		return err
	}

	branchTemplates, branchDirCleanup, err := RenderTemplates(chart, false)
	defer branchDirCleanup()
	if err != nil {
		return err
	}

	mainWorktree, worktreeCleanup, err := git.Worktree("main")
	defer worktreeCleanup()
	if err != nil {
		return err
	}
	// create a chart object for the chart in the main branch
	mainChart := HelmChart{
		path:       filepath.Join(mainWorktree, chart.path),
		env:        chart.env,
		valueFiles: chart.valueFiles,
	}
	mainTemplates, mainBranchCleanup, err := RenderTemplates(mainChart, true)
	defer mainBranchCleanup()
	if err != nil {
		return err
	}

	args := []string{"between"}
	env := make(map[string]string)
	// simply assumming that if CI is set, we are in github actions
	if _, found := os.LookupEnv("CI"); found {
		args = append(args, "--output", "github")
		env["OUTPUT_FILE"] = fmt.Sprintf("%s-%s-%s.diff", filepath.Base(chart.path), chart.env, currentBranch)
	}
	args = append(args, mainTemplates, branchTemplates)
	return dyff.Run(env, args...)
}

// FindHelmCharts will search through the base directory to find the
// all helm charts
func FindHelmCharts(base string) ([]HelmChart, error) {
	directories := []string{}
	charts := []HelmChart{}
	envs := []string{"dev", "test", "staging", "production"}

	err := filepath.WalkDir(base, func(workDir string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if core.IsDotDirectory(workDir, d) {
			return filepath.SkipDir
		}
		if !isHelmChart(workDir, d) {
			return nil
		}
		directories = append(directories, workDir)
		return nil
	})
	if err != nil {
		return nil, err
	}
	for _, dir := range directories {
		for _, env := range envs {
			valueFiles, err := findHelmValues(dir, env)
			if err != nil {
				return nil, err
			}
			// skip if we find no env specific values
			if len(valueFiles) == 0 {
				continue
			}
			slices.Reverse(valueFiles)
			charts = append(charts, HelmChart{
				path:       dir,
				env:        env,
				valueFiles: valueFiles,
			})
		}
	}
	return charts, nil
}

func ListHelmCharts(charts []HelmChart) {
	for _, chart := range charts {
		fmt.Sprintf("---\n")
		fmt.Sprintf("path: %s\n", chart.path)
		fmt.Sprintf("environment: %s\n", chart.env)
		fmt.Sprintf("valueFiles: [%s]\n", strings.Join(chart.valueFiles, "\", \""))
	}
}

func ValidateWithKubeConform(chart HelmChart) error {
	dir, cleanup, err := RenderTemplates(chart, false)
	defer cleanup()
	if err != nil {
		return err
	}
	args := []string{
		"-schema-location", "default",
		"--schema-location", "https://raw.githubusercontent.com/coopnorge/kubernetes-schemas/main/api-platform/{{ .ResourceKind }}{{ .KindSuffix }}.json",
	}
	files, err := core.ListRescursiveFiles(dir, "*.yaml")
	if err != nil {
		return err
	}
	args = append(args, files...)
	return kubeconform.Run(nil, args...)
}

func ValidateWithKubeScore(chart HelmChart) error {
	dir, cleanup, err := RenderTemplates(chart, false)
	defer cleanup()
	if err != nil {
		return err
	}
	args := []string{
		"score",
	}
	files, err := core.ListRescursiveFiles(dir, "*.yaml")
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return nil
	}
	args = append(args, files...)
	return kubescore.Run(nil, args...)
}

func findHelmValues(dir string, env string) ([]string, error) {
	// order of finding value files is
	// case only env files
	// values.yaml, values-<env>.yaml
	// case with extra name
	// values.yaml, values-<name>.yaml, values-<name>-<env>.yaml
	// We are finding in reverse because if no env values are found we assume
	// no env
	files := []string{}
	pattern := fmt.Sprintf("%s/values-*-%s.yaml", dir, env)
	envValues, err := filepath.Glob(pattern)
	if err != nil {
		return []string{}, err
	}
	// specific named value files exists
	if len(envValues) > 0 {
		for _, envval := range envValues {
			files = append(files, filepath.Base(envval))
		}
		if core.FileExists(filepath.Join(dir, "values.yaml")) {
			files = append(files, "values.yaml")
		}
		return files, nil
	}

	if core.FileExists(filepath.Join(dir, fmt.Sprintf("values-%s.yaml", env))) {
		files = append(files, fmt.Sprintf("values-%s.yaml", env))
	}
	// no env files are found, returning a chart without value files
	if len(files) == 0 {
		return files, nil
	}
	if core.FileExists(filepath.Join(dir, "values.yaml")) {
		files = append(files, "values.yaml")
	}
	return files, nil
}
