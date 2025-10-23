package logparse

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSlogText(t *testing.T) {
	tests := []struct {
		name        string
		log         string
		wantErr     bool
		expectedMsg string
	}{
		{
			name:        "valid slog text",
			log:         `time=2025-10-05T09:52:58.046+11:00 level=INFO msg="Test message" key=value`,
			wantErr:     false,
			expectedMsg: "Test message",
		},
		{
			name:    "JSON log (should fail)",
			log:     `{"time":"2025-10-05T09:52:58.045477+11:00","level":"INFO","msg":"Test"}`,
			wantErr: true,
		},
		{
			name:        "log with quoted values",
			log:         `time=2025-10-05T09:52:58.046+11:00 level=INFO msg="Message with spaces" key="value with spaces"`,
			wantErr:     false,
			expectedMsg: "Message with spaces",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry, err := ParseSlogText(strings.NewReader(tt.log))

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedMsg, entry.Message)
		})
	}
}

func TestParseSlogText_Examples(t *testing.T) {
	tests := []struct {
		name               string
		log                string
		expectedMsg        string
		expectedLevel      string
		expectedAdditional []LogEntryAdditionalProp
	}{
		{
			name:          "info message with string field",
			log:           `time=2025-10-05T09:52:58.046+11:00 level=INFO msg="This is an info message" key1=value1`,
			expectedMsg:   "This is an info message",
			expectedLevel: "INFO",
			expectedAdditional: []LogEntryAdditionalProp{
				{Key: "key1", Value: "value1"},
			},
		},
		{
			name:          "debug message with int field",
			log:           `time=2025-10-05T09:52:58.046+11:00 level=DEBUG msg="This is a debug message" key3=42`,
			expectedMsg:   "This is a debug message",
			expectedLevel: "DEBUG",
			expectedAdditional: []LogEntryAdditionalProp{
				{Key: "key3", Value: "42"},
			},
		},
		{
			name:          "warning message with float field",
			log:           `time=2025-10-05T09:52:58.046+11:00 level=WARN msg="This is a warning message" key4=3.14`,
			expectedMsg:   "This is a warning message",
			expectedLevel: "WARN",
			expectedAdditional: []LogEntryAdditionalProp{
				{Key: "key4", Value: "3.14"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry, err := ParseSlogText(strings.NewReader(tt.log))

			require.NoError(t, err)
			assert.Equal(t, tt.expectedMsg, entry.Message)
			assert.Equal(t, tt.expectedLevel, entry.Level)
			assert.False(t, entry.Time.IsZero(), "Time should not be zero")
			assert.ElementsMatch(t, tt.expectedAdditional, entry.Additional)
		})
	}
}
