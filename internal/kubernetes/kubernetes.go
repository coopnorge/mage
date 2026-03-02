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

// RenderTemplates renders the templates of a specific helm chart. It required
// a destination. If the dest is a folder it will render the files separate. If
// it is a file, then it will render all in 1 temiplate.
// When third argument is set to true it will try to render even if some
// files are not there. This is used when rendering a template which is in
// unkown state
func RenderTemplates(chart HelmChart, dest string, try bool) error {
	if try {
		// if the chart does not exist it will just return an empty dir, which
		// we can diff against
		if !core.FileExists(filepath.Join(chart.path, "Chart.yaml")) {
			return nil
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
		valueFilesFlags = append(valueFilesFlags, file)
	}
	args := []string{}
	args = append(args, "template")
	args = append(args, valueFilesFlags...)
	if filepath.Ext(dest) == "" {
		args = append(args, "--output-dir")
		args = append(args, dest)
	}
	args = append(args, ".")

	// make path abs when it is not, required for running in docker
	path := chart.path
	if filepath.IsLocal(chart.path) {
		base, err := core.GetRepoRoot()
		if err != nil {
			return err
		}
		path = filepath.Join(base, chart.path)
	}
	// make sure dependencies are there
	_, _, err := helm.Run(nil, path, "dep", "up", ".")
	if err != nil {
		return err
	}
	out, _, err := helm.Run(nil, path, args...)
	if filepath.Ext(dest) != "" {
		fmt.Printf("write to file %s\n", dest)
		return os.WriteFile(dest, []byte(out), 0o644)
	}
	return err
}

func DiffTemplates(chart HelmChart) error {
	// dyff between a/helloworld/charts/app/templates/ b/helloworld/charts/app/templates/ -o github

	diffDir, cleanupDiffDir, err := core.MkdirTemp()
	defer cleanupDiffDir()
	if err != nil {
		return err
	}
	branchFilename := fmt.Sprintf("branch-%s-%s.yaml", filepath.Base(chart.path), chart.env)
	mainFilename := fmt.Sprintf("main-%s-%s.yaml", filepath.Base(chart.path), chart.env)

	err = RenderTemplates(chart, filepath.Join(diffDir, branchFilename), false)
	if err != nil {
		return err
	}

	mainWorktree, worktreeCleanup, err := git.Worktree("main")
	if err != nil {
		return err
	}
	defer worktreeCleanup()
	// create a chart object for the chart in the main branch
	mainChart := HelmChart{
		path:       filepath.Join(mainWorktree, chart.path),
		env:        chart.env,
		valueFiles: chart.valueFiles,
	}

	err = RenderTemplates(mainChart, filepath.Join(diffDir, mainFilename), true)
	if err != nil {
		return err
	}

	args := []string{
		"--color", "on",
		"--truecolor", "on",
		"between",
	}
	// simply assumming that if CI is set, we are in github actions
	_, inCI := os.LookupEnv("CI")
	if inCI {
		args = append(args, "--output", "github")
	}

	args = append(args, branchFilename, mainFilename)

	fmt.Printf("---\nDiff compared to main of \nchart: %s\nenv: %s\n---\n", chart.path, chart.env)
	out, _, err := dyff.Run(nil, diffDir, args...)

	if inCI {
		path := filepath.Join("var", "kubernetes", "diff", fmt.Sprintf("%s-%s.diff", filepath.Base(chart.path), chart.env))
		err := os.MkdirAll(filepath.Dir(path), 0o755)
		if err != nil {
			return err
		}
		if out == "" {
			out = fmt.Sprintf("# no diff for %s %s", filepath.Base(chart.path), chart.env)
		}
		err = os.WriteFile(path, []byte(out), 0o644)
		if err != nil {
			return err
		}
	}
	return err
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
		fmt.Printf("---\n")
		fmt.Printf("path: %s\n", chart.path)
		fmt.Printf("environment: %s\n", chart.env)
		fmt.Printf("valueFiles: [\"%s\"]\n", strings.Join(chart.valueFiles, "\", \""))
	}
}

func ValidateWithKubeConform(chart HelmChart) error {
	dest, cleanup, err := core.MkdirTemp()
	defer cleanup()
	if err != nil {
		return err
	}
	err = RenderTemplates(chart, dest, false)
	if err != nil {
		return err
	}
	args := []string{
		"-schema-location", "default",
		"--schema-location", "https://raw.githubusercontent.com/coopnorge/kubernetes-schemas/main/api-platform/{{ .ResourceKind }}{{ .KindSuffix }}.json",
	}
	files, err := core.ListRescursiveFiles(dest, "*.yaml")
	if err != nil {
		return err
	}
	args = append(args, files...)
	_, _, err = kubeconform.Run(nil, dest, args...)
	return err
}

func ValidateWithKubeScore(chart HelmChart) error {
	dest, cleanup, err := core.MkdirTemp()
	defer cleanup()
	if err != nil {
		return err
	}
	err = RenderTemplates(chart, dest, false)
	if err != nil {
		return err
	}
	args := []string{
		"score",
	}

	files, err := core.ListRescursiveFiles(dest, "*.yaml")
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return nil
	}
	args = append(args, files...)
	// if filepath.IsLocal(dir) {
	// 	root, err := core.GetRepoRoot()
	// 	if err != nil {
	// 		return err
	// 	}
	// 	dir = filepath.Join(root, dir)
	// }
	_, _, err = kubescore.Run(nil, dest, args...)
	return err
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
