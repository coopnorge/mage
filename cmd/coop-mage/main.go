// Command coop-mage is a standalone CLI that exposes the coopnorge/mage CI
// targets without requiring consumers to import the mage packages in their own
// magefiles. It is intended to be checked out at a pinned version inside GitHub
// Actions jobs and invoked directly:
//
//   - uses: actions/checkout@v6
//     with:
//     repository: coopnorge/mage
//     ref: v0.x.y
//     path: .coop-mage
//   - run: go build -C .coop-mage -o /tmp/coop-mage ./cmd/coop-mage
//   - run: /tmp/coop-mage git:listChanges
//
// Commands follow the namespace:function convention used by mage targets so
// that the workflow steps remain readable and familiar.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"

	cataloginfoInternal "github.com/coopnorge/mage/internal/cataloginfo"
	"github.com/coopnorge/mage/internal/core"
	"github.com/coopnorge/mage/internal/devtool"
	golangcilint "github.com/coopnorge/mage/internal/devtool/golangci-lint"
	"github.com/coopnorge/mage/internal/docker"
	"github.com/coopnorge/mage/internal/git"
	"github.com/coopnorge/mage/internal/golang"
	"github.com/coopnorge/mage/internal/pallets"
	"github.com/coopnorge/mage/internal/policybot"
	"github.com/coopnorge/mage/internal/terraform"
	kubernetesTargets "github.com/coopnorge/mage/internal/targets/kubernetes"
	terraformTargets "github.com/coopnorge/mage/internal/targets/terraform"
)

const (
	cmdDir = "cmd"
	binDir = "bin"
)

