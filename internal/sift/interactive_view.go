package sift

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/timtatt/sift/internal/tests"
	"github.com/timtatt/sift/pkg/viewbuilder"
)

func (m *siftModel) interactiveView() string {
	s := ""

	var header string
	header += styleHeader.Render("\u2207 sift")

	if m.autoToggleMode {
		header += styleSecondary.Render(" [AUTO TOGGLE MODE]")
	}

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

func formatDuration(d time.Duration) string {
	if d.Milliseconds() < 1000 {
		return fmt.Sprintf("%dms", d.Milliseconds())
	} else if d.Seconds() < 60 {
		return fmt.Sprintf("%.0fs", d.Seconds())
	} else {
		minutes := int(d.Minutes())
		seconds := int(d.Seconds()) % 60
		return fmt.Sprintf("%dm%ds", minutes, seconds)
	}
}

func (m *siftModel) testView() (string, *tests.Summary) {
	vb := viewbuilder.New()

	summary := tests.NewSummary()

	stack := newTestStack()
	var lastPackage string

	for i, test := range m.testManager.GetTests {

		ts, ok := m.testState[test.Ref]
		if !ok {
			ts = &testState{}
			m.testState[test.Ref] = ts
		}

		// if the pkg has a build failure, always show it
		if test.Ref.Test == "" {
			ts.toggled = true
		}

		if !m.isTestVisible(test) {
			continue
		}

		if test.Ref.Package != lastPackage {
			if lastPackage != "" {
				vb.AddLine()
			}

			style := styleSecondary
			prefix := ""

			// if the pkg had a build error, highlight it in red
			if test.Ref.Test == "" {
				style = style.Foreground(colorMutedRed)
				prefix = style.Foreground(colorRed).Render("! ")
			}

			vb.Add(prefix + style.Render(test.Ref.Package))
			vb.AddLine()
			lastPackage = test.Ref.Package
		}

		testHighlighted := m.cursor.test == i

		summary.AddToPackage(test.Ref.Package, test.Status)

		statusIcon := getStatusIcon(test.Status)

		prefixTest := stack.PopUntilPrefix(test.Ref.Test)
		testName, _ := strings.CutPrefix(test.Ref.Test, prefixTest)

		indent := getIndentWithBars(stack.Len())

		if test.Ref.Test != "" {

			if testHighlighted {
				testName = styleHighlighted.Render(testName)
			}

			elapsed := ""
			if test.Status != "run" {
				elapsed = styleSecondary.Render(
					formatDuration(test.Elapsed),
				)
			}

			ts.viewportPos = vb.Lines()

			vb.Add(fmt.Sprintf("%s%s %s %s", indent, statusIcon, testName, elapsed))
			if m.opts.Debug {
				vb.Add(fmt.Sprintf(" [%d]", ts.viewportPos))
			}
			vb.AddLine()
		}

		if ts.toggled {
			logs := m.testManager.GetLogs(test.Ref)

			for logIdx, log := range logs {

				logStyle := lipgloss.NewStyle()
				prefix := "  "
				if testHighlighted && logIdx == m.cursor.log {
					prefix = "> "
					logStyle = lipgloss.NewStyle().Bold(true)
				} else if !testHighlighted {
					logStyle = styleSecondary
				}

				var styledLog string
				if m.opts.PrettifyLogs {
					styledLog = prettifyLogEntry(log, logStyle)
				} else {
					styledLog = logStyle.Render(log.Message)
				}

				styledLog = styleLog.Width(m.viewport.Width - 2).Render(styledLog)

				vb.Add(indent + prefix + styledLog)
				vb.AddLine()
			}
		}

		stack.Push(test.Ref.Test)
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

func getIndentWithBars(indentLevel int) string {
	if indentLevel == 0 {
		return ""
	}

	var indent strings.Builder
	for range indentLevel {
		indent.WriteString(styleSecondary.Render("│ "))
	}

	return indent.String()
}
