package samples

import (
	"testing"
)

func TestNestingWithForwardSlashes(t *testing.T) {
	t.Run("this is an api test for /api/v1/health endpoint", func(t *testing.T) {
		t.Log("doing a thing")
		t.Run("do a thing", func(t *testing.T) {
			t.Log("doing a thing")
		})
	})
}
