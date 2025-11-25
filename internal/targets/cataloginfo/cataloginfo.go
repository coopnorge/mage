package cataloginfo

import (
	"context"
	_ "embed"

	"github.com/coopnorge/mage/internal/cataloginfo"
	"github.com/coopnorge/mage/internal/targets/devtool"
	"github.com/magefile/mage/mg"
)

var (
	//go:embed tools.Dockerfile
	// CatalogInfoToolsDockerfile the content of tools.Dockerfile
	CatalogInfoToolsDockerfile string
)

// HasChanges checks if the current branch has any catalog-info changes compared
// to the main branch
func HasChanges() (bool, error) {
	return cataloginfo.HasChanges()
}

// Validate validates catalog-info files
func Validate(ctx context.Context) error {
	mg.CtxDeps(ctx, mg.F(devtool.Build, "backstage-entity-validator", CatalogInfoToolsDockerfile))
	return cataloginfo.Validate()
}
