package pallets_test

import (
	"context"
	"testing"

	"github.com/coopnorge/mage/targets/pallets"
	"github.com/stretchr/testify/assert"
)

// mg.CtxDeps runs each target once per process and caches the result, so only
// one scenario per target can be exercised in-process. Failure paths are
// covered by the subprocess tests in internal/targets/pallets.
func TestValidate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Chdir("../../internal/targets/pallets/testdata/success")
		assert.NoError(t, pallets.Pallets{}.Validate(context.Background()))
	})
}

func TestChanges(t *testing.T) {
	t.Run("changes", func(t *testing.T) {
		t.Setenv("CHANGED_FILES", ".pallet/gitconfig.yaml")
		assert.NoError(t, pallets.Pallets{}.Changes(context.Background()))
	})
}
