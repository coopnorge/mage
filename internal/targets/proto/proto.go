package proto

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/coopnorge/mage/internal/devtool"
	"github.com/coopnorge/mage/internal/git"
	devtoolTarget "github.com/coopnorge/mage/internal/targets/devtool"
	"github.com/magefile/mage/mg"
)

var (
	//go:embed tools.Dockerfile
	toolsDockerfile string
)

// Generate all code
func Generate(ctx context.Context) error {
	mg.CtxDeps(ctx, Validate, mg.F(generate, "."))
	return nil
}

func generate(ctx context.Context, input string) error {
	mg.CtxDeps(ctx, mg.F(devtoolTarget.Build, "buf", toolsDockerfile))
	return devtool.Run("buf", "buf", "generate", input)
}

// Validate all code
func Validate(ctx context.Context) error {
	mg.CtxDeps(ctx, mg.F(validate, "."))
	return nil
}

func validate(ctx context.Context, input string) error {
	mg.SerialCtxDeps(ctx, mg.F(lint, input), mg.F(formatCheck, input), mg.F(breaking, input))
	return nil
}

func lint(ctx context.Context, input string) error {
	mg.CtxDeps(ctx, mg.F(devtoolTarget.Build, "buf", toolsDockerfile))
	return devtool.Run("buf", "buf", "lint", input)
}

func formatCheck(ctx context.Context, input string) error {
	mg.CtxDeps(ctx, mg.F(devtoolTarget.Build, "buf", toolsDockerfile))
	return devtool.Run("buf", "buf", "format", "--diff", "--exit-code", input)
}

func breaking(ctx context.Context, input string) error {
	mg.CtxDeps(ctx, mg.F(devtoolTarget.Build, "buf", toolsDockerfile))
	tag, err := git.LatestTag()
	if err != nil {
		return err
	}
	return devtool.Run("buf", "buf", "breaking", input, "--against", fmt.Sprintf(".git#tag=%s", tag))
}

// Fix files
func Fix(ctx context.Context) error {
	mg.CtxDeps(ctx, mg.F(fix, "."))
	return nil
}

func fix(ctx context.Context, input string) error {
	mg.CtxDeps(ctx, mg.F(devtoolTarget.Build, "buf", toolsDockerfile))
	return devtool.Run("buf", "buf", "format", input, "--write")
}
