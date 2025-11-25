package policybot

import (
	"context"
	"embed"
	"fmt"

	"github.com/coopnorge/mage/internal/devtool"
	"github.com/coopnorge/mage/internal/policybot"
	"github.com/magefile/mage/mg"
)

var (
	//go:embed tools.Dockerfile policy-bot.yml
	// PolicyBotConfigCheckDocker the content of tools.Dockerfile
	PolicyBotConfigCheckDocker embed.FS
)

// Validate validates policybot config file
func Validate(ctx context.Context) error {
	mg.CtxDeps(ctx, runPolicyBotConfigCheck)

	err := policybot.Validate()
	if err != nil {
		return err
	}
	return nil
}

// Changes implements a target that check if the current branch has changes
// related to main branch
func Changes(_ context.Context) error {
	changes, err := policybot.HasChanges()
	if err != nil {
		return err
	}

	if changes {
		fmt.Println("true")
		return nil
	}
	fmt.Println("false")
	return nil
}

func runPolicyBotConfigCheck(_ context.Context) error {
	return devtool.Build("policy-bot", PolicyBotConfigCheckDocker)
}
