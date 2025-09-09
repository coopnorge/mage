// Package jslib implements the [mage targets] for working with JS libraries.
//
// To enable the targets in a repository [import] them in
// magefiles/magefile.go
//
//// [mage targets]: https://magefile.org/targets/
// [import]: https://magefile.org/importing/

package jslib

import (
	"context"
	"github.com/magefile/mage/mg"
)

func Lint(ctx context.Context) error {
	mg.CtxDeps(ctx, JS.Lint)
	return nil
}
