package slogparse

import (
	"encoding/json"
	"fmt"
	"io"
)

func ParseJSON(reader io.Reader) (SlogEntry, error) {
	var entry SlogEntry

	err := json.NewDecoder(reader).Decode(&entry)
	if err != nil {
		return SlogEntry{}, fmt.Errorf("unable to parse as slog json %s", err)
	}

	return entry, nil
}
