package parser

import (
	"regexp"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
)

func DocumentValidationTestHelper(
	t *testing.T,
	fileName string,
	expectedPatterns []string,
	validator CommonValidator,
) {
	messages := validator.Validate()
	t.Log("Actual messages from validation:")
	for _, msg := range messages {
		t.Log(msg)
	}
	for _, pattern := range expectedPatterns {
		regex, compErr := regexp.Compile(pattern)
		require.NoError(t, compErr, "Failed to compile regex for pattern: %s", pattern)

		found := slices.ContainsFunc(messages, regex.MatchString)

		require.True(t, found, "Expected error message pattern not found: %s", pattern)
	}
}
