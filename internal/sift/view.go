package sift

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/timtatt/sift/internal/tests"
	"github.com/timtatt/sift/pkg/viewbuilder"
)

type testState struct {
	toggled     bool
	viewportPos int
}

type siftModel struct {
	opts SiftOptions

	testManager *tests.TestManager
	testState   map[tests.TestReference]*testState

	cursor *cursor

	startTime time.Time
	endTime   time.Time

	ready     bool
	started   bool
	viewport  viewport.Model
	keyBuffer []string

	help help.Model

	windowSize tea.WindowSizeMsg

	// search functionality
	searchInput textinput.Model
}

type cursor struct {
	test int // tracks the test currently selected
	log  int // tracks the cursor log line
}

func NewSiftModel(opts SiftOptions) *siftModel {
	ti := textinput.New()
	ti.Placeholder = "search for tests"
	ti.PlaceholderStyle = styleSecondary
	ti.Prompt = "Search: /"
	ti.CharLimit = 100

	return &siftModel{
		opts:        opts,
		testManager: tests.NewTestManager(),
		testState:   make(map[tests.TestReference]*testState),
		help:        help.New(),
		cursor: &cursor{
			test: 0,
			log:  0,
		},
		searchInput: ti,
	}
}

// isTestVisible checks if a test passes the current search filter
func (m *siftModel) isTestVisible(testIndex int) bool {
	test := m.testManager.GetTest(testIndex)
	if test == nil {
		return false
	}

	searchQuery := m.searchInput.Value()
	if searchQuery != "" && !fuzzy.MatchFold(searchQuery, test.Ref.Test) {
		return false
	}

	return true
}

// ensureCursorVisible ensures the cursor is on a visible test
// If the current test is hidden, moves to the nearest visible test
func (m *siftModel) ensureCursorVisible() {
	// If current test is visible, we're good
	if m.isTestVisible(m.cursor.test) {
		return
	}

	// Try to find the next visible test
	for i := m.cursor.test + 1; i < m.testManager.GetTestCount(); i++ {
		if m.isTestVisible(i) {
			m.cursor.test = i
			m.cursor.log = 0
			return
		}
	}

	// If no test found forward, try backward
	for i := m.cursor.test - 1; i >= 0; i-- {
		if m.isTestVisible(i) {
			m.cursor.test = i
			m.cursor.log = 0
			return
		}
	}

	// If no visible tests at all, reset to 0
	m.cursor.test = 0
	m.cursor.log = 0
}

func (m *siftModel) PrevTest() {
	if m.cursor.test > 0 {
		// Find the previous visible test
		for i := m.cursor.test - 1; i >= 0; i-- {
			if m.isTestVisible(i) {
				m.cursor.test = i
				m.cursor.log = 0
				return
			}
		}
	}
}

func (m *siftModel) NextTest() {
	if m.cursor.test < m.testManager.GetTestCount()-1 {
		// Find the next visible test
		for i := m.cursor.test + 1; i < m.testManager.GetTestCount(); i++ {
			if m.isTestVisible(i) {
				m.cursor.test = i
				m.cursor.log = 0
				return
			}
		}
	}
}

func (m *siftModel) CursorDown() {
	test := m.testManager.GetTest(m.cursor.test)

	state := m.testState[test.Ref]

	logCount := 0
	if state.toggled {
		logCount = m.testManager.GetLogCount(test.Ref)
	}

	// check if there are more logs we can highlight.
	if state.toggled && m.cursor.log < logCount-1 {
		m.cursor.log++
		return
	}

	if m.cursor.test == m.testManager.GetTestCount()-1 {
		// this is the last test
		return
	}

	// go to the next visible test
	for i := m.cursor.test + 1; i < m.testManager.GetTestCount(); i++ {
		if m.isTestVisible(i) {
			m.cursor.test = i
			m.cursor.log = 0
			return
		}
	}
}

