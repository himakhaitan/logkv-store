package output

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// captureOutput is a helper function that redirects os.Stdout to a buffer,
// executes the provided function 'f', and returns the captured output string.
func captureOutput(f func()) string {
	// Create a pipe to capture stdout
	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}

	// Save the original stdout
	stdout := os.Stdout
	// Restore stdout when done
	defer func() {
		os.Stdout = stdout
	}()

	// Redirect stdout to the pipe writer
	os.Stdout = w

	// Run the function that prints output
	f()

	// Close the writer and read all output from the reader
	w.Close()
	out, err := io.ReadAll(r)
	if err != nil {
		panic(err)
	}

	return string(out)
}

// buildExpectedString generates the fully colored string expected from printMessage
func buildExpectedString(title, colorCode, message string) string {
	// Format: color + bold + [TITLE] + reset + color + message + reset + \n
	return fmt.Sprintf("%s%s[%s]%s %s%s%s\n",
		colorCode, bold, title, reset, colorCode, message, reset,
	)
}

func TestPrintFunctions(t *testing.T) {
	const testMsg = "Test message content"

	tests := []struct {
		name          string
		callFunc      func(msg string)
		expectedTitle string
		expectedColor string
	}{
		{
			name:          "Info",
			callFunc:      Info,
			expectedTitle: "INFO",
			expectedColor: blue,
		},
		{
			name:          "Warn",
			callFunc:      Warn,
			expectedTitle: "WARN",
			expectedColor: yellow,
		},
		{
			name:          "Error",
			callFunc:      Error,
			expectedTitle: "ERROR",
			expectedColor: red,
		},
		{
			name:          "Success",
			callFunc:      Success,
			expectedTitle: "SUCCESS",
			expectedColor: green,
		},
		{
			name:          "Debug",
			callFunc:      Debug,
			expectedTitle: "DEBUG",
			expectedColor: cyan,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			captured := captureOutput(func() {
				tt.callFunc(testMsg)
			})
			expected := buildExpectedString(tt.expectedTitle, tt.expectedColor, testMsg)
			assert.Equal(t, expected, captured, "The output string with color codes should match the expected format.")
			assert.True(t, strings.Contains(captured, testMsg), "The captured output must contain the test message.")
			assert.True(t, strings.Contains(captured, fmt.Sprintf("[%s]", tt.expectedTitle)), "The captured output must contain the correct title tag.")
		})
	}
}

func TestDim(t *testing.T) {
	const testMsg = "Dim message content"
	captured := captureOutput(func() {
		Dim(testMsg)
	})
	expected := fmt.Sprintf("%s%s%s\n", grey, testMsg, reset)
	assert.Equal(t, expected, captured, "Dim output string with color codes should match the expected grey format.")
}
