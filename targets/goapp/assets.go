package goapp

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/coopnorge/mage/internal/core"
)

// findAssets maps each app with assets to the absolute path of its directory under
// the AssetsEnv directory. Apps whose directory is empty are left out. Without
// AssetsEnv set, no app has assets.
func findAssets(apps []string) (map[string]string, error) {
	root := os.Getenv(AssetsEnv)
	if root == "" {
		return map[string]string{}, nil
	}

	info, err := os.Stat(root)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", AssetsEnv, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s: %s is not a directory", AssetsEnv, root)
	}

	content := map[string]string{}
	missing := []string{}
	for _, app := range apps {
		if app == "." {
			return nil, fmt.Errorf("%s: assets are not supported for an app in the repository root", AssetsEnv)
		}

		directory := filepath.Join(root, app)
		info, err := os.Stat(directory)
		if errors.Is(err, fs.ErrNotExist) {
			missing = append(missing, app)
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("%s: %w", AssetsEnv, err)
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("%s: %s is not a directory", AssetsEnv, directory)
		}

		entries, err := os.ReadDir(directory)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", AssetsEnv, err)
		}
		if len(entries) == 0 {
			continue
		}

		absolute, err := filepath.Abs(directory)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", AssetsEnv, err)
		}
		content[app] = absolute
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf(
			"%s: %s has no directory for %s; create empty directory for an app without assets",
			AssetsEnv, root, strings.Join(missing, ", "),
		)
	}

	return content, nil
}

// stageAssets replaces the app's assets in the build context, at
// var/<app>/assets, with the content of source. An empty source leaves the
// directory empty.
func stageAssets(app, source string) error {
	destination := assetsPath(app)
	if err := os.RemoveAll(destination); err != nil {
		return fmt.Errorf("clear the assets of %s: %w", app, err)
	}
	if err := os.MkdirAll(destination, 0o755); err != nil {
		return fmt.Errorf("create the assets directory of %s: %w", app, err)
	}
	if source == "" {
		return nil
	}
	if err := os.CopyFS(destination, os.DirFS(source)); err != nil {
		return fmt.Errorf("copy the assets of %s: %w", app, err)
	}
	return nil
}

func assetsPath(app string) string {
	return filepath.Join(core.OutputDir, app, "assets")
}
