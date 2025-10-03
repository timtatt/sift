package sift

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/timtatt/sift/internal/tests"
	"github.com/timtatt/sift/pkg/viewbuilder"
)

func (m *siftModel) interactiveView() string {
	s := ""

	var header string
	header += styleHeader.Render("\u2207 sift")
	if m.opts.Debug {
		header += fmt.Sprintf(" cursor: [%d, %d] %d | yoffset: %d, bottom %d", m.cursor.test, m.cursor.log, m.GetCursorPos(), m.viewport.YOffset, m.viewport.YOffset+m.viewport.Height)

	}
	if m.searchInput.Focused() {
		header += "\n\n" + m.searchInput.View()
	} else if m.searchInput.Value() != "" {
		header += "\n\n" + fmt.Sprintf("Search: /%s", m.searchInput.Value()) + styleSecondary.Render(" (esc to clear)")
	}
	header += "\n\n"

	s += header

	if !m.started {
		s += "Waiting for test results..."
	}

	if m.started {
		testView, summary := m.testView()

		m.viewport.SetContent(testView)

		var footer string
		footer += "\n"
		footer += m.summaryView(summary)

		if statusView := m.statusView(summary); statusView != "" {
			footer += "\n\n"
			footer += statusView
		}

		footer += "\n"
		footer += lipgloss.NewStyle().PaddingTop(1).Render(m.help.View(keys))

		testViewHeight := lipgloss.Height(testView)
		maxTestViewHeight := m.windowSize.Height - lipgloss.Height(footer) - lipgloss.Height(header)
		m.viewport.Height = min(testViewHeight, maxTestViewHeight)

		s += m.viewport.View()

		s += footer
	}

	return styleBody.Render(s)
}

func (m *siftModel) statusView(summary *tests.Summary) string {

	total := summary.Total()

	if m.endTime.IsZero() {
		return ""
	} else if total.Failed > 0 {
		return styleOutcomeFail.Render("FAILED")
	}

	return styleOutcomePass.Render("PASSED")
}

func (m *siftModel) testView() (string, *tests.Summary) {
	vb := viewbuilder.New()

	summary := tests.NewSummary()

	for i, test := range m.testManager.GetTests {

		ts, ok := m.testState[test.Ref]
		if !ok {
			ts = &testState{}
			m.testState[test.Ref] = ts
		}

		searchQuery := m.searchInput.Value()
		if searchQuery != "" && !fuzzy.MatchFold(searchQuery, test.Ref.Test) {
			continue
		}

		highlighted := m.cursor.test == i

		var statusIcon string
		summary.AddPackage(test.Ref.Package, test.Status)
		switch test.Status {
		case "skip":
			statusIcon = styleSkip.Render("\u23ED")
		case "run":
			statusIcon = styleProgress.Render("\u2022")
		case "fail":
			statusIcon = styleCross.Render("\u00D7")
		case "pass":
			statusIcon = styleTick.Render("\u2713")
		}

		indentLevel := getIndentLevel(test.Ref.Test)
		indent := getIndentWithLines(indentLevel)
		testName := getDisplayName(test.Ref.Test)

		if highlighted {
			testName = styleHighlighted.Render(testName)
		}

		elapsed := ""
		if test.Status != "run" {
			elapsed = styleSecondary.Render(
				fmt.Sprintf("(%.2fs)", test.Elapsed.Seconds()),
			)
		}

		ts.viewportPos = vb.Lines()

		vb.Add(fmt.Sprintf("%s%s %s %s", indent, statusIcon, testName, elapsed))
		if m.opts.Debug {
			vb.Add(fmt.Sprintf(" [%d]", ts.viewportPos))
		}
		vb.AddLine()

		if ts.toggled {
			logs := m.testManager.GetLogs(test.Ref)

			for logIdx, log := range logs {

				prefix := "  "
				if highlighted && logIdx == m.cursor.log {
					prefix = "> "
					log = lipgloss.NewStyle().Bold(true).Render(log)
				} else if !highlighted {
					log = styleSecondary.Render(log)
				}

				log = styleLog.Width(m.viewport.Width - 2).Render(log)

				vb.Add(indent + prefix + log)
				vb.AddLine()
			}
		}
	}

	return vb.String(), summary
}

func getIndentLevel(testName string) int {
	return strings.Count(testName, "/")
}

func getDisplayName(testName string) string {
	lastSlash := strings.LastIndex(testName, "/")
	if lastSlash == -1 {
		return testName
	}
	return testName[lastSlash+1:]
}

func getIndentWithLines(indentLevel int) string {
	if indentLevel == 0 {
		return ""
	}

	var indent strings.Builder
	for range indentLevel {
		indent.WriteString(styleSecondary.Render("│ "))
	}

	return indent.String()
}
