package logparse

import (
	"encoding/json"
	"io"
	"strconv"
	"strings"
	"time"
)

type LogEntry struct {
	Time       time.Time                `json:"time"`
	Level      string                   `json:"level"`
	Message    string                   `json:"msg"`
	Additional []LogEntryAdditionalProp `json:"-"`
}

type LogEntryAdditionalProp struct {
	Key   string
	Value string
}

func Stringify(v any) (string, bool) {
	switch val := v.(type) {
	case string:
		return val, true
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64), true
	case int:
		return strconv.Itoa(val), true
	case bool:
		return strconv.FormatBool(val), true
	default:
		return "", false
	}

}

// TODO: remove this once json/v2 is GA
func (se *LogEntry) UnmarshalJSON(data []byte) error {
	type Alias LogEntry
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

	for key, value := range rawMap {
		v, ok := Stringify(value)
		if ok {
			se.Additional = append(se.Additional, LogEntryAdditionalProp{
				Key:   key,
				Value: v,
			})
		}
	}

	return nil
}

func ParseLog(log string) LogEntry {

	for _, parser := range []func(io.Reader) (LogEntry, error){
		ParseDefaultLog,
		ParseSlogJSON,
		ParseSlogText,
	} {
		parsedLog, err := parser(strings.NewReader(log))

		if err == nil {
			return parsedLog
		}
	}

	return LogEntry{Message: log}
}
