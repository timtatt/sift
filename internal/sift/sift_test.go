package sift

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsStdinTerminal(t *testing.T) {
	// This test validates that isStdinTerminal returns a boolean value
	// The actual value depends on how the test is run (piped or not)
	// We just verify it doesn't panic and returns a value
	result := isStdinTerminal()
	assert.IsType(t, true, result, "isStdinTerminal should return a boolean")
}
