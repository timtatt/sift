package sift

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/timtatt/sift/pkg/logparse"
)

func prettifyLogEntry(entry logparse.LogEntry, baseStyle lipgloss.Style) string {
	timeFormatted := styleSecondary.
		Inherit(baseStyle).
		Render(entry.Time.Format(time.TimeOnly + ".000"))

	additionalFields := ""
	for key, value := range entry.Additional {
		if v, ok := value.(string); ok {
			additionalFields += fmt.Sprintf("%s=%s ", key, v)
		}
	}
	if additionalFields != "" {
		additionalFields = styleSecondary.Inherit(baseStyle).Render(" | " + additionalFields[:len(additionalFields)-1])
	}

	level := ""
	if entry.Level != "" {
		level = fmt.Sprintf(" %-5s", entry.Level)
		level = getLogLevelStyle(entry.Level).
			Inherit(baseStyle).
			Render(level)
	}

	message := baseStyle.Render(entry.Message)

	prettifiedLog := fmt.Sprintf("%s%s %s%s", timeFormatted, level, message, additionalFields)

	return prettifiedLog
}

func getLogLevelStyle(level string) lipgloss.Style {
	switch strings.ToLower(level) {
	case "info":
		return lipgloss.NewStyle()
	case "error":
		return lipgloss.NewStyle().Foreground(colorMutedRed)
	case "warn":
		return lipgloss.NewStyle().Foreground(colorMutedOrange)
	case "debug":
		return lipgloss.NewStyle().Foreground(colorMutedBlue)
	default:
		return lipgloss.NewStyle()
	}
}