var toolGo devtool.Go

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: coop-mage <namespace:command>\n\navailable commands:\n")
		for name := range commands() {
			fmt.Fprintf(os.Stderr, "  %s\n", name)
		}
		os.Exit(1)
	}

	name := os.Args[1]
	cmds := commands()
	fn, ok := cmds[name]
	if !ok {
		fmt.Fprintf(os.Stderr, "unknown command: %q\n\navailable commands:\n", name)
		for n := range cmds {
			fmt.Fprintf(os.Stderr, "  %s\n", n)
		}
		os.Exit(1)
	}

	if err := fn(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func commands() map[string]func(context.Context) error {
	return map[string]func(context.Context) error{
		"git:listChanges":           gitListChanges,
		"go:changes":                goChanges,
		"go:downloadModules":        goDownloadModules,
		"go:lint":                   goLint,
		"go:test":                   goTest,
		"go:buildBinaries":          goBuildBinaries,
		"docker:buildImages":        dockerBuildImages,
		"docker:buildAndPush":       dockerBuildAndPush,
		"terraform:changes":         terraformChanges,
		"terraform:validate":        terraformValidate,
		"policyBotConfig:changes":   policyBotChanges,
		"policyBotConfig:validate":  policyBotValidate,
		"pallets:changes":           palletsChanges,
		"pallets:validate":          palletsValidate,
		"catalogInfo:changes":       catalogInfoChanges,
		"catalogInfo:validate":      catalogInfoValidate,
		"k8s:changes":               k8sChanges,
		"k8s:list":                  k8sList,
		"k8s:validate":              k8sValidate,
		"k8s:diff":                  k8sDiff,
	}
}

// --- git ---

func gitListChanges(_ context.Context) error {
	changes, err := git.DiffToMain()
	if err != nil {
		return err
	}
	for _, c := range changes {
		fmt.Println(c)
	}
	return nil
}

// --- go ---

func goChanges(_ context.Context) error {
	dirs, err := golang.FindGoSourceCodeFolders(".")
	if err != nil {
		return err
	}
	changes, err := golang.HasChanges(dirs, "Go OCI Release")
	if err != nil {
		return err
	}
	if changes {
		fmt.Println("true")
	} else {
		fmt.Println("false")
	}
	return nil
}

func goDownloadModules(_ context.Context) error {
	dirs, err := golang.FindGoModules(".")
	if err != nil {
		return err
	}
	for _, dir := range dirs {
		if err := golang.DownloadModules(dir); err != nil {
			return err
		}
	}
	return nil
}

func goLint(_ context.Context) error {
	dirs, err := golang.FindGoModules(".")
	if err != nil {
		return err
	}
	cfg := golangcilint.Cfg()
	for _, dir := range dirs {
		if err := golang.Lint(dir, cfg); err != nil {
			return err
		}
	}
	return nil
}

func goTest(_ context.Context) error {
	dirs, err := golang.FindGoModules(".")
	if err != nil {
		return err
	}
	for _, dir := range dirs {
		if err := golang.Test(dir); err != nil {
			return err
		}
	}
	return nil
}

func goBuildBinaries(_ context.Context) error {
	rootPath, err := os.Getwd()
	if err != nil {
		return err
	}

	goModules, err := golang.FindGoModules(".")
	if err != nil {
		return err
	}

	cmds, err := findCommands(goModules)
	if err != nil {
		return err
	}

	for _, cmd := range cmds {
		relativeRootPath, err := core.GetRelativeRootPath(rootPath, cmd.goModule)
		if err != nil {
			return err
		}
		input := strings.Join(cmd.pkgs, " ")

		for _, osArch := range golang.OSArch() {
			outputDir := path.Join(relativeRootPath, binaryOutputDir(cmd.goModule, osArch["GOOS"], osArch["GOARCH"]))
			if err := os.MkdirAll(path.Join(cmd.goModule, outputDir), os.ModePerm); err != nil {
				return err
			}
			if err := buildBinary(cmd.goModule, input, outputDir, osArch["GOOS"], osArch["GOARCH"]); err != nil {
				return err
			}
		}
	}
	return nil
}

// --- docker ---

func dockerBuildImages(_ context.Context) error {
	goModules, err := golang.FindGoModules(".")
	if err != nil {
		return err
	}

	cmds, err := findCommands(goModules)
	if err != nil {
		return err
	}

	push, err := shouldPush()
	if err != nil {
		return err
	}

	for _, cmd := range cmds {
		for _, binary := range cmd.binaries {
			imageName := docker.FullyQualifiedlImageName(cmd.goModule, binary)
			imgPath := imagePath(cmd.goModule, binary)
			metaPath := metadataPath(cmd.goModule, binary)
			if err := docker.BuildAndPush(docker.DefaultDockerfile, golang.DockerPlatforms(), imageName, ".", imgPath, metaPath, cmd.goModule, binary, push); err != nil {
				return err
			}
		}
	}
	return writeImageMetadata()
}

func dockerBuildAndPush(ctx context.Context) error {
	if err := goBuildBinaries(ctx); err != nil {
		return err
	}
	return dockerBuildImages(ctx)
}

// --- terraform ---

func terraformChanges(_ context.Context) error {
	dirs, err := terraform.FindTerraformProjects(".")
	if err != nil {
		return err
	}
	changes, err := terraform.HasChanges(dirs)
	if err != nil {
		return err
	}
	if changes {
		fmt.Println("true")
	} else {
		fmt.Println("false")
	}
	return nil
}

func terraformValidate(_ context.Context) error {
	dirs, err := terraform.FindTerraformProjects(".")
	if err != nil {
		return err
	}
	for _, dir := range dirs {
		if err := terraform.CheckLock(dir); err != nil {
			return err
		}
		if err := terraform.Init(dir); err != nil {
			return err
		}
		if err := terraform.Test(dir); err != nil {
			return err
		}
		if err := terraform.Lint(dir, terraformTargets.TFlintCfg); err != nil {
			return err
		}
		if err := terraform.Security(dir); err != nil {
			return err
		}
	}
	return nil
}

// --- policybot ---

func policyBotChanges(_ context.Context) error {
	changes, err := policybot.HasChanges()
	if err != nil {
		return err
	}
	if changes {
		fmt.Println("true")
	} else {
		fmt.Println("false")
	}
	return nil
}

func policyBotValidate(_ context.Context) error {
	return policybot.Validate()
}

// --- pallets ---

func palletsChanges(_ context.Context) error {
	changes, err := pallets.HasChanges()
	if err != nil {
		return err
	}
	if changes {
		fmt.Println("true")
	} else {
		fmt.Println("false")
	}
	return nil
}

func palletsValidate(_ context.Context) error {
	return pallets.Validate()
}

// --- catalog-info ---

func catalogInfoChanges(_ context.Context) error {
	changes, err := cataloginfoInternal.HasChanges()
	if err != nil {
		return err
	}
	if changes {
		fmt.Println("true")
	} else {
		fmt.Println("false")
	}
	return nil
}

func catalogInfoValidate(_ context.Context) error {
	return cataloginfoInternal.Validate()
}

// --- kubernetes ---

func k8sChanges(ctx context.Context) error {
	return kubernetesTargets.Changes(ctx)
}

func k8sList(ctx context.Context) error {
	return kubernetesTargets.List(ctx)
}

func k8sValidate(ctx context.Context) error {
	return kubernetesTargets.Validate(ctx)
}

func k8sDiff(ctx context.Context) error {
	return kubernetesTargets.Diff(ctx)
}

// --- build helpers ---

type goCmdInfo struct {
	goModule string
	pkgs     []string
	binaries []string
}

func findCommands(goModules []string) ([]goCmdInfo, error) {
	result := []goCmdInfo{}
	for _, mod := range goModules {
		cmdPath := path.Join(mod, cmdDir)
		if _, err := os.Stat(cmdPath); os.IsNotExist(err) {
			continue
		}
		entries, err := os.ReadDir(cmdPath)
		if err != nil {
			return nil, err
		}
		pkgs := []string{}
		bins := []string{}
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			pkgs = append(pkgs, fmt.Sprintf("./%s", path.Join(cmdDir, entry.Name())))
			bins = append(bins, entry.Name())
		}
		if len(pkgs) == 0 {
			continue
		}
		result = append(result, goCmdInfo{goModule: mod, pkgs: pkgs, binaries: bins})
	}
	return result, nil
}

func buildBinary(workDir, input, outputDir, goos, goarch string) error {
	env := map[string]string{"GOOS": goos, "GOARCH": goarch, "CGO_ENABLED": "0"}
	inputs := strings.Split(input, " ")
	args := append([]string{"-C", workDir, "build", "-tags='datadog.no_waf'", "-o", outputDir}, inputs...)
	return toolGo.Run(env, args...)
}

func binaryOutputDir(mod, goos, goarch string) string {
	return path.Join(core.OutputDir, mod, binDir, goos, goarch)
}

func imagePath(mod, binary string) string {
	return path.Join(core.OutputDir, mod, "oci", binary, "image.tar")
}

func metadataPath(mod, binary string) string {
	return path.Join(core.OutputDir, mod, "oci", binary, "metadata.json")
}

func shouldPush() (bool, error) {
	val, ok := os.LookupEnv("PUSH_IMAGE")
	if !ok || val == "" {
		return false, nil
	}
	b := strings.ToLower(strings.TrimSpace(val))
	return b == "true" || b == "1" || b == "yes", nil
}

func writeImageMetadata() error {
	if err := os.MkdirAll(core.OutputDir, 0755); err != nil {
		return err
	}
	images, err := docker.Images(core.OutputDir)
	if err != nil {
		return err
	}
	data, err := json.Marshal(images)
	if err != nil {
		return err
	}
	return os.WriteFile(path.Join(core.OutputDir, "oci-images.json"), data, 0o644)
}
