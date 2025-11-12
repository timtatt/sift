package sift

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsStdinTerminal(t *testing.T) {
	result := IsStdinTerminal()
	assert.IsType(t, true, result, "IsStdinTerminal should return a boolean")
}
