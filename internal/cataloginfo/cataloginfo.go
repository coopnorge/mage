package cataloginfo

import (
	"fmt"
	"os"

	"github.com/coopnorge/mage/internal/core"
	"github.com/coopnorge/mage/internal/devtool"
	"github.com/coopnorge/mage/internal/git"
)

// Validate validates the catalog-info files
func Validate() error {
	return DevtoolCatalogInfo(getCatalogInfoPaths()...)
}

// HasChanges checks if the current branch has policy bot config file changes
// from the main branch
func HasChanges() (bool, error) {
	changedFiles, err := git.DiffToMain()
	if err != nil {
		return false, err
	}
	// always trigger on go.mod/sum and workflows because of changes in ci.
	additionalGlobs := []string{"go.mod", "go.sum", ".github/workflows/*"}
	return core.CompareChangesToPaths(changedFiles, getCatalogInfoPaths(), additionalGlobs)
}

// DevtoolCatalogInfo runs the devtool for backstage-entity-validator
func DevtoolCatalogInfo(args ...string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	dockerArgs := []string{
		"--volume", fmt.Sprintf("%s:/src", cwd),
		"--workdir", "/src",
	}

	return devtool.Run("backstage-entity-validator", dockerArgs, "validate-entity", args...)
}

func getCatalogInfoPaths() []string {
	// This is configured here: https://github.com/coopnorge/backstage/blob/54a68fc5202c1b3e3bd492d4f54f2254aef553a9/backstage/app-config.yaml#L94
	return []string{"catalog-info*.yaml"}
}
