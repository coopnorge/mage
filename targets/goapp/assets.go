package goapp

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
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
