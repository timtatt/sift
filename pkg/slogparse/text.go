package slogparse

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

func ParseText(reader io.Reader) (SlogEntry, error) {

	items, err := scan(reader)

	if err != nil {
		return SlogEntry{}, fmt.Errorf("unable to parse as slog text %s", err)
	}

	entry := SlogEntry{
		Additional: make(map[string]any),
	}

	for _, item := range items {
		switch item.key {
		case "time":
			t, err := time.Parse(time.RFC3339, item.value)
			if err == nil {
				entry.Time = t
			}
		case "level":
			entry.Level = item.value
		case "msg":
			entry.Message = item.value
		default:
			entry.Additional[item.key] = item.value
		}
	}

	return entry, nil
}

type kv struct {
	key   string
	value string
}

func scan(reader io.Reader) ([]kv, error) {
	logBytes, err := io.ReadAll(reader)
	log := string(logBytes)

	if err != nil {
		return nil, fmt.Errorf("unable to read text: %v", err)
	}

	if strings.HasPrefix(log, "{") {
		return nil, fmt.Errorf("not a slog text log")
	}

	var items []kv

	keyNumber := 0
	for len(log) > 0 {
		var err error
		var key, value string
		key, log, err = cutString(log, true)
		if err != nil {
			return nil, fmt.Errorf("%s at key %d: %v", err, keyNumber, log)
		}
		if len(log) <= 1 {
			return nil, fmt.Errorf("unterminated string key %d: %v", keyNumber, log)
		} else if len(key) == 0 {
			return nil, fmt.Errorf("malformed key %d: %v", keyNumber, log)
		}
		value, log, err = cutString(log, false)
		if err != nil {
			return nil, fmt.Errorf("%s at value %d: %v", err, keyNumber, log)
		}
		items = append(items, kv{key: key, value: value})
		keyNumber++
	}
	return items, nil
}

func cutString(s string, key bool) (result, rest string, err error) {
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return "", "", nil
	}
	lookFor := byte(' ')
	if key {
		lookFor = '='
	}

	if s[0] != '"' {
		// Simplest case, is not quoted string.
		spaceIdx := strings.IndexByte(s, lookFor)
		if spaceIdx < 0 {
			// No space found, return entire string.
			return s, "", nil
		}
		return s[:spaceIdx], s[spaceIdx+1:], nil
	} else if len(s) > 1 && s[1] == '"' {
		// Empty string case.
		return "", s[2:], nil
	}

	// Parse quoted string case.
	maybeQuoteIdx := 1
	for {
		off := strings.IndexByte(s[maybeQuoteIdx:], '"')
		if off < 0 {
			return "", "", fmt.Errorf("unterminated quoted string: %v", s)
		}
		maybeQuoteIdx += off
		// We now count the number of backslashes before the quote.
		bsCount := 0
		for ; bsCount < maybeQuoteIdx && s[maybeQuoteIdx-1-bsCount] == '\\'; bsCount++ {
		}

		if bsCount%2 == 0 {
			// If the number of backslashes is even,
			// the quote is not escaped, we may terminate the string here.
			break
		}
		// This quote is escaped, continue searching for the next one.
		maybeQuoteIdx++
	}
	result, err = strconv.Unquote(s[:maybeQuoteIdx+1])
	rest = s[maybeQuoteIdx+1:]
	if len(rest) > 0 {
		rest = s[maybeQuoteIdx+2:]
	}
	return result, rest, err
}
