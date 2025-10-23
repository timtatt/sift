package logparse

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseLog_SlogJSON(t *testing.T) {
	log := `{"time":"2025-10-05T09:52:58.045477+11:00","level":"INFO","msg":"This is an info message","key1":"value1"}`
	entry := ParseLog(log)

	assert.Equal(t, "This is an info message", entry.Message)
	assert.Equal(t, "INFO", entry.Level)
	assert.False(t, entry.Time.IsZero(), "Time should not be zero")
	assert.Contains(t, entry.Additional, LogEntryAdditionalProp{Key: "key1", Value: "value1"})
}

func TestParseLog_StandardLog(t *testing.T) {
	log := "2025/10/05 09:52:58 This is a standard log message"
	entry := ParseLog(log)

	assert.Equal(t, "This is a standard log message", entry.Message)
	assert.False(t, entry.Time.IsZero(), "Time should not be zero")
}

func TestParseLog_SlogText(t *testing.T) {
	log := `time=2025-10-05T09:52:58.046+11:00 level=INFO msg="This is an info message" key1=value1`
	entry := ParseLog(log)

	assert.Equal(t, "This is an info message", entry.Message)
	assert.Equal(t, "INFO", entry.Level)
	assert.False(t, entry.Time.IsZero(), "Time should not be zero")
	assert.Contains(t, entry.Additional, LogEntryAdditionalProp{Key: "key1", Value: "value1"})
}

func TestParseLog_RawLog(t *testing.T) {
	log := "This is a raw log message"
	entry := ParseLog(log)

	assert.Equal(t, "This is a raw log message", entry.Message)
	assert.Empty(t, entry.Level, "Level should be empty for raw logs")
}

func TestLogEntry_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name               string
		jsonStr            string
		wantErr            bool
		expectedMsg        string
		expectedLevel      string
		expectedAdditional []LogEntryAdditionalProp
	}{
		{
			name:               "basic log entry",
			jsonStr:            `{"time":"2025-10-05T09:52:58.045477+11:00","level":"INFO","msg":"Test"}`,
			wantErr:            false,
			expectedMsg:        "Test",
			expectedLevel:      "INFO",
			expectedAdditional: []LogEntryAdditionalProp{},
		},
		{
			name:          "log entry with additional fields",
			jsonStr:       `{"time":"2025-10-05T09:52:58.045477+11:00","level":"ERROR","msg":"Error occurred","error":"something went wrong"}`,
			wantErr:       false,
			expectedMsg:   "Error occurred",
			expectedLevel: "ERROR",
			expectedAdditional: []LogEntryAdditionalProp{
				{Key: "error", Value: "something went wrong"},
			},
		},
		{
			name:    "invalid JSON",
			jsonStr: `{invalid json}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var entry LogEntry
			err := json.Unmarshal([]byte(tt.jsonStr), &entry)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedMsg, entry.Message)
			assert.Equal(t, tt.expectedLevel, entry.Level)
			assert.ElementsMatch(t, tt.expectedAdditional, entry.Additional)
		})
	}
}
