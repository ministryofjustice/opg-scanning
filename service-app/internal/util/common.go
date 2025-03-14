package util

import (
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
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

func LoadXMLFileTesting(t *testing.T, filepath string) []byte {
	validPath, err := ValidatePath(filepath)
	if err != nil {
		require.FailNow(t, "Invalid file path", err.Error())
	}
	data, err := os.ReadFile(validPath)
	// reading the file.
	if err != nil {
		require.FailNow(t, "Failed to read XML file", err.Error())
	}
	return data
}

func Contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

// Serializes a map of string keys and values of type string or int into a PHP serialized string format.
// It supports only string and integer types for values.
// Only supports flat arrays.
func PhpSerialize(data map[string]interface{}) string {
	var sb strings.Builder
	// Serialize the map as a PHP array
	sb.WriteString("a:" + strconv.Itoa(len(data)) + ":{")

	for key, value := range data {
		// Serialize the key
		sb.WriteString("s:" + strconv.Itoa(len(key)) + `:"` + key + `";`)

		// Serialize the value based on type
		switch v := value.(type) {
		case string:
			sb.WriteString("s:" + strconv.Itoa(len(v)) + `:"` + v + `";`)
		case int:
			sb.WriteString("i:" + strconv.Itoa(v) + ";")
		}
	}

	sb.WriteString("}")
	return sb.String()
}

func IsValidXML(data []byte) error {
	var v interface{}
	if err := xml.Unmarshal(data, &v); err != nil {
		return fmt.Errorf("xml unmarshal error: %w", err)
	}
	return nil
}

func ValidatePath(inputPath string) (string, error) {
	trustedPath, err := GetProjectRoot()
	if err != nil {
		return "", err
	}

	absPath, err := filepath.Abs(inputPath)
	if err != nil {
		return "", err
	}

	// Ensure the file is within the allowed directory
	if !strings.HasPrefix(absPath, trustedPath) {
		return "", errors.New("invalid file path")
	}

	return absPath, nil
}
