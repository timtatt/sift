package logparse

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseDefaultLog(t *testing.T) {
	tests := []struct {
		name        string
		log         string
		wantErr     bool
		expectedMsg string
	}{
		{
			name:        "valid default log",
			log:         "2025/10/05 09:52:58 Test message",
			wantErr:     false,
			expectedMsg: "Test message",
		},
		{
			name:    "slog text (should fail)",
			log:     `time=2025-10-05T09:52:58.046+11:00 level=INFO msg="Test"`,
			wantErr: true,
		},
		{
			name:    "JSON log (should fail)",
			log:     `{"time":"2025-10-05T09:52:58.045477+11:00","level":"INFO","msg":"Test"}`,
			wantErr: true,
		},
		{
			name:        "log with multiword message",
			log:         "2025/10/05 09:52:58 This is a formatted log message: formatted value",
			wantErr:     false,
			expectedMsg: "This is a formatted log message: formatted value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry, err := ParseDefaultLog(strings.NewReader(tt.log))

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedMsg, entry.Message)
		})
	}
}

func TestParseDefaultLog_Examples(t *testing.T) {
	tests := []struct {
		name        string
		log         string
		expectedMsg string
	}{
		{
			name:        "standard log message",
			log:         "2025/10/05 09:52:58 This is a standard log message",
			expectedMsg: "This is a standard log message",
		},
		{
			name:        "formatted log message",
			log:         "2025/10/05 09:52:58 This is a formatted log message: formatted value",
			expectedMsg: "This is a formatted log message: formatted value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry, err := ParseDefaultLog(strings.NewReader(tt.log))

			require.NoError(t, err)
			assert.Equal(t, tt.expectedMsg, entry.Message)

			assert.Equal(t, "2025-10-05T09:52:58Z", entry.Time.Format(time.RFC3339))
		})
	}
}
