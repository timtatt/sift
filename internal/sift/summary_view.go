package sift

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/timtatt/sift/internal/tests"
)

func (m *siftModel) summaryView(summary *tests.Summary) string {
	var s string

	summaryLabel := styleSecondary.Width(9).Align(lipgloss.Right).PaddingRight(1)

	ps := summary.PackageSummary()

	s += summaryLabel.Render("Packages")
	if ps.Passed > 0 {
		s += styleTick.Bold(true).Render(fmt.Sprintf("%d passed ", ps.Passed))
	}

	if ps.Failed > 0 {
		s += styleCross.Bold(true).Render(fmt.Sprintf("%d failed ", ps.Failed))
	}

	if ps.Running > 0 {
		s += styleSecondary.Render(fmt.Sprintf("%d running ", ps.Running))
	}
	s += styleSecondary.Render(fmt.Sprintf("(%d)", ps.Passed+ps.Failed+ps.Running))
	s += "\n"

	s += summaryLabel.Render("Tests")
	total := summary.Total()

	if total.Passed > 0 {
		s += styleTick.Bold(true).Render(fmt.Sprintf("%d passed ", total.Passed))
	}

	if total.Failed > 0 {
		s += styleCross.Bold(true).Render(fmt.Sprintf("%d failed ", total.Failed))
	}

	if total.Skipped > 0 {
		s += styleSkip.Bold(true).Render(fmt.Sprintf("%d skipped ", total.Skipped))
	}

	if total.Running > 0 {
		s += styleSecondary.Render(fmt.Sprintf("%d running ", total.Running))
	}

	s += styleSecondary.Render(fmt.Sprintf("(%d)", total.Passed+total.Failed+total.Running))
	s += "\n"

	s += summaryLabel.Render("Start At")
	s += m.startTime.Format(time.TimeOnly)

	duration := m.endTime.Sub(m.startTime)
	if m.endTime.IsZero() {
		duration = time.Since(m.startTime)
	}

	s += "\n"

	s += summaryLabel.Render("Duration")
	s += duration.Truncate(time.Millisecond).String()

	return s
}
