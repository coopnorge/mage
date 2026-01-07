// Package golangcilint provides the embedded golangci-lint configuration.
package golangcilint

import (
	_ "embed"
)

//go:embed golangci-lint.yml
var golangCILintCfg string

// Cfg returns the embedded golangci-lint configuration
func Cfg() string {
	return golangCILintCfg
}
