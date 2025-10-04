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
