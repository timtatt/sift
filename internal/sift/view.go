package sift

import (
	"fmt"
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
	ready     bool
	viewport  viewport.Model
	keys      []tea.KeyMsg

	help help.Model

	windowSize tea.WindowSizeMsg
}

func NewSiftModel() *siftModel {
	return &siftModel{
		testManager:  tests.NewTestManager(),
		toggledTests: make(map[tests.TestReference]bool),
		help:         help.New(),
		startTime:    time.Now(),
	}
}

type TestsUpdatedMsg struct{}

func (m *siftModel) Init() tea.Cmd {
	return nil
}

func (m *siftModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.windowSize = msg
		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height)
			m.viewport.KeyMap = keys.viewport
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
		}
	case tea.KeyMsg:
		// TODO: use keys here
		switch msg.String() {
		// TODO: change this keymap
		case "?":
			m.help.ShowAll = !m.help.ShowAll
		case "a":
			for _, test := range m.testManager.GetTests {
				m.toggledTests[test.Ref] = true
			}
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

var (
	iconStyle = lipgloss.NewStyle().Bold(true)
	greenText = iconStyle.
			Foreground(lipgloss.Color("28"))

	redText = iconStyle.
		Foreground(lipgloss.Color("161"))

	orangeText = iconStyle.
			Foreground(lipgloss.Color("214"))

	dimmed = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	highlightedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("229")).Background(lipgloss.Color("25"))

	logStyle = lipgloss.NewStyle().PaddingLeft(4)
)

func (m *siftModel) View() string {
	s := ""

	testView, summary := m.testView()

	m.viewport.SetContent(testView)

	var footer string
	footer += m.summaryView(summary)
	footer += "\n"
	footer += lipgloss.NewStyle().Padding(1).Render(m.help.View(keys))

	testViewHeight := lipgloss.Height(testView)
	if testViewHeight < m.windowSize.Height {
		m.viewport.Height = testViewHeight
	} else {
		m.viewport.Height = m.windowSize.Height - lipgloss.Height(footer)
	}

	s += m.viewport.View()

	s += "\n"
	s += footer

	return s
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
		case "run":
			statusIcon = orangeText.Render("\u2022")
		case "fail":
			statusIcon = redText.Render("\u00D7")
		case "pass":
			statusIcon = greenText.Render("\u2713")
		}

		testName := test.Ref.Test

		if highlighted {
			testName = highlightedStyle.Render(test.Ref.Test)
		}

		// Render the row
		s += fmt.Sprintf(" %s %s\n", statusIcon, testName)

		// print the logs
		if m.toggledTests[test.Ref] {
			log, ok := m.testManager.GetLogs(test.Ref)

			if !highlighted {
				log = dimmed.Render(log)
			}

			if ok {
				s += logStyle.Render(log) + "\n"
			}
		}
	}

	return s, summary
}

func (m *siftModel) summaryView(summary *tests.Summary) string {
	var s string

	summaryLabel := dimmed.Width(10).Align(lipgloss.Right).PaddingLeft(1).PaddingRight(1)

	ps := summary.PackageSummary()

	s += summaryLabel.Render("Packages")
	if ps.Passed > 0 {
		s += greenText.Bold(true).Render(fmt.Sprintf("%d passed ", ps.Passed))
	}

	if ps.Failed > 0 {
		s += redText.Bold(true).Render(fmt.Sprintf("%d failed ", ps.Failed))
	}

	if ps.Running > 0 {
		s += dimmed.Render(fmt.Sprintf("%d running ", ps.Running))
	}
	s += dimmed.Render(fmt.Sprintf("(%d)", ps.Passed+ps.Failed+ps.Running))
	s += "\n"

	s += summaryLabel.Render("Tests")
	total := summary.Total()

	if total.Passed > 0 {
		s += greenText.Bold(true).Render(fmt.Sprintf("%d passed ", total.Passed))
	}

	if total.Failed > 0 {
		s += redText.Bold(true).Render(fmt.Sprintf("%d failed ", total.Failed))
	}

	if total.Running > 0 {
		s += dimmed.Render(fmt.Sprintf("%d running ", total.Running))
	}

	s += dimmed.Render(fmt.Sprintf("(%d)", total.Passed+total.Failed+total.Running))
	s += "\n"

	s += summaryLabel.Render("Start At")
	s += m.startTime.Format(time.TimeOnly)

	return s
}
