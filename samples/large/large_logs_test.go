package large

import (
	"strings"
	"testing"
)

func TestLargeLogs(t *testing.T) {
	for range 3 {
		t.Run("long test", func(t *testing.T) {
			for i := range 2000 {
				extra := strings.Repeat("ab ", 80)
				t.Logf("This is log message number %d - %s", i+1, extra)
			}
		})
	}
}
