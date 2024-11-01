package util

import (
	"fmt"
	"path/filepath"
	"runtime"
)

func GetProjectRoot() (string, error) {
	// Get the directory of the current file
	_, b, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("could not determine current file path")
	}

	basePath := filepath.Dir(b)

	// Adjust the path to go to the project root. This depends on where this file is located.
	// If `common.go` is in `internal/util`, we need to go up 3 directories to reach the project root.
	projectRoot := filepath.Join(basePath, "../../..")
	absPath, err := filepath.Abs(projectRoot)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	return absPath, nil
}
