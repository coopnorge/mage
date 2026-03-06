// Package kubernetes has the concern of validating pallets
package kubernetes

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"text/template"
	"unicode/utf8"

	"github.com/coopnorge/mage/internal/core"
	"github.com/coopnorge/mage/internal/devtool"
	"github.com/coopnorge/mage/internal/git"
	"github.com/coopnorge/mage/internal/github"
)

var (
	helm        devtool.Helm
	kubeconform devtool.KubeConform
	kubescore   devtool.KubeScore
	dyff        devtool.Dyff
)

// HelmChart represents a helmchart with the path env and valuefiles
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
// unknown state
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
	depstatus, _, err := helm.Run(nil, path, "dep", "list", ".")
	if err != nil {
		return fmt.Errorf("failed to check dependencies. Please remove all contents %s/charts. Error: %s", chart.path, err)
	}
	if strings.Contains(depstatus, "missing") {
		_, _, err := helm.Run(nil, path, "dep", "up", ".")
		if err != nil {
			return err
		}
	}
	out, _, err := helm.Run(nil, path, args...)
	if filepath.Ext(dest) != "" {
		fmt.Printf("write to file %s\n", dest)
		return os.WriteFile(dest, []byte(out), 0o644)
	}
	return err
}

// DiffTemplates will create a diff of the rendered templates of a helmchart
// compared to the main branch
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
	if github.InCI() {
		args = append(args, "--output", "github")
	}

	args = append(args, mainFilename, branchFilename)

	fmt.Printf("---\nDiff compared to main of \nchart: %s\nenv: %s\n---\n", chart.path, chart.env)
	out, _, err := dyff.Run(nil, diffDir, args...)

	if github.InCI() {
		path := filepath.Join("var", "kubernetes", "diff", fmt.Sprintf("%s-%s.md", filepath.Base(chart.path), chart.env))
		err := os.MkdirAll(filepath.Dir(path), 0o755)
		if err != nil {
			return err
		}

		title := fmt.Sprintf("%s %s", filepath.Base(chart.path), chart.env)
		changes := strings.Count(out, "!")
		summary := fmt.Sprintf("found %d change(s)", changes)
		if changes > 0 {
			summary = fmt.Sprintf("found **%d** change(s)", changes)
		}
		md, err := diffMarkdownTemplate(title, summary, out, 64000)
		if err != nil {
			return err
		}
		err = os.WriteFile(path, []byte(md), 0o644)
		if err != nil {
			return err
		}

		searchString := fmt.Sprintf("### Kubernetes templates for %s", title)

		found, id, err := github.FindCommentInPR(searchString)
		if err != nil {
			return err
		}
		if found {
			err := github.HideComment(id)
			if err != nil {
				return err
			}
		}
		return github.CreateCommentInPR(path)
	}
	return err
}

func diffMarkdownTemplate(title, summary, diff string, limit int) (string, error) {
	// make sure template are not to long
	diffNote := ""
	if utf8.RuneCountInString(diff) > limit {
		diff = diff[:limit]
		diffNote = fmt.Sprintf("# !!NOTE diff has been cut of because it is longer than %d. Full diff is in action log.", limit)
	}

	// cleanup colorcoding
	const ansi = "[\u001b\u009b][[()#;?]*(?:[0-9]{1,4}(?:;[0-9]{0,4})*)?[0-9A-ORZcf-nqry=><]"
	re := regexp.MustCompile(ansi)
	diff = re.ReplaceAllString(diff, "")
	data := map[string]string{
		"Title":    title,
		"Summary":  summary,
		"Diff":     diff,
		"DiffNote": diffNote,
	}

	funcMap := template.FuncMap{
		"tripplebacktick": func() string { return "```" },
	}

	const mdTemplate = `### Kubernetes templates for {{.Title}}

<details><summary>{{ .Summary }}</summary>
{{.DiffNote}}
{{tripplebacktick}}diff
{{ .Diff }}
{{tripplebacktick}}
</details>
`
	tmpl, err := template.New("md").Funcs(funcMap).Parse(mdTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	return buf.String(), err
}

// FindHelmCharts will search through the base directory to find the
// all helm charts
func FindHelmCharts(base string) ([]HelmChart, error) {
	directories := []string{}
	charts := []HelmChart{}

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
		envs, err := detectHelmEnvironments(dir)
		if err != nil {
			return charts, err
		}
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

// ListHelmCharts list the found helm charts in this repository
func ListHelmCharts(charts []HelmChart) {
	for _, chart := range charts {
		fmt.Printf("---\n")
		fmt.Printf("path: %s\n", chart.path)
		fmt.Printf("environment: %s\n", chart.env)
		fmt.Printf("valueFiles: [\"%s\"]\n", strings.Join(chart.valueFiles, "\", \""))
	}
}

// ValidateWithKubeConform will run kubeconform validation on a supplied
// HelmChart
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
	files, err := core.ListFilesRecursively(dest, "*.yaml")
	if err != nil {
		return err
	}
	args = append(args, files...)
	github.StartLogGroup("kubeconform")
	out, _, err := kubeconform.Run(nil, dest, args...)
	github.EndLogGroup()
	if github.InCI() && err != nil {
		err := github.PrintActionMessage("error", fmt.Sprintf("kubeconform failed for %s %s", filepath.Base(chart.path), chart.env), out)
		if err != nil {
			return err
		}
	}
	return err
}

// ValidateWithKubeScore will run kube-score validation on a supplied HelmChart
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

	files, err := core.ListFilesRecursively(dest, "*.yaml")
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return nil
	}
	args = append(args, files...)
	github.StartLogGroup("kube-score")
	out, _, err := kubescore.Run(nil, dest, args...)
	github.EndLogGroup()
	if github.InCI() && err != nil {
		err := github.PrintActionMessage("error", fmt.Sprintf("kubecore failed for %s %s", filepath.Base(chart.path), chart.env), out)
		if err != nil {
			return err
		}
	}
	return err
}

// HasChanges checks if the current branch has helmchart changes
// from the main branch
func HasChanges() (bool, error) {
	changedFiles, err := git.DiffToMain()
	if err != nil {
		return false, err
	}
	charts, err := FindHelmCharts(".")
	if err != nil {
		return false, err
	}
	paths := []string{}
	for _, chart := range charts {
		paths = append(paths, chart.path)
	}
	// always trigger on go.mod/sum and workflows because of changes in ci.
	additionalGlobs := []string{"go.mod", "go.sum", ".github/workflows/*"}
	return core.CompareChangesToPaths(changedFiles, paths, additionalGlobs)
}

// findHelmValues will find value yaml files for a  specific environment. It
// will return them in the correct rendering order.
// order of finding value files is
// case only env files
// values.yaml, values-<env>.yaml
// case with extra name
// values.yaml, values-<name>.yaml, values-<name>-<env>.yaml
func findHelmValues(dir string, env string) ([]string, error) {
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

// detectHelmEnvironments will try to detect all environments for helm values
func detectHelmEnvironments(dir string) ([]string, error) {
	// Try to detect environments
	environments := []string{}
	allEnvironmentFiles, err := filepath.Glob(fmt.Sprintf("%s/values-*.yaml", dir))
	if err != nil {
		return environments, err
	}
	for _, environmentFile := range allEnvironmentFiles {
		environmentFileSlice := strings.Split(environmentFile, "-")
		environment := strings.Split(environmentFileSlice[len(environmentFileSlice)-1], ".")[0]
		if slices.Contains(environments, environment) {
			continue
		}
		environments = append(environments, environment)
	}
	return environments, nil
}
