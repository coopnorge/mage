package policybotconfig_test

import (
	"context"
	"testing"

	"github.com/coopnorge/mage/targets/policybotconfig"
	"github.com/stretchr/testify/assert"
)

// mg.CtxDeps runs each target once per process and caches the result, so only
// one scenario per target can be exercised in-process. Failure paths are
// covered by the subprocess tests in internal/targets/policybot.
func TestValidate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Chdir("../../internal/targets/policybot/testdata/success")
		assert.NoError(t, policybotconfig.PolicyBotConfig{}.Validate(context.Background()))
	})
}

func TestChanges(t *testing.T) {
	t.Run("changes", func(t *testing.T) {
		t.Setenv("CHANGED_FILES", ".policy.yml")
		assert.NoError(t, policybotconfig.PolicyBotConfig{}.Changes(context.Background()))
	})
}
