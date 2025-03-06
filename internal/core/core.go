package core

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
)

// OutputDir is the destination for CI artifacts
const OutputDir = "./var"

// GetRelativeRootPath ...
func GetRelativeRootPath(absRootPath, workDirRel string) (string, error) {
	workDirAbs := path.Join(absRootPath, workDirRel)
	relativeRootPath, err := filepath.Rel(workDirAbs, absRootPath)
	if err != nil {
		return "", err
	}
	return relativeRootPath, nil
}

// WriteTempFile writes the content to a temp file in ./var with a random prefix and the
// provided suffix.
func WriteTempFile(directory, suffix, content string) (*os.File, error) {
	err := os.MkdirAll(directory, 0700)
	if err != nil {
		return nil, err
	}
	file, err := os.CreateTemp(directory, fmt.Sprintf("*-%s", suffix))
	if err != nil {
		return nil, err
	}
	_, err = file.WriteString(content)
	if err != nil {
		return nil, err
	}
	return file, nil
}
