package golang

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeFile creates a file at path (relative to dir) with the given content,
// creating intermediate directories as needed.
func writeFile(t *testing.T, dir, relPath, content string) {
	t.Helper()
	full := filepath.Join(dir, relPath)
	require.NoError(t, os.MkdirAll(filepath.Dir(full), 0o755))
	require.NoError(t, os.WriteFile(full, []byte(content), 0o644))
}

// readFile reads a file at path (relative to dir) and returns its content.
func readFile(t *testing.T, dir, relPath string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(dir, relPath))
	require.NoError(t, err)
	return string(content)
}

// newGoModule creates a temporary directory with a minimal Go module and
// returns the path to it. The caller owns cleanup via t.Cleanup (t.TempDir).
func newGoModule(t *testing.T, moduleName, goVersion string, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()

	writeFile(t, dir, "go.mod", "module "+moduleName+"\n\ngo "+goVersion+"\n")

	for relPath, content := range files {
		writeFile(t, dir, relPath, content)
	}
	return dir
}

func TestFix(t *testing.T) {
	tests := []struct {
		name    string
		files   map[string]string
		wantErr bool
	}{
		{
			name: "valid module with no fixable code succeeds",
			files: map[string]string{
				"main.go": `package main

func main() {}
`,
			},
			wantErr: false,
		},
		{
			name: "valid module with multiple packages succeeds",
			files: map[string]string{
				"main.go": `package main

import "example.com/fix/lib"

func main() { lib.Hello() }
`,
				"lib/lib.go": `package lib

import "fmt"

func Hello() { fmt.Println("hello") }
`,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := newGoModule(t, "example.com/fix", "1.26.0", tt.files)
			err := Fix(dir)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFix_nonExistentDirectory(t *testing.T) {
	err := Fix("/this/path/does/not/exist")
	assert.Error(t, err)
}

// TestFix_any verifies that the "any" fixer replaces interface{} with any.
//
// The any alias was introduced in Go 1.18. go fix rewrites interface{}
// to any as a purely stylistic modernisation.
func TestFix_any(t *testing.T) {
	before := `package main

func accept(v interface{}) {}

func main() {
	var m map[string]interface{}
	_ = m
}
`
	dir := newGoModule(t, "example.com/fix", "1.26.0", map[string]string{
		"main.go": before,
	})

	require.NoError(t, Fix(dir))

	after := readFile(t, dir, "main.go")
	assert.NotEqual(t, before, after)
	assert.NotContains(t, after, "interface{}", "expected interface{} to be replaced with any")
	assert.Contains(t, after, "any")
}

// TestFix_rangeint verifies that the "rangeint" fixer replaces a traditional
// 3-clause for loop with a range-over-integer loop.
//
// Range-over-integer was introduced in Go 1.22. go fix replaces:
//
//	for i := 0; i < n; i++ { ... }
//
// with:
//
//	for i := range n { ... }
func TestFix_rangeint(t *testing.T) {
	before := `package main

func f(n int) {
	for i := 0; i < n; i++ {
		_ = i
	}
}
`
	dir := newGoModule(t, "example.com/fix", "1.26.0", map[string]string{
		"main.go": before,
	})

	require.NoError(t, Fix(dir))

	after := readFile(t, dir, "main.go")
	assert.NotEqual(t, before, after)
	assert.NotContains(t, after, "i < n", "expected 3-clause for loop to be replaced with range-over-int")
	assert.Contains(t, after, "range n")
}

// TestFix_minmax verifies that the "minmax" fixer replaces an if/else
// conditional assignment with a call to the built-in min or max function.
//
// min and max were introduced in Go 1.21. go fix replaces:
//
//	if a < b { x = a } else { x = b }
//
// with:
//
//	x = min(a, b)
func TestFix_minmax(t *testing.T) {
	before := `package main

func smaller(a, b int) int {
	var x int
	if a < b {
		x = a
	} else {
		x = b
	}
	return x
}
`
	dir := newGoModule(t, "example.com/fix", "1.26.0", map[string]string{
		"main.go": before,
	})

	require.NoError(t, Fix(dir))

	after := readFile(t, dir, "main.go")
	assert.NotEqual(t, before, after)
	assert.NotContains(t, after, "if a < b", "expected if/else to be replaced with min()")
	assert.Contains(t, after, "min(")
}
