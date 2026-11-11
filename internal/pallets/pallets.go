// Package pallets has the concern of validating pallets
package pallets

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"

	"github.com/coopnorge/mage/internal/cataloginfo"
	"github.com/coopnorge/mage/internal/core"
	"github.com/coopnorge/mage/internal/devtool"
	"github.com/coopnorge/mage/internal/git"
	"gopkg.in/yaml.v3"
)

var kubeconform devtool.KubeConform

// Validate submits policy file to policy-bot docker app to validate it
func Validate() error {
	palletList, err := getPalletFiles()
	if err != nil {
		return err
	}
	if len(palletList) == 0 {
		return nil
	}
	if err := validatePalletOwners(palletList); err != nil {
		return err
	}

	// palletMountPath := "./.pallet"
	// dockerArgs := []string{
	// 	"--volume", fmt.Sprintf("%s:%s", palletMountPath, "/.pallet"),
	// 	"--workdir", "/",
	// }
	// //  kubeconform --strict -verbose  -schema-location "https://raw.githubusercontent.com/coopnorge/kubernetes-schemas/main/pallets/{{ .ResourceKind }}{{ .KindSuffix }}.json" .pallet/gitconfig.yaml
	//
	// cmd := "--strict"
	args := []string{
		"--schema-location",
		"https://raw.githubusercontent.com/coopnorge/kubernetes-schemas/main/pallets/{{ .ResourceKind }}{{ .KindSuffix }}.json",
	}
	args = append(args, palletList...)
	_, _, err = kubeconform.Run(nil, ".", args...)
	return err
	// return devtool.Run("kubeconform", dockerArgs, cmd, args...)
}

// HasChanges checks if the current branch has policy bot config file changes
// from the main branch
func HasChanges() (bool, error) {
	changedFiles, err := git.DiffToMain()
	if err != nil {
		return false, err
	}
	palletFiles, err := getPalletFiles()
	if err != nil {
		return false, err
	}
	// always trigger on go.mod/sum and workflows because of changes in ci.
	additionalGlobs := []string{"go.mod", "go.sum", ".github/workflows/*"}
	return core.CompareChangesToPaths(changedFiles, palletFiles, additionalGlobs)
}

func getPalletFiles() ([]string, error) {
	palletList := []string{}
	pallets, err := os.ReadDir(".pallet")
	if os.IsNotExist(err) {
		return palletList, nil
	}
	if err != nil {
		return palletList, err
	}

	for _, p := range pallets {
		palletList = append(palletList, filepath.Join(".pallet", p.Name()))
	}
	return palletList, nil
}

type palletOwnerConfig struct {
	Kind string `yaml:"kind"`
	Spec struct {
		Fields map[string]any `yaml:",inline"`
	} `yaml:"spec"`
}

func validatePalletOwners(palletFiles []string) error {
	owners, err := cataloginfo.SystemAndComponentOwners()
	if err != nil {
		return fmt.Errorf("failed to parse catalog-info files: %w", err)
	}

	for _, file := range palletFiles {
		if err := validatePalletOwner(file, owners); err != nil {
			return err
		}
	}
	return nil
}

func validatePalletOwner(file string, owners []string) (err error) {
	f, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("failed to open pallet file %s: %w", file, err)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("failed to close pallet file %s: %w", file, closeErr))
		}
	}()

	dec := yaml.NewDecoder(f)
	for docIdx := 1; ; docIdx++ {
		var config palletOwnerConfig
		decodeErr := dec.Decode(&config)
		if errors.Is(decodeErr, io.EOF) {
			return nil
		}
		if decodeErr != nil {
			return fmt.Errorf("failed to parse YAML doc %d in %s: %w", docIdx, file, decodeErr)
		}
		if config.Kind != "RepositoryConfig" {
			continue
		}
		ownerValue, exists := config.Spec.Fields["owner"]
		if !exists {
			continue
		}
		owner, ok := ownerValue.(string)
		if !ok || owner == "" {
			return fmt.Errorf("doc %d in %s: spec.owner must be a non-empty string when set", docIdx, file)
		}
		if !slices.Contains(owners, owner) {
			return fmt.Errorf("doc %d in %s: spec.owner %q does not match an owner declared by a System or Component in catalog-info",
				docIdx, file, owner)
		}
	}
}
