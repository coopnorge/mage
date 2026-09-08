package cataloginfo

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/coopnorge/mage/internal/core"
	"github.com/coopnorge/mage/internal/git"
	backstage "github.com/datolabs-io/go-backstage/v3"
	"gopkg.in/yaml.v3"
)

type catalogInfoData struct {
	System     *backstage.SystemEntityV1alpha1
	Components []backstage.ComponentEntityV1alpha1
	Resources  []backstage.ResourceEntityV1alpha1
}

type entityHeader struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   struct {
		Name string `yaml:"name"`
	} `yaml:"metadata"`
}

// Validate validates the catalog-info files
func Validate() error {
	_, err := parseCatalogInfoFiles()
	return err
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
	additionalGlobs = append(additionalGlobs, getCatalogInfoPaths()...)
	return core.CompareChangesToPaths(changedFiles, []string{}, additionalGlobs)
}

func parseCatalogInfoFiles() (*catalogInfoData, error) {
	var matches []string
	for _, pattern := range getCatalogInfoPaths() {
		files, err := filepath.Glob(pattern)
		if err != nil {
			return nil, fmt.Errorf("failed to glob pattern %s: %w", pattern, err)
		}
		matches = append(matches, files...)
	}

	data := &catalogInfoData{}
	for _, file := range matches {
		if err := parseCatalogInfoFile(file, data); err != nil {
			return nil, err
		}
	}

	return data, nil
}

func parseCatalogInfoFile(filePath string, data *catalogInfoData) (err error) {
	f, openErr := os.Open(filePath)
	if openErr != nil {
		return fmt.Errorf("failed to open %s: %w", filePath, openErr)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			err = errors.Join(err, fmt.Errorf("failed to close %s: %w", filePath, closeErr))
		}
	}()

	dec := yaml.NewDecoder(f)
	docIdx := 0
	for {
		docIdx++
		var node yaml.Node
		decodeErr := dec.Decode(&node)
		if errors.Is(decodeErr, io.EOF) {
			break
		}
		if decodeErr != nil {
			return fmt.Errorf("failed to parse YAML doc %d in %s: %w", docIdx, filePath, decodeErr)
		}

		if node.Kind == 0 || (node.Kind == yaml.DocumentNode && len(node.Content) == 0) {
			continue
		}

		if docErr := parseDocument(&node, docIdx, filePath, data); docErr != nil {
			return docErr
		}
	}

	return nil
}

func parseDocument(node *yaml.Node, docIdx int, filePath string, data *catalogInfoData) error {
	var header entityHeader
	if err := node.Decode(&header); err != nil {
		return fmt.Errorf("failed to decode entity header in doc %d of %s: %w", docIdx, filePath, err)
	}

	if header.APIVersion == "" {
		return fmt.Errorf("doc %d in %s: missing required field apiVersion", docIdx, filePath)
	}
	if header.Metadata.Name == "" {
		return fmt.Errorf("doc %d in %s: missing required field metadata.name", docIdx, filePath)
	}

	switch header.Kind {
	case backstage.KindSystem:
		return parseSystem(node, docIdx, filePath, data)
	case backstage.KindComponent:
		return parseComponent(node, docIdx, filePath, data)
	case backstage.KindResource:
		return parseResource(node, docIdx, filePath, data)
	default:
		return fmt.Errorf("doc %d in %s: unknown or unsupported entity kind %q", docIdx, filePath, header.Kind)
	}
}

func parseSystem(node *yaml.Node, docIdx int, filePath string, data *catalogInfoData) error {
	if data.System != nil {
		return fmt.Errorf("more than one System defined: found second system in doc %d of %s", docIdx, filePath)
	}
	var sys backstage.SystemEntityV1alpha1
	if err := node.Decode(&sys); err != nil {
		return fmt.Errorf("failed to decode System in doc %d of %s: %w", docIdx, filePath, err)
	}
	if sys.Spec == nil {
		return fmt.Errorf("doc %d in %s: System %q is missing spec", docIdx, filePath, sys.Metadata.Name)
	}
	if sys.Spec.Owner == "" {
		return fmt.Errorf("doc %d in %s: System %q spec is missing owner", docIdx, filePath, sys.Metadata.Name)
	}
	data.System = &sys
	return nil
}

func parseComponent(node *yaml.Node, docIdx int, filePath string, data *catalogInfoData) error {
	var comp backstage.ComponentEntityV1alpha1
	if err := node.Decode(&comp); err != nil {
		return fmt.Errorf("failed to decode Component in doc %d of %s: %w", docIdx, filePath, err)
	}
	if comp.Spec == nil {
		return fmt.Errorf("doc %d in %s: Component %q is missing spec", docIdx, filePath, comp.Metadata.Name)
	}
	if comp.Spec.Type == "" {
		return fmt.Errorf("doc %d in %s: Component %q spec is missing type", docIdx, filePath, comp.Metadata.Name)
	}
	if comp.Spec.Lifecycle == "" {
		return fmt.Errorf("doc %d in %s: Component %q spec is missing lifecycle", docIdx, filePath, comp.Metadata.Name)
	}
	if comp.Spec.Owner == "" {
		return fmt.Errorf("doc %d in %s: Component %q spec is missing owner", docIdx, filePath, comp.Metadata.Name)
	}
	data.Components = append(data.Components, comp)
	return nil
}

func parseResource(node *yaml.Node, docIdx int, filePath string, data *catalogInfoData) error {
	var res backstage.ResourceEntityV1alpha1
	if err := node.Decode(&res); err != nil {
		return fmt.Errorf("failed to decode Resource in doc %d of %s: %w", docIdx, filePath, err)
	}
	if res.Spec == nil {
		return fmt.Errorf("doc %d in %s: Resource %q is missing spec", docIdx, filePath, res.Metadata.Name)
	}
	if res.Spec.Type == "" {
		return fmt.Errorf("doc %d in %s: Resource %q spec is missing type", docIdx, filePath, res.Metadata.Name)
	}
	if res.Spec.Owner == "" {
		return fmt.Errorf("doc %d in %s: Resource %q spec is missing owner", docIdx, filePath, res.Metadata.Name)
	}
	data.Resources = append(data.Resources, res)
	return nil
}

func getCatalogInfoPaths() []string {
	// This is configured here: https://github.com/coopnorge/backstage/blob/54a68fc5202c1b3e3bd492d4f54f2254aef553a9/backstage/app-config.yaml#L94
	return []string{"catalog-info*.yaml"}
}
