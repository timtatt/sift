package sift

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
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

type sift struct {
	program *tea.Program
	model   *siftModel
}

// IsStdinTerminal checks if stdin is a terminal (no piped input)
func IsStdinTerminal() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}

func (s *sift) ScanStdin(ctx context.Context) error {

	lines := make(chan []byte)
	errChan := make(chan error)

	// scans in a separate channel to allow context cancellation
	go func() {
		scanner := bufio.NewScanner(os.Stdin)

		for scanner.Scan() {
			// without a copy, the underlying array changes whilst it is being processed by the consumer
			lineCopy := make([]byte, len(scanner.Bytes()))
			copy(lineCopy, scanner.Bytes())
			lines <- lineCopy
		}

		if err := scanner.Err(); err != nil {
			errChan <- fmt.Errorf("failed to scan stdin: %w", err)
		}

		close(lines)
		close(errChan)
	}()

out:
	for {
		select {
		// exit early if context is cancelled
		case <-ctx.Done():
			return nil
		case err := <-errChan:
			return err
		case line, ok := <-lines:

			// channel closed, finished processing
			if !ok {
				break out
			}

			var testOutputLine tests.TestOutputLine

			err := json.Unmarshal(line, &testOutputLine)
			if err != nil {
				slog.ErrorContext(ctx, "unable to parse json input", "err", err)
				return errors.New("unable to parse json input. ensure to use the `-json` flag when running go tests")
			}

			s.model.testManager.AddTestOutput(testOutputLine)
		}
	}

	s.model.endTime = time.Now()

	if s.model.testManager.GetTestCount() == 0 {
		return errors.New("no tests received, ensure to specify a package to run `go test` with")
	}

	return nil
}

type FrameMsg struct{}

// sends a msg to bubbletea model on an interval to ensure the view is being updated according to framerate
func (s *sift) Frame(ctx context.Context, tps int) {
	tick := time.NewTicker(time.Second / time.Duration(tps))
	defer tick.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			s.program.Send(FrameMsg{})
		}
	}
}

type SiftOptions struct {
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

func Run(ctx context.Context, opts SiftOptions) error {

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

	m := NewSiftModel(opts)

	programOpts := []tea.ProgramOption{
		tea.WithFPS(fps),
		tea.WithContext(ctx),
	}

	if !opts.NonInteractive {
		programOpts = append(programOpts, tea.WithAltScreen())
	}

	p := tea.NewProgram(m, programOpts...)

	sift := &sift{
		model:   m,
		program: p,
	}

	g.Go(func() error {
		if err := sift.ScanStdin(ctx); err != nil {
			return err
		}

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
		sift.Frame(ctx, fps)

		return nil
	})

	err := g.Wait()
	if err != nil {
		return err
	}

	m.quitting = false
	fmt.Println(m.View())
	return nil
}
