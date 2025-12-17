// Package pallets has the concern of validating pallets
package pallets

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/coopnorge/mage/internal/core"
	"github.com/coopnorge/mage/internal/devtool"
	"github.com/coopnorge/mage/internal/git"
)

// Validate submits policy file to policy-bot docker app to validate it
func Validate() error {
	palletList, err := getPalletFiles()
	if err != nil {
		return err
	}
	if len(palletList) == 0 {
		return nil
	}

	palletMountPath := "./.pallet"
	dockerArgs := []string{
		"--volume", fmt.Sprintf("%s:%s", palletMountPath, "/.pallet"),
		"--workdir", "/",
	}
	//  kubeconform --strict -verbose  -schema-location "https://raw.githubusercontent.com/coopnorge/kubernetes-schemas/main/pallets/{{ .ResourceKind }}{{ .KindSuffix }}.json" .pallet/gitconfig.yaml

	cmd := "--strict"
	args := []string{
		"--verbose",
		"--schema-location",
		"https://raw.githubusercontent.com/coopnorge/kubernetes-schemas/main/pallets/{{ .ResourceKind }}{{ .KindSuffix }}.json",
	}
	args = append(args, palletList...)
	return devtool.Run("kubeconform", dockerArgs, cmd, args...)
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
