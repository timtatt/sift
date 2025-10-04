package prettylogs

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/timtatt/sift/pkg/slogparse"
)

func PrettifyLog(log string) string {

	for _, parser := range []func(io.Reader) (string, error){
		prettifyDefaultLog,
		prettifySlogJSON,
		prettifySlogText,
	} {
		prettifiedLog, err := parser(strings.NewReader(log))

		if err == nil {
			return prettifiedLog
		}
	}

	return log
}

func prettifyDefaultLog(reader io.Reader) (string, error) {

	// check if the first 20 characters match the default log format
	first, err := io.ReadAll(io.LimitReader(reader, 20))
	if err != nil {
		return "", fmt.Errorf("unable to read log %s", err)
	}

	// exit early if it looks like a slog log or doesn't start with a timestamp
	if strings.HasPrefix(string(first), "time=") || strings.HasPrefix(string(first), "{") {
		return "", fmt.Errorf("not a default log")
	}

	timestamp := string(first)
	t, err := time.Parse("2006/01/02 15:04:05", strings.TrimSpace(timestamp))
	if err != nil {
		return "", fmt.Errorf("unable to parse as default log %s", err)
	}

	message, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("unable to read log message %s", err)
	}

	entry := slogparse.SlogEntry{
		Time:    t,
		Message: strings.TrimSpace(string(message)),
	}

	return prettifySlogEntry(entry)
}

func prettifySlogText(reader io.Reader) (string, error) {
	entry, err := slogparse.ParseText(reader)

	if err != nil {
		return "", fmt.Errorf("unable to parse as slog text %s", err)
	}

	return prettifySlogEntry(entry)
}

func prettifySlogJSON(reader io.Reader) (string, error) {
	entry, err := slogparse.ParseJSON(reader)

	if err != nil {
		return "", fmt.Errorf("unable to parse as slog json %s", err)
	}

	return prettifySlogEntry(entry)
}

func prettifySlogEntry(entry slogparse.SlogEntry) (string, error) {
	timeFormatted := entry.Time.Format(time.TimeOnly)

	additionalFields := ""
	for key, value := range entry.Additional {
		if v, ok := value.(string); ok {
			additionalFields += fmt.Sprintf("%s=%s ", key, v)
		}
	}
	if additionalFields != "" {
		additionalFields = " | " + additionalFields[:len(additionalFields)-1] // Remove trailing space
	}

	level := ""
	if entry.Level != "" {
		level = fmt.Sprintf(" %s", entry.Level)
	}

	prettifiedLog := fmt.Sprintf("%s%s %s%s", timeFormatted, level, entry.Message, additionalFields)

	return prettifiedLog, nil
}
