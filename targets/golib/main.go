package golib

import (
	"context"
	"os"

	"github.com/magefile/mage/sh"
)

const (
	outputDir = "./var"
)

// Generate files
func Generate(_ context.Context) error {
	return nil
}

// Build all code
func Build(_ context.Context) error {
	return os.MkdirAll(outputDir, 0700)
}

// Validate all code
func Validate(_ context.Context) error {
	return nil
}

// Fix files
func Fix(_ context.Context) error {
	return nil
}

// Clean validate and build output
func Clean(_ context.Context) error {
	return sh.Rm(outputDir)
}
