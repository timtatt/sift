package sift

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/timtatt/sift/internal/tests"
	"github.com/timtatt/sift/pkg/prettylogs"
	"golang.org/x/sync/errgroup"
)

type sift struct {
	program *tea.Program
	model   *siftModel
}

func (s *sift) ScanStdin() error {
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		var line tests.TestOutputLine

		err := json.Unmarshal(scanner.Bytes(), &line)
		if err != nil {
			// TODO: write to a temp dir log
			return errors.New("unable to parse json input. ensure to use the `-json` flag when running go tests")
		}

		if s.model.opts.PrettifyLogs && line.Output != "" {
			line.Output = prettylogs.PrettifyLog(line.Output)
		}

		s.model.testManager.AddTestOutput(line)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to scan stdin: %w", err)
	}

	s.model.endTime = time.Now()

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

func Run(ctx context.Context, opts SiftOptions) error {

	ctx, cancel := context.WithCancel(ctx)
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
		if err := sift.ScanStdin(); err != nil {
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

	return g.Wait()
}
