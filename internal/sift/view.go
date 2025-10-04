package sift

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/timtatt/sift/internal/tests"
	"github.com/timtatt/sift/pkg/helpview"
)

type testState struct {
	toggled     bool
	viewportPos int
}

type viewMode int

const (
	viewModeAlternate viewMode = iota
	viewModeInline
)

type siftModel struct {
	opts SiftOptions

	testManager *tests.TestManager
	testState   map[tests.TestReference]*testState

	cursor *cursor

	autoToggleMode bool

	startTime time.Time
	endTime   time.Time

	ready     bool
	started   bool
	viewport  viewport.Model
	keyBuffer []string

	help *helpview.WrappingHelpView

	windowSize tea.WindowSizeMsg

	searchInput textinput.Model

	mode viewMode
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

	mode := viewModeAlternate
	if opts.NonInteractive {
		mode = viewModeInline
	}

	return &siftModel{
		opts: opts,
		testManager: tests.NewTestManager(tests.TestManagerOpts{
			ParseLogs: opts.PrettifyLogs,
		}),
		testState:      make(map[tests.TestReference]*testState),
		autoToggleMode: true,
		help:           helpview.New(),
		cursor: &cursor{
			test: 0,
			log:  0,
		},
		searchInput: ti,
		mode:        mode,
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
	if m.cursor.test <= 0 {
		return
	}

	// Find the previous visible test
	for i := m.cursor.test - 1; i >= 0; i-- {
		if !m.isTestVisible(i) {
			continue
		}

		if m.autoToggleMode {
			// close the current test if it's open
			m.ToggleTest(m.cursor.test, false)

			// open the next text
			m.ToggleTest(i, true)
		}

		m.cursor.test = i
		m.cursor.log = 0
		return
	}
}

func (m *siftModel) ToggleTest(index int, toggled bool) {
	test := m.testManager.GetTest(index)
	if test != nil {
		m.testState[test.Ref].toggled = toggled
	}
}

func (m *siftModel) NextTest() {
	if m.cursor.test >= m.testManager.GetTestCount() {
		return
	}

	// Find the next visible test
	for i := m.cursor.test + 1; i < m.testManager.GetTestCount(); i++ {
		if !m.isTestVisible(i) {
			continue
		}

		if m.autoToggleMode {
			// close the current test if it's open
			m.ToggleTest(m.cursor.test, false)

			// open the next text
			m.ToggleTest(i, true)
		}

		m.cursor.test = i
		m.cursor.log = 0
		return
	}
}

func (m *siftModel) PrevFailingTest() {
	if m.cursor.test <= 0 {
		return
	}

	// Find the previous visible failing test
	for i := m.cursor.test - 1; i >= 0; i-- {
		if !m.isTestVisible(i) {
			continue
		}

		test := m.testManager.GetTest(i)
		if test != nil && test.Status != "fail" {
			continue
		}

		if m.autoToggleMode {
			// close the current test if it's open
			m.ToggleTest(m.cursor.test, false)

			// open the next text
			m.ToggleTest(i, true)
		}

		m.cursor.test = i
		m.cursor.log = 0
		return
	}
}

func (m *siftModel) NextFailingTest() {
	if m.cursor.test >= m.testManager.GetTestCount() {
		return
	}

	// Find the next visible failing test
	for i := m.cursor.test + 1; i < m.testManager.GetTestCount(); i++ {
		if !m.isTestVisible(i) {
			continue
		}

		test := m.testManager.GetTest(i)
		if test != nil && test.Status != "fail" {
			continue
		}

		if m.autoToggleMode {
			// close the current test if it's open
			m.ToggleTest(m.cursor.test, false)

			// open the next text
			m.ToggleTest(i, true)
		}

		m.cursor.test = i
		m.cursor.log = 0
		return
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
		if !m.isTestVisible(i) {
			continue
		}

		if m.autoToggleMode {
			// close the current test if it's open
			m.ToggleTest(m.cursor.test, false)

			// open the next text
			m.ToggleTest(i, true)
		}

		m.cursor.test = i
		m.cursor.log = 0
		return
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
		if !m.isTestVisible(i) {
			continue
		}

		if m.autoToggleMode {
			// close the current test
			m.ToggleTest(m.cursor.test, false)

			// open the next text
			m.ToggleTest(i, true)
		}

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

func (m *siftModel) Init() tea.Cmd {
	// initialise key ring buffer with size 2
	m.keyBuffer = make([]string, 2)
	return nil
}

func (m *siftModel) LastKeysMatch(binding key.Binding) bool {

	for _, key := range binding.Keys() {
		n := len(key)

		if n > len(m.keyBuffer) {
			continue
		}

		lastNKeys := strings.Join(m.keyBuffer[len(m.keyBuffer)-n:], "")

		if lastNKeys == key {
			return true
		}
	}

	return false
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

	if m.mode == viewModeInline && !m.endTime.IsZero() {
		return m, tea.Quit
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowSize = msg
		if !m.ready {
			m.help.Width = msg.Width
			m.help.ColumnWidth = 20
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
		if m.mode == viewModeInline {
			return m, nil
		}

		m.BufferKey(msg)

		if m.searchInput.Focused() {
			switch {
			case key.Matches(msg, keys.Quit):
				// Exit search mode and clear query
				m.searchInput.Blur()
				m.searchInput.SetValue("")
				m.ensureCursorVisible()
			case msg.String() == "enter":
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
			return m, tea.Batch(cmds...)
		}

		if key.Matches(msg, keys.Search) {
			m.searchInput.Focus()
			m.searchInput.SetValue("")
			return m, textinput.Blink
		}

		switch {
		case m.LastKeysMatch(keys.ToggleTestsRecursively):
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

		case m.LastKeysMatch(keys.ExpandAllTests):
			// expand all
			for _, test := range m.testManager.GetTests {
				m.testState[test.Ref].toggled = true
			}
		case m.LastKeysMatch(keys.CollapseAllTests):
			// collapse all
			for _, test := range m.testManager.GetTests {
				m.testState[test.Ref].toggled = false
			}
			m.cursor.log = 0
		case m.LastKeysMatch(keys.ToggleTest):
			// toggle over cursor
			test := m.testManager.GetTest(m.cursor.test)
			m.testState[test.Ref].toggled = !m.testState[test.Ref].toggled
		case m.LastKeysMatch(keys.ExpandTest):
			// expand over cursor
			test := m.testManager.GetTest(m.cursor.test)
			m.testState[test.Ref].toggled = true
		case m.LastKeysMatch(keys.CollapseTest):
			// collapse over cursor
			test := m.testManager.GetTest(m.cursor.test)
			m.testState[test.Ref].toggled = false

			m.cursor.log = 0
		}

		switch {
		case key.Matches(msg, keys.ChangeMode):
			m.autoToggleMode = !m.autoToggleMode

			if m.autoToggleMode {
				// close all tests except the current one
				for i, test := range m.testManager.GetTests {
					if i != m.cursor.test {
						m.testState[test.Ref].toggled = false
					}
				}
			}

		case key.Matches(msg, keys.PrevTest):
			m.PrevTest()

			// scroll up if selected line is within 'scrollBuffer' of the top
			cursorDelta := m.viewport.YOffset - m.GetCursorPos() + scrollBuffer
			if cursorDelta > 0 {
				m.viewport.ScrollUp(cursorDelta)
			}
		case key.Matches(msg, keys.NextTest):
			m.NextTest()

			// scroll down if selected line is within 'scrollBuffer' of the bottom
			cursorDelta := m.GetCursorPos() - m.viewport.YOffset - m.viewport.Height + scrollBuffer
			if cursorDelta > 0 {
				m.viewport.ScrollDown(cursorDelta)
			}
		case key.Matches(msg, keys.PrevFailingTest):
			m.PrevFailingTest()

			// scroll up if selected line is within 'scrollBuffer' of the top
			cursorDelta := m.viewport.YOffset - m.GetCursorPos() + scrollBuffer
			if cursorDelta > 0 {
				m.viewport.ScrollUp(cursorDelta)
			}
		case key.Matches(msg, keys.NextFailingTest):
			m.NextFailingTest()

			// scroll down if selected line is within 'scrollBuffer' of the bottom
			cursorDelta := m.GetCursorPos() - m.viewport.YOffset - m.viewport.Height + scrollBuffer
			if cursorDelta > 0 {
				m.viewport.ScrollDown(cursorDelta)
			}

		case key.Matches(msg, keys.Help):
			m.help.ShowAll = !m.help.ShowAll
		case key.Matches(msg, keys.Quit):
			if m.mode == viewModeAlternate {
				m.mode = viewModeInline
				return m, tea.ExitAltScreen
			}
			if !m.endTime.IsZero() {
				return m, tea.Quit
			}
		case key.Matches(msg, keys.ClearSearch):
			// Clear search filter when esc is pressed and not in search mode
			if m.searchInput.Value() != "" {
				m.searchInput.SetValue("")
				m.ensureCursorVisible()
			}
		case key.Matches(msg, keys.Up):
			m.CursorUp()

			// scroll up if selected line is within 'scrollBuffer' of the top
			cursorDelta := m.viewport.YOffset - m.GetCursorPos() + scrollBuffer
			if cursorDelta > 0 {
				m.viewport.ScrollUp(cursorDelta)
			}
		case key.Matches(msg, keys.Down):
			m.CursorDown()

			// scroll down if selected line is within 'scrollBuffer' of the bottom
			cursorDelta := m.GetCursorPos() - m.viewport.YOffset - m.viewport.Height + scrollBuffer
			if cursorDelta > 0 {
				m.viewport.ScrollDown(cursorDelta)
			}
		case key.Matches(msg, keys.ToggleTestAlt):
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

	if m.mode == viewModeAlternate {
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *siftModel) View() string {
	if m.mode == viewModeInline {
		return m.inlineView()
	}
	return m.interactiveView()
}
