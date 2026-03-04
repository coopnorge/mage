package testhelpers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/template"

	"github.com/magefile/mage/sh"
	"github.com/stretchr/testify/require"
)

var goModTemplateString = `module dummy
go {{ .GoVersion }}

require github.com/coopnorge/mage v0.0.0 // Version does not matter when using replace

tool github.com/magefile/mage
replace github.com/coopnorge/mage => {{ .MageRoot }}`

// CreateGoModuleFactory returns a function that creates a go module in the current directory with a go.mod file that has a replace directive pointing to the mage root.
// It's safe to call this function multiple times, e.g. once per test.
// This function should be called before any directory changes are made in the tests, to ensure that it reads the mage root's go.mod file, instead of any go.mod/go.sum file created in testdata.
func CreateGoModuleFactory(t *testing.T) func(t *testing.T) {
	t.Helper()
	// Read parameters from environment before sub-tests are run, to ensure that we read the root mage's directory, instead of anything created in testdata.
	goModMage, err := sh.Output("go", "env", "GOMOD")
	require.NoError(t, err, "failed to read path to mage root's go.mod file")

	goVersion, err := sh.Output("go", "env", "GOVERSION")
	require.NoError(t, err, "failed to read go version from environment")
	goVersion = strings.TrimPrefix(goVersion, "go") // Convert "go1.26.0" to "1.26.0"

	// Create template
	mageRoot := filepath.Dir(goModMage)
	goModTemplate, err := template.New("gomod").Parse(goModTemplateString)
	require.NoError(t, err, "failed to parse go.mod template")

	return func(t *testing.T) {
		t.Helper()
		// Create files from template
		goMod, err := os.Create("go.mod")
		require.NoError(t, err)
		err = goModTemplate.Execute(goMod, map[string]any{
			"GoVersion": goVersion,
			"MageRoot":  mageRoot,
		})
		require.NoError(t, err)
		err = goMod.Close()
		require.NoError(t, err)

		// Tidy depndencies to get go.sum file and list all indirect dependency.
		err = sh.RunV("go", "mod", "tidy")
		require.NoError(t, err)
	}
}