// determine the cursor position with respect to the viewport
func (m *siftModel) GetCursorPos() int {
	test := m.testManager.GetTest(m.cursor.test)

	if test == nil {
		return -1
	}

	ts, ok := m.testState[test.Ref]
	if !ok {
		return -1
	}

	pos := ts.viewportPos + m.cursor.log

	if ts.toggled {
		// if the test is toggled it has 1 extra line
		pos += 1
	}

	return pos
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

	// go to the previous visible test
	for i := m.cursor.test - 1; i >= 0; i-- {
		if m.isTestVisible(i) {
			m.cursor.test = i

			test := m.testManager.GetTest(m.cursor.test)
			if state := m.testState[test.Ref]; state.toggled {
				// set the log to the last log in previous test
				logCount := m.testManager.GetLogCount(test.Ref)
				m.cursor.log = logCount - 1
			} else {
				m.cursor.log = 0
			}
			return
		}
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

const (
	scrollBuffer = 5
)

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
			m.searchInput.Width = msg.Width
			m.ready = true
		} else {
			m.help.Width = msg.Width
			m.viewport.Width = msg.Width
			m.searchInput.Width = msg.Width
		}
	case tea.KeyMsg:
		// capture the most recent keypress into the ring buffer
		m.BufferKey(msg)

		// Handle search mode
		if m.searchInput.Focused() {
			switch msg.String() {
			case "esc", "ctrl+c":
				// Exit search mode and clear query
				m.searchInput.Blur()
				m.searchInput.SetValue("")
				m.ensureCursorVisible()
			case "enter":
				// Exit search mode but keep the filter
				m.searchInput.Blur()
				m.ensureCursorVisible()
			default:
				// Update the textinput with the key
				var inputCmd tea.Cmd
				m.searchInput, inputCmd = m.searchInput.Update(msg)
				cmds = append(cmds, inputCmd)
				m.ensureCursorVisible()
			}
			// Don't process other keys when in search mode
			return m, tea.Batch(cmds...)
		}

		// Check if we should enter search mode
		if msg.String() == "/" {
			m.searchInput.Focus()
			m.searchInput.SetValue("")
			return m, textinput.Blink
		}

		// TODO: rewrite how the cursor is managed
		// 2. can we get the state to me managed more centrally

		switch m.LastKeys(2) {
		case "zA":
			// toggle recursively
			parentTest := m.testManager.GetTest(m.cursor.test)

			newToggleState := !m.testState[parentTest.Ref].toggled
			m.testState[parentTest.Ref].toggled = newToggleState

			for _, test := range m.testManager.GetTests {
				if test.Ref.Package == parentTest.Ref.Package && strings.HasPrefix(test.Ref.Test, parentTest.Ref.Test) {
					m.testState[test.Ref].toggled = newToggleState
				}
			}

			// if collapsing the tests, set the cursor to the top element
			if !newToggleState {
				m.cursor.log = 0
			}

		case "zR":
			// expand all
			for _, test := range m.testManager.GetTests {
				m.testState[test.Ref].toggled = true
			}
		case "zM":
			// collapse all
			for _, test := range m.testManager.GetTests {
				m.testState[test.Ref].toggled = false
			}
			m.cursor.log = 0
		case "za":
			// toggle over cursor
			test := m.testManager.GetTest(m.cursor.test)
			m.testState[test.Ref].toggled = !m.testState[test.Ref].toggled
		case "zo":
			// expand over cursor
			test := m.testManager.GetTest(m.cursor.test)
			m.testState[test.Ref].toggled = true
		case "zc":
			// collapse over cursor
			test := m.testManager.GetTest(m.cursor.test)
			m.testState[test.Ref].toggled = false

			m.cursor.log = 0
		}

		// TODO: use keys here
		switch msg.String() {
		case "{":
			m.PrevTest()

			// scroll up if selected line is within 'scrollBuffer' of the top
			cursorDelta := m.viewport.YOffset - m.GetCursorPos() + scrollBuffer
			if cursorDelta > 0 {
				m.viewport.ScrollUp(cursorDelta)
			}
		case "}":
			m.NextTest()

			// scroll down if selected line is within 'scrollBuffer' of the bottom
			cursorDelta := m.GetCursorPos() - m.viewport.YOffset - m.viewport.Height + scrollBuffer
			if cursorDelta > 0 {
				m.viewport.ScrollDown(cursorDelta)
			}

		case "?":
			m.help.ShowAll = !m.help.ShowAll
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			// Clear search filter when esc is pressed and not in search mode
			if m.searchInput.Value() != "" {
				m.searchInput.SetValue("")
				m.ensureCursorVisible()
			}
		case "up", "k":
			m.CursorUp()

			// scroll up if selected line is within 'scrollBuffer' of the top
			cursorDelta := m.viewport.YOffset - m.GetCursorPos() + scrollBuffer
			if cursorDelta > 0 {
				m.viewport.ScrollUp(cursorDelta)
			}
		case "down", "j":
			m.CursorDown()

			// scroll down if selected line is within 'scrollBuffer' of the bottom
			cursorDelta := m.GetCursorPos() - m.viewport.YOffset - m.viewport.Height + scrollBuffer
			if cursorDelta > 0 {
				m.viewport.ScrollDown(cursorDelta)
			}
		case "enter", " ":
			test := m.testManager.GetTest(m.cursor.test)

			if test != nil {
				newToggleState := !m.testState[test.Ref].toggled
				m.testState[test.Ref].toggled = newToggleState

				if !newToggleState {
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
	if m.opts.Debug {
		header += fmt.Sprintf(" cursor: [%d, %d] %d | yoffset: %d, bottom %d", m.cursor.test, m.cursor.log, m.GetCursorPos(), m.viewport.YOffset, m.viewport.YOffset+m.viewport.Height)

	}
	// Display search input when in search mode
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

// TODO: don't like how summary is being handled
func (m *siftModel) testView() (string, *tests.Summary) {
	vb := viewbuilder.New()

	summary := tests.NewSummary()

	for i, test := range m.testManager.GetTests {

		ts, ok := m.testState[test.Ref]
		if !ok {
			ts = &testState{}
			m.testState[test.Ref] = ts
		}

		// Filter tests based on search query
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
		indent := strings.Repeat("  ", indentLevel)
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

		// print the logs
		if ts.toggled {
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
