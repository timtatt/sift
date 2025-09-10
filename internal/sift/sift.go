package sift

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type sift struct {
	program *tea.Program
	model   *siftModel
}

func (s *sift) ScanStdin() error {
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		var line TestOutputLine

		err := json.Unmarshal(scanner.Bytes(), &line)
		if err != nil {
			// TODO: write to a temp dir log
			continue
		}

		s.model.AddTest(line)
		s.program.Send(TestsUpdatedMsg{})
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to scan stdin: %w", err)
	}

	return nil
}

func Run(ctx context.Context) error {

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	m := &siftModel{
		tests:     make([]TestNode, 0),
		testLogs:  make(map[TestReference]string),
		startTime: time.Now(),
	}

	p := tea.NewProgram(
		m,
		tea.WithFPS(24),
		tea.WithAltScreen(),
		tea.WithContext(ctx),
		tea.WithMouseCellMotion(),
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

	if _, err := p.Run(); err != nil {
		return err
	}

	return nil
}
