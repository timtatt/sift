package prettylogs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrettyLogs(t *testing.T) {

	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "JSON Slog",
			input:    `{"time":"2025-10-04T09:47:26.67466+10:00","level":"INFO","msg":"This is an info message","key1":"value1"}`,
			expected: `09:47:26 INFO This is an info message | key1=value1`,
		},
		{
			name:     "Standard log",
			input:    `2025/10/04 09:47:26 This is a standard log message`,
			expected: `09:47:26 This is a standard log message`,
		},
		{
			name:     "Default Slog",
			input:    `time=2025-10-04T09:47:46.811+10:00 level=INFO msg="This is an info message" key1=value1`,
			expected: `09:47:46 INFO This is an info message | key1=value1`,
		},
		{
			name:     "Raw log",
			input:    `This is a raw log message`,
			expected: `This is a raw log message`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			output := PrettifyLog(tc.input)
			assert.Equal(t, tc.expected, output)
		})
	}

}
