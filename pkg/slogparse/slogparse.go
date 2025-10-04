package slogparse

import (
	"encoding/json"
	"time"
)

type SlogEntry struct {
	Time       time.Time      `json:"time"`
	Level      string         `json:"level"`
	Message    string         `json:"msg"`
	Additional map[string]any `json:"-"`
}

// TODO: remove this once json/v2 is GA
func (se *SlogEntry) UnmarshalJSON(data []byte) error {
	type Alias SlogEntry
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(se),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Unmarshal remaining fields into the Additional map
	var rawMap map[string]any
	if err := json.Unmarshal(data, &rawMap); err != nil {
		return err
	}

	delete(rawMap, "time")
	delete(rawMap, "level")
	delete(rawMap, "msg")

	se.Additional = rawMap

	return nil
}
