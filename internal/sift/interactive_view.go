package sift

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/timtatt/sift/internal/tests"
	"github.com/timtatt/sift/pkg/logparse"
	"github.com/timtatt/sift/pkg/vviewport"
)

func (m *siftModel) interactiveView() string {
	s := ""

	var header string
	header += styleHeader.Render("\u2207 sift")

	if m.autoToggleMode {
		header += styleSecondary.Render(" [AUTO TOGGLE MODE]")
	}

	header += " " + lipgloss.NewStyle().Foreground(colorMutedBlue).Render(Version)

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
		s += m.compileSpinner.View() + "Compiling"
	}

	if m.started {
		testViewBuilder, summary := m.testView()

		m.viewport.SetContent(testViewBuilder)

		var footer string
		footer += "\n"
		footer += m.summaryView(summary)

		if statusView := m.statusView(summary); statusView != "" {
			footer += "\n\n"
			footer += statusView
		}

		footer += "\n"
		footer += lipgloss.NewStyle().PaddingTop(1).Render(m.help.View(keys))

		// TODO: fix this, the viewport height shouldn't change before it is rendered
		// calculate viewport height
		testViewHeight := testViewBuilder.Lines()
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

func (m *siftModel) testView() (*vviewport.ContentBuilder, *tests.Summary) {
	vb := m.viewport.NewContentBuilder()

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
			lastPackage = test.Ref.Package
		}

		testHighlighted := m.cursor.test == i

		summary.AddToPackage(test.Ref.Package, test.Status)

		statusIcon := m.getStatusIcon(test.Status)

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

			s := fmt.Sprintf("%s%s %s %s", indent, statusIcon, testName, elapsed)
			if m.opts.Debug {
				s += fmt.Sprintf(" [%d]", ts.viewportPos)
			}
			vb.Add(s)
		}

		if ts.toggled {
			logs := m.testManager.GetLogs(test.Ref)

			for logIdx, log := range logs {

				styledLog := m.renderLog(logIdx, log, testHighlighted, indent)

				// recalculate log height for viewport
				// TODO: move this outside the render function
				if len(ts.logHeights) <= logIdx || ts.logHeights[logIdx] == 0 {
					logHeight := lipgloss.Height(styledLog)

					if len(ts.logHeights) <= logIdx {
						ts.logHeights = append(ts.logHeights, logHeight)
					} else {
						ts.logHeights[logIdx] = logHeight
					}
				}

				vb.Add(styledLog)

				// hack to stop rendering logs if we're outside the viewport
				// this doesn't handle logs above the viewport, but it's a start
				// this starts to slow down after we're 300 log lines in above
				if vb.Lines() > m.viewport.YOffset+m.viewport.Height {
					break
				}
			}
		}

		stack.Push(test.Ref.Test)
	}

	return vb, summary
}

func (m *siftModel) renderLog(logIdx int, log logparse.LogEntry, testHighlighted bool, indent string) string {
	selectedLog := testHighlighted && logIdx == m.cursor.log

	logStyle := lipgloss.NewStyle()
	prefix := "  "
	if selectedLog {
		prefix = "> "
		logStyle = lipgloss.NewStyle().Bold(true)
	} else {
		logStyle = styleSecondary
	}

	var styledLog string
	if m.opts.PrettifyLogs {
		styledLog = prettifyLogEntry(log, logStyle)
	} else {
		styledLog = logStyle.Render(log.Message)
	}

	styledLog = indent + prefix + styleLog.Render(styledLog)

	wrapLog := lipgloss.NewStyle().Width(m.viewport.Width).Render(styledLog)

	return wrapLog
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
