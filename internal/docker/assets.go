package docker

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/coopnorge/mage/internal/core"
)

// FindAssets maps each app with assets to the absolute path of its directory
// under root, which holds one directory per app. Apps whose directory is empty
// are left out. With an empty root, no app has assets.
func FindAssets(root string, apps []string) (map[string]string, error) {
	if root == "" {
		return map[string]string{}, nil
	}

	info, err := os.Stat(root)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a directory", root)
	}

	content := map[string]string{}
	missing := []string{}
	for _, app := range apps {
		if app == "." {
			return nil, errors.New("assets are not supported for an app in the repository root")
		}

		directory := filepath.Join(root, app)
		info, err := os.Stat(directory)
		if errors.Is(err, fs.ErrNotExist) {
			missing = append(missing, app)
			continue
		}
		if err != nil {
			return nil, err
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("%s is not a directory", directory)
		}

		entries, err := os.ReadDir(directory)
		if err != nil {
			return nil, err
		}
		if len(entries) == 0 {
			continue
		}

		absolute, err := filepath.Abs(directory)
		if err != nil {
			return nil, err
		}
		content[app] = absolute
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf(
			"%s has no directory for %s; create empty directory for an app without assets",
			root, strings.Join(missing, ", "),
		)
	}

	return content, nil
}

// StageAssets replaces the app's assets in the build context, at
// var/<app>/assets, with the content of source. An empty source leaves the
// directory empty.
func StageAssets(app, source string) error {
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
