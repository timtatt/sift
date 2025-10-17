package sift

import (
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
	"github.com/timtatt/sift/pkg/logparse"
)

func TestPrettifyLogEntry(t *testing.T) {
	baseTime := time.Date(2024, 1, 1, 12, 30, 45, 123000000, time.UTC)
	baseStyle := lipgloss.NewStyle()

	tests := []struct {
		name     string
		entry    logparse.LogEntry
		validate func(t *testing.T, result string)
	}{
		{
			name: "log entry without additional fields",
			entry: logparse.LogEntry{
				Time:    baseTime,
				Level:   "INFO",
				Message: "test message",
			},
			validate: func(t *testing.T, result string) {
				assert.Contains(t, result, "test message")
				assert.NotContains(t, result, " | ")
			},
		},
		{
			name: "log entry with single additional field",
			entry: logparse.LogEntry{
				Time:    baseTime,
				Level:   "INFO",
				Message: "test message",
				Additional: map[string]any{
					"key1": "value1",
				},
			},
			validate: func(t *testing.T, result string) {
				assert.Contains(t, result, "test message")
				assert.Contains(t, result, " | ")
				assert.Contains(t, result, "key1=value1")
			},
		},
		{
			name: "log entry with multiple additional fields",
			entry: logparse.LogEntry{
				Time:    baseTime,
				Level:   "INFO",
				Message: "test message",
				Additional: map[string]any{
					"key1": "value1",
					"key2": "value2",
					"key3": "value3",
				},
			},
			validate: func(t *testing.T, result string) {
				assert.Contains(t, result, "test message")
				assert.Contains(t, result, " | ")
				assert.Contains(t, result, "key1=value1")
				assert.Contains(t, result, "key2=value2")
				assert.Contains(t, result, "key3=value3")
			},
		},
		{
			name: "log entry with many additional fields",
			entry: logparse.LogEntry{
				Time:    baseTime,
				Level:   "INFO",
				Message: "test message",
				Additional: map[string]any{
					"field1":  "value1",
					"field2":  "value2",
					"field3":  "value3",
					"field4":  "value4",
					"field5":  "value5",
					"field6":  "value6",
					"field7":  "value7",
					"field8":  "value8",
					"field9":  "value9",
					"field10": "value10",
				},
			},
			validate: func(t *testing.T, result string) {
				assert.Contains(t, result, "test message")
				assert.Contains(t, result, " | ")
				// Verify all fields are present
				for i := 1; i <= 10; i++ {
					assert.Contains(t, result, "field")
					assert.Contains(t, result, "value")
				}
			},
		},
		{
			name: "log entry with non-string additional fields",
			entry: logparse.LogEntry{
				Time:    baseTime,
				Level:   "INFO",
				Message: "test message",
				Additional: map[string]any{
					"key1": "value1",
					"key2": 123, // non-string value, should be ignored
				},
			},
			validate: func(t *testing.T, result string) {
				assert.Contains(t, result, "test message")
				assert.Contains(t, result, " | ")
				assert.Contains(t, result, "key1=value1")
				// Non-string values are ignored, so we shouldn't see "key2" either
				assert.NotContains(t, result, "key2")
			},
		},
		{
			name: "log entry with empty additional fields map",
			entry: logparse.LogEntry{
				Time:       baseTime,
				Level:      "INFO",
				Message:    "test message",
				Additional: map[string]any{},
			},
			validate: func(t *testing.T, result string) {
				assert.Contains(t, result, "test message")
				assert.NotContains(t, result, " | ")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := prettifyLogEntry(tt.entry, baseStyle)
			assert.NotEmpty(t, result)
			tt.validate(t, result)
		})
	}
}

func TestPrettifyLogEntryPerformance(t *testing.T) {
	baseTime := time.Date(2024, 1, 1, 12, 30, 45, 123000000, time.UTC)
	baseStyle := lipgloss.NewStyle()

	// Create an entry with many additional fields to test performance
	additional := make(map[string]any)
	for i := 0; i < 50; i++ {
		additional["field"+string(rune('A'+i%26))+string(rune('0'+i/26))] = "value" + string(rune('0'+i%10))
	}

	entry := logparse.LogEntry{
		Time:       baseTime,
		Level:      "INFO",
		Message:    "test message with many fields",
		Additional: additional,
	}

	// Run the function multiple times to ensure it doesn't panic or hang
	for i := 0; i < 1000; i++ {
		result := prettifyLogEntry(entry, baseStyle)
		assert.NotEmpty(t, result)
		assert.Contains(t, result, "test message with many fields")
		assert.Contains(t, result, " | ")
	}
}

// BenchmarkPrettifyLogEntry benchmarks the performance of prettifyLogEntry
func BenchmarkPrettifyLogEntry(b *testing.B) {
	baseTime := time.Date(2024, 1, 1, 12, 30, 45, 123000000, time.UTC)
	baseStyle := lipgloss.NewStyle()

	tests := []struct {
		name       string
		fieldCount int
	}{
		{"no_fields", 0},
		{"few_fields", 3},
		{"many_fields", 20},
		{"very_many_fields", 50},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			additional := make(map[string]any)
			for i := 0; i < tt.fieldCount; i++ {
				additional["field"+string(rune('A'+i%26))+string(rune('0'+i/26))] = "value" + string(rune('0'+i%10))
			}

			entry := logparse.LogEntry{
				Time:       baseTime,
				Level:      "INFO",
				Message:    "test message",
				Additional: additional,
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = prettifyLogEntry(entry, baseStyle)
			}
		})
	}
}
