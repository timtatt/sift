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
	"github.com/timtatt/sift/pkg/viewbuilder"
)

type siftModel struct {
	testManager  *tests.TestManager
	toggledTests map[tests.TestReference]bool

	cursor *cursor

	startTime time.Time
	endTime   time.Time

	ready     bool
	started   bool
	viewport  viewport.Model
	keyBuffer []string

	help help.Model

	windowSize tea.WindowSizeMsg
}

type cursor struct {
	test int // tracks the test currently selected
	log  int // tracks the cursor log line
}

func NewSiftModel() *siftModel {
	return &siftModel{
		testManager:  tests.NewTestManager(),
		toggledTests: make(map[tests.TestReference]bool),
		help:         help.New(),
		cursor: &cursor{
			test: 0,
			log:  0,
		},
	}
}

func (m *siftModel) PrevTest() {
	if m.cursor.test > 0 {
		m.cursor.test--
		m.cursor.log = 0
	}
}

func (m *siftModel) NextTest() {
	if m.cursor.test < m.testManager.GetTestCount()-1 {
		m.cursor.test++
		m.cursor.log = 0
	}
}

func (m *siftModel) CursorDown() {
	test := m.testManager.GetTest(m.cursor.test)

	toggled := m.toggledTests[test.Ref]

	logCount := 0
	if toggled {
		logCount = m.testManager.GetLogCount(test.Ref)
	}

	// check if there are more logs we can highlight.
	if toggled && m.cursor.log < logCount-1 {
		m.cursor.log++
		return
	}

	if m.cursor.test == m.testManager.GetTestCount()-1 {
		// this is the last test
		return
	}

	// go to the next test
	m.cursor.test++
	m.cursor.log = 0
}

func (m *siftModel) CursorUp() {
	if m.cursor.log > 0 {
		m.cursor.log--
		return
	}

	if m.cursor.test == 0 {
		// this is the first test
		return
	}

	// go to the next test
	m.cursor.test--

	test := m.testManager.GetTest(m.cursor.test)

	if m.toggledTests[test.Ref] {
		// set the log to the last log in previous test
		logCount := m.testManager.GetLogCount(test.Ref)
		m.cursor.log = logCount - 1
	} else {
		m.cursor.log = 0
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
		// capture the most recent keypress into the ring buffer
		m.BufferKey(msg)

		// TODO: rewrite how the cursor is managed
		// 1. do we want the cursor to count for the empty lines inbetween tests
		// 2. can we get the state to me managed more centrally

		switch m.LastKeys(2) {
		case "zA":
			// toggle recursively
			parentTest := m.testManager.GetTest(m.cursor.test)

			newState := !m.toggledTests[parentTest.Ref]
			m.toggledTests[parentTest.Ref] = newState
			for _, test := range m.testManager.GetTests {
				if test.Ref.Package == parentTest.Ref.Package && strings.HasPrefix(test.Ref.Test, parentTest.Ref.Test) {
					m.toggledTests[test.Ref] = newState
				}
			}

			// if collapsing the tests, set the cursor to the top element
			if !newState {
				m.cursor.log = 0
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
			m.cursor.log = 0
		case "za":
			// toggle over cursor
			test := m.testManager.GetTest(m.cursor.test)
			m.toggledTests[test.Ref] = !m.toggledTests[test.Ref]
		case "zo":
			// expand over cursor
			test := m.testManager.GetTest(m.cursor.test)
			m.toggledTests[test.Ref] = true
		case "zc":
			// collapse over cursor
			test := m.testManager.GetTest(m.cursor.test)
			m.toggledTests[test.Ref] = false

			m.cursor.log = 0
		}

		// TODO: use keys here
		switch msg.String() {
		case "{":
			m.PrevTest()
		case "}":
			m.NextTest()
		// TODO: change this keymap
		case "?":
			m.help.ShowAll = !m.help.ShowAll
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			m.CursorUp()
		case "down", "j":
			m.CursorDown()
		case "enter", " ":
			test := m.testManager.GetTest(m.cursor.test)

			if test != nil {
				newState := !m.toggledTests[test.Ref]
				m.toggledTests[test.Ref] = newState

				if !newState {
					m.cursor.log = 0
				}
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
	header += fmt.Sprintf("[%d, %d]", m.cursor.test, m.cursor.log)
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

// TODO: don't like how summary is being handled
func (m *siftModel) testView() (string, *tests.Summary) {
	vb := viewbuilder.New()

	summary := tests.NewSummary()

	for i, test := range m.testManager.GetTests {

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

		testName := test.Ref.Test

		if highlighted {
			testName = styleHighlighted.Render(testName)
		}

		elapsed := ""
		if test.Status != "run" {
			elapsed = styleSecondary.Render(
				fmt.Sprintf("(%.2fs)", test.Elapsed.Seconds()),
			)
		}

		// Render the row
		vb.Add(fmt.Sprintf("%s %s %s", statusIcon, testName, elapsed))
		vb.AddLine()

		// print the logs
		if m.toggledTests[test.Ref] {
			logs := m.testManager.GetLogs(test.Ref)

			for logIdx, log := range logs {

				if highlighted && logIdx == m.cursor.log {
					log = styleHighlightedLog.Render(log)
				} else if !highlighted {
					log = styleSecondary.Render(log)
				}

				log = styleLog.Width(m.viewport.Width - 2).Render(log)

				vb.Add(log)
				vb.AddLine()
			}
			vb.AddLine()
		}
	}

	return vb.String(), summary
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
