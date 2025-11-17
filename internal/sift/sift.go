package sift

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/timtatt/sift/internal/tests"
	"golang.org/x/sync/errgroup"
)

// IsStdinTerminal checks if stdin is a terminal (no piped input)
func IsStdinTerminal() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}

type FrameMsg struct{}

// sends a msg to bubbletea model on an interval to ensure the view is being updated according to framerate
func FrameTicker(ctx context.Context, program *tea.Program, tps int) {
	tick := time.NewTicker(time.Second / time.Duration(tps))
	defer tick.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			program.Send(FrameMsg{})
		}
	}
}

type ProgramOptions struct {
	Debug          bool
	NonInteractive bool
	PrettifyLogs   bool
}

func initLogging() error {
	// TODO: change the file

	logFile, err := os.OpenFile("/tmp/sift.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)

	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	handler := slog.New(slog.NewTextHandler(logFile, &slog.HandlerOptions{Level: slog.LevelDebug}))

	slog.SetDefault(handler)

	return nil
}

func Run(ctx context.Context, opts ProgramOptions) error {

	if opts.Debug {
		if err := initLogging(); err != nil {
			return err
		}
		slog.DebugContext(ctx, "starting sift", "options", opts)
	}

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	fps := 120

	g, ctx := errgroup.WithContext(ctx)

	testManager := tests.NewTestManager(tests.TestManagerOpts{
		ParseLogs: opts.PrettifyLogs,
	})

	m, err := NewSiftModel(SiftModelOptions{
		ProgramOptions: opts,
		TestManager:    testManager,
	})

	if err != nil {
		return fmt.Errorf("unable to create sift model: %w", err)
	}

	programOpts := []tea.ProgramOption{
		tea.WithFPS(fps),
		tea.WithContext(ctx),
	}

	if !opts.NonInteractive {
		programOpts = append(programOpts, tea.WithAltScreen())
	}

	p := tea.NewProgram(m, programOpts...)

	g.Go(func() error {
		if err := testManager.ScanStdin(ctx, os.Stdin); err != nil {
			return err
		}

		m.endTime = time.Now()

		return nil
	})

	g.Go(func() error {
		if _, err := p.Run(); err != nil {
			return err
		}

		cancel()
		return nil
	})

	g.Go(func() error {
		FrameTicker(ctx, p, fps)

		return nil
	})

	err = g.Wait()
	if err != nil {
		return err
	}

	m.quitting = false
	fmt.Println(m.View())
	return nil
}
