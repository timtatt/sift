package logparse

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSlogJSON(t *testing.T) {
	tests := []struct {
		name        string
		log         string
		wantErr     bool
		expectedMsg string
	}{
		{
			name:        "valid JSON log",
			log:         `{"time":"2025-10-05T09:52:58.045477+11:00","level":"INFO","msg":"Test message","key":"value"}`,
			wantErr:     false,
			expectedMsg: "Test message",
		},
		{
			name:    "invalid JSON",
			log:     `not a json log`,
			wantErr: true,
		},
		{
			name:    "empty string",
			log:     ``,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry, err := ParseSlogJSON(strings.NewReader(tt.log))

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedMsg, entry.Message)
		})
	}
}

func TestParseSlogJSON_Examples(t *testing.T) {
	tests := []struct {
		name           string
		log            string
		expectedMsg    string
		expectedLevel  string
		expectedFields map[string]any
	}{
		{
			name:          "info message with string field",
			log:           `{"time":"2025-10-05T09:52:58.045477+11:00","level":"INFO","msg":"This is an info message","key1":"value1"}`,
			expectedMsg:   "This is an info message",
			expectedLevel: "INFO",
			expectedFields: map[string]any{
				"key1": "value1",
			},
		},
		{
			name:          "debug message with int field",
			log:           `{"time":"2025-10-05T09:52:58.045984+11:00","level":"DEBUG","msg":"This is a debug message","key3":42}`,
			expectedMsg:   "This is a debug message",
			expectedLevel: "DEBUG",
			expectedFields: map[string]any{
				"key3": float64(42), // JSON numbers are decoded as float64
			},
		},
		{
			name:          "warning message with float field",
			log:           `{"time":"2025-10-05T09:52:58.046013+11:00","level":"WARN","msg":"This is a warning message","key4":3.14}`,
			expectedMsg:   "This is a warning message",
			expectedLevel: "WARN",
			expectedFields: map[string]any{
				"key4": 3.14,
			},
		},
		{
			name:          "error message with string field",
			log:           `{"time":"2025-10-05T09:52:58.046107+11:00","level":"ERROR","msg":"This is an error message","key5":"value5"}`,
			expectedMsg:   "This is an error message",
			expectedLevel: "ERROR",
			expectedFields: map[string]any{
				"key5": "value5",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry, err := ParseSlogJSON(strings.NewReader(tt.log))

			require.NoError(t, err)
			assert.Equal(t, tt.expectedMsg, entry.Message)
			assert.Equal(t, tt.expectedLevel, entry.Level)
			assert.False(t, entry.Time.IsZero(), "Time should not be zero")

			for key, expectedValue := range tt.expectedFields {
				assert.Contains(t, entry.Additional, key)
				assert.Equal(t, expectedValue, entry.Additional[key])
			}
		})
	}
}
