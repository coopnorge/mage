package javascript

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/coopnorge/mage/internal/core"
	"github.com/magefile/mage/sh"
)

// IsNpmrcConfiguredForPrivateRepo checks if the .npmrc file is configured for GitHub
// Packages.
func IsNpmrcConfiguredForPrivateRepo(directory string) bool {
	if directory == "" {
		directory = "."
	}

	registryURL := "npm.pkg.github.com"
	scope := "@coopnorge"
	tokenIndicator := "_authToken="

	npmrcContent, err := os.ReadFile(fmt.Sprintf("%s/.npmrc", directory))

	if err != nil {
		return false
	}

	contentStr := string(npmrcContent)

	if !strings.Contains(contentStr, registryURL) && !strings.Contains(contentStr, scope) && !strings.Contains(contentStr, tokenIndicator) {
		return false
	}

	return true
}

func HasBiomeConfig() bool {
	return core.FileExists("biome.json")
}

func HasPackageConfig() bool {
	return core.FileExists("package.json")
}
