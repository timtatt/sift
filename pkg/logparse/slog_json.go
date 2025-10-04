package logparse

import (
	"encoding/json"
	"fmt"
	"io"
)

func ParseSlogJSON(reader io.Reader) (LogEntry, error) {
	var entry LogEntry

	err := json.NewDecoder(reader).Decode(&entry)
	if err != nil {
		return LogEntry{}, fmt.Errorf("unable to parse as slog json %s", err)
	}

	return entry, nil
}
