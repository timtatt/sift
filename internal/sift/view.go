package sift

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/timtatt/sift/internal/tests"
)

type siftModel struct {
	testManager  *tests.TestManager
	toggledTests map[tests.TestReference]bool

	cursor    int
	startTime time.Time
	endTime   time.Time

	ready     bool
	started   bool
	viewport  viewport.Model
	keyBuffer []string

	help help.Model

	windowSize tea.WindowSizeMsg
}

func NewSiftModel() *siftModel {
	return &siftModel{
		testManager:  tests.NewTestManager(),
		toggledTests: make(map[tests.TestReference]bool),
		help:         help.New(),
	}
}

func (m *siftModel) Init() tea.Cmd {
	// initialise key ring buffer with size 2
	m.keyBuffer = make([]string, 2)
	return nil
}

func (m *siftModel) LastKeys(n int) string {
	if n > len(m.keyBuffer) {
		n = len(m.keyBuffer)
	}

	var s string
	for i := len(m.keyBuffer) - n; i < len(m.keyBuffer); i++ {
		s += m.keyBuffer[i]
	}

	return s
}

func (m *siftModel) BufferKey(msg tea.KeyMsg) {
	// shift ring buffer left
	for i := range len(m.keyBuffer) - 1 {
		m.keyBuffer[i] = m.keyBuffer[i+1]
	}

	// add new key to end of buffer
	m.keyBuffer[len(m.keyBuffer)-1] = msg.String()
}

func (m *siftModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	if !m.started && m.testManager.GetTestCount() > 0 {
		m.started = true
		m.startTime = time.Now()
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowSize = msg
		if !m.ready {
			m.help.Width = msg.Width
			m.viewport = viewport.New(msg.Width, msg.Height)
			m.viewport.KeyMap = keys.viewport
			m.ready = true
		} else {
			m.help.Width = msg.Width
			m.viewport.Width = msg.Width
		}
	case tea.KeyMsg:
		m.BufferKey(msg)

		switch m.LastKeys(2) {
		case "zA":
			// toggle recursively
			parentTest := m.testManager.GetTest(m.cursor)

			newState := !m.toggledTests[parentTest.Ref]
			m.toggledTests[parentTest.Ref] = newState
			for _, test := range m.testManager.GetTests {
				if test.Ref.Package == parentTest.Ref.Package && strings.HasPrefix(test.Ref.Test, parentTest.Ref.Test) {
					m.toggledTests[test.Ref] = newState
				}
			}

		case "zR":
			// expand all
			for _, test := range m.testManager.GetTests {
				m.toggledTests[test.Ref] = true
			}
		case "zM":
			// collapse all
			for _, test := range m.testManager.GetTests {
				m.toggledTests[test.Ref] = false
			}
		case "za":
			// toggle over cursor
			test := m.testManager.GetTest(m.cursor)
			m.toggledTests[test.Ref] = !m.toggledTests[test.Ref]
		case "zo":
			// expand over cursor
			test := m.testManager.GetTest(m.cursor)
			m.toggledTests[test.Ref] = true
		case "zc":
			// collapse over cursor
			test := m.testManager.GetTest(m.cursor)
			m.toggledTests[test.Ref] = false
		}

		// TODO: use keys here
		switch msg.String() {
		// TODO: change this keymap
		case "?":
			m.help.ShowAll = !m.help.ShowAll
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < m.testManager.GetTestCount()-1 {
				m.cursor++
			}
		case "enter", " ":
			test := m.testManager.GetTest(m.cursor)

			if test != nil {
				m.toggledTests[test.Ref] = !m.toggledTests[test.Ref]
			}

		}
	}

	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *siftModel) View() string {
	s := ""

	var header string
	header += styleHeader.Render("\u2207 sift")
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
		footer += lipgloss.NewStyle().PaddingTop(1).PaddingBottom(1).Render(m.help.View(keys))

		testViewHeight := lipgloss.Height(testView)
		maxTestViewHeight := m.windowSize.Height - lipgloss.Height(footer) - lipgloss.Height(header) - 2
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

// TODO: don't like how summary is being handled
func (m *siftModel) testView() (string, *tests.Summary) {
	var s string

	summary := tests.NewSummary()

	for i, test := range m.testManager.GetTests {

		highlighted := m.cursor == i

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

		testName := test.Ref.Test

		if highlighted {
			testName = styleHighlighted.Render(test.Ref.Test)
		}

		elapsed := ""
		if test.Status != "run" {
			elapsed = styleSecondary.Render(
				fmt.Sprintf("(%.2fs)", test.Elapsed.Seconds()),
			)
		}

		// Render the row
		s += fmt.Sprintf("%s %s %s", statusIcon, testName, elapsed)

		s += "\n"

		// print the logs
		if m.toggledTests[test.Ref] {
			log, ok := m.testManager.GetLogs(test.Ref)

			if !highlighted {
				log = styleSecondary.Render(log)
			}

			if ok {
				s += styleLog.Render(log) + "\n"
			}
		}
	}

	return s, summary
}

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

	if total.Running > 0 {
		s += styleSecondary.Render(fmt.Sprintf("%d running ", total.Running))
	}

	s += styleSecondary.Render(fmt.Sprintf("(%d)", total.Passed+total.Failed+total.Running))
	s += "\n"

	s += summaryLabel.Render("Start At")
	s += m.startTime.Format(time.TimeOnly)

	duration := m.endTime.Sub(m.startTime)
	if m.endTime.IsZero() {
		duration = time.Now().Sub(m.startTime)
	}

	s += "\n"

	s += summaryLabel.Render("Duration")
	s += duration.Truncate(time.Millisecond).String()

	return s
}
