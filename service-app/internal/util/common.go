package util

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
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

func WriteToFile(fileName string, message string, path string) {
	f, err := os.OpenFile(path+fileName+".txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	if _, err := f.Write([]byte(message + "\n")); err != nil {
		log.Fatal(err)
	}
}

func LoadXMLFile(t *testing.T, filepath string) string {
	data, err := os.ReadFile(filepath)
	require.NoError(t, err, "failed to read XML file")
	return base64.StdEncoding.EncodeToString(data)
}
