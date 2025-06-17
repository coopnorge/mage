package testhelpers

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// Chdir calls os.Chdir(dir) and uses Cleanup to restore the current
// working directory to its original value after the test. On Unix, it
// also sets PWD environment variable for the duration of the test.
//
// Because Chdir affects the whole process, it cannot be used
// in parallel tests or tests with parallel ancestors.
//
// This is duplicated from Go 1.24.2's implementation: https://cs.opensource.google/go/go/+/refs/tags/go1.24.4:src/testing/testing.go;l=1345-1385
// When Go 1.25 is released, we can remove this and use the standard library's implementation instead.
func Chdir(t *testing.T, dir string) {
	t.Helper()

	oldwd, err := os.Open(".")
	if err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	// On POSIX platforms, PWD represents “an absolute pathname of the
	// current working directory.” Since we are changing the working
	// directory, we should also set or update PWD to reflect that.
	switch runtime.GOOS {
	case "windows", "plan9":
		// Windows and Plan 9 do not use the PWD variable.
	default:
		if !filepath.IsAbs(dir) {
			dir, err = os.Getwd()
			if err != nil {
				t.Fatal(err)
			}
		}
		t.Setenv("PWD", dir)
	}
	t.Cleanup(func() {
		err := oldwd.Chdir()
		oldwd.Close() //nolint:errcheck // Copy of stdlibs implementation, which ignores the error here.
		if err != nil {
			// It's not safe to continue with tests if we can't
			// get back to the original working directory. Since
			// we are holding a dirfd, this is highly unlikely.
			panic("testing.Chdir: " + err.Error())
		}
	})
}
