package sift

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/timtatt/sift/internal/tests"
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
			continue
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

func Run(ctx context.Context) error {

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	fps := 24

	m := NewSiftModel()
	p := tea.NewProgram(
		m,
		tea.WithFPS(fps),
		tea.WithAltScreen(),
		tea.WithContext(ctx),
	)

	sift := &sift{
		model:   m,
		program: p,
	}

	go func() {
		if err := sift.ScanStdin(); err != nil {
			cancel()
		}
	}()

	go sift.Frame(ctx, fps)

	if _, err := p.Run(); err != nil {
		return err
	}

	return nil
}
