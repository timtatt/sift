package logs

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"testing"
)

func TestJsonLogger(t *testing.T) {

	ctx := t.Context()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	logger.Info("This is an info message", slog.String("key1", "value1"))
	logger.InfoContext(ctx, "This is an info message with context", slog.String("key2", "value"))

	logger.Debug("This is a debug message", slog.Int("key3", 42))
	logger.DebugContext(ctx, "This is a debug message with context", slog.Int("key3", 42))

	logger.Warn("This is a warning message", slog.Float64("key4", 3.14))
	logger.WarnContext(ctx, "This is a warning message with context", slog.Float64("key4", 3.14))

	logger.Error("This is an error message", slog.String("key5", "value5"))
	logger.ErrorContext(ctx, "This is an error message with context", slog.String("key5", "value5"))
}

func TestStandardLogger(t *testing.T) {
	log.Println("This is a standard log message")
	log.Printf("This is a formatted log message: %s", "formatted value")
}

func TestStandardSlog(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	logger.Info("This is an info message", slog.String("key1", "value1"))
	logger.Debug("This is a debug message", slog.Int("key3", 42))
	logger.Warn("This is a warning message", slog.Float64("key4", 3.14))
	logger.Error("This is an error message", slog.String("key5", "value5"))
}

func TestRawLogging(t *testing.T) {
	fmt.Println("This is a raw log message")
}

func TestSlogWith20Attributes(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	logger.Info("Log entry with 20 attributes",
		slog.String("attr1", "value1"),
		slog.String("attr2", "value2"),
		slog.Int("attr3", 3),
		slog.Int("attr4", 4),
		slog.Float64("attr5", 5.5),
		slog.Float64("attr6", 6.6),
		slog.Bool("attr7", true),
		slog.Bool("attr8", false),
		slog.String("attr9", "value9"),
		slog.String("attr10", "value10"),
		slog.Int("attr11", 11),
		slog.Int("attr12", 12),
		slog.Float64("attr13", 13.13),
		slog.Float64("attr14", 14.14),
		slog.String("attr15", "value15"),
		slog.String("attr16", "value16"),
		slog.Int("attr17", 17),
		slog.Int("attr18", 18),
		slog.String("attr19", "value19"),
		slog.String("attr20", "value20"),
	)
}
