package proto

import (
	"context"

	"github.com/coopnorge/mage/internal/targets/proto"
	"github.com/magefile/mage/mg"
)

// Proto is the magefile namespace to group Protocol Buffers commands
type Proto mg.Namespace

const (
	outputDir = "./var"
)

// Generate all code
func (Proto) Generate(ctx context.Context) error {
	mg.SerialCtxDeps(ctx, Proto.Validate, proto.Generate)
	return nil
}

// Validate all code
func (Proto) Validate(ctx context.Context) error {
	mg.CtxDeps(ctx, proto.Validate)
	return nil
}

// Fix files
func (Proto) Fix(ctx context.Context) error {
	mg.CtxDeps(ctx, proto.Fix)
	return nil
}
