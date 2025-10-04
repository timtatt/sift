package logparse

import (
	"fmt"
	"io"
	"strings"
	"time"
)

func ParseDefaultLog(reader io.Reader) (LogEntry, error) {

	// check if the first 20 characters match the default log format
	first, err := io.ReadAll(io.LimitReader(reader, 20))
	if err != nil {
		return LogEntry{}, fmt.Errorf("unable to read log %s", err)
	}

	// exit early if it looks like a slog log or doesn't start with a timestamp
	if strings.HasPrefix(string(first), "time=") || strings.HasPrefix(string(first), "{") {
		return LogEntry{}, fmt.Errorf("not a default log")
	}

	timestamp := string(first)
	t, err := time.Parse("2006/01/02 15:04:05", strings.TrimSpace(timestamp))
	if err != nil {
		return LogEntry{}, fmt.Errorf("unable to parse as default log %s", err)
	}

	message, err := io.ReadAll(reader)
	if err != nil {
		return LogEntry{}, fmt.Errorf("unable to read log message %s", err)
	}

	return LogEntry{
		Time:    t,
		Message: strings.TrimSpace(string(message)),
	}, nil

}
