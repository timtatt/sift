package sift

import (
	"fmt"
	"slices"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type siftModel struct {
	tests    []*TestNode
	testLogs map[TestReference]string
	testLock sync.Mutex

	cursor    int
	startTime time.Time
}

type TestOutputLine struct {
	Time    time.Time `json:"time"`
	Action  string    `json:"action"`
	Package string    `json:"package"`
	Test    string    `json:"test,omitempty"`
	Elapsed float64   `json:"elapsed,omitempty"`
	Output  string    `json:"output,omitempty"`
}

type TestsUpdatedMsg struct{}

type TestReference struct {
	Package string
	Test    string
}

type TestNode struct {
	Ref     TestReference
	Status  string // pass, fail, run
	Toggled bool
}

func (m *siftModel) AddTest(testOutput TestOutputLine) {
	testRef := TestReference{
		Package: testOutput.Package,
		Test:    testOutput.Test,
	}

	switch testOutput.Action {
	case "output":
		// TODO: maybe use a different lock
		m.testLock.Lock()

		_, ok := m.testLogs[testRef]

		if !ok {
			m.testLogs[testRef] = testOutput.Output
		} else {
			m.testLogs[testRef] += testOutput.Output
		}

		m.testLock.Unlock()
	case "run":
		m.testLock.Lock()

		m.tests = append(m.tests, &TestNode{
			Ref:     testRef,
			Status:  "run",
			Toggled: false,
		})

		m.testLock.Unlock()
	case "pass", "fail":
		m.testLock.Lock()

		testIdx := slices.IndexFunc(m.tests, func(t *TestNode) bool {
			return t.Ref == testRef
		})
		if testIdx > -1 {
			m.tests[testIdx].Status = testOutput.Action
		}

		m.testLock.Unlock()
	}

}

func (m *siftModel) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m *siftModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {

	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {
		// TODO: change this keymap
		case "a":
			for _, test := range m.tests {
				test.Toggled = true
			}

		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit

		// The "up" and "k" keys move the cursor up
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		// The "down" and "j" keys move the cursor down
		case "down", "j":
			if m.cursor < len(m.tests)-1 {
				m.cursor++
			}

		// The "enter" key and the spacebar (a literal space) toggle
		// the selected state for the item that the cursor is pointing at.
		case "enter", " ":
			m.tests[m.cursor].Toggled = !m.tests[m.cursor].Toggled
		}
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
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

type TestSummary struct {
	Passed  int
	Failed  int
	Running int
}

type Summary struct {
	packages map[string]TestSummary
	total    TestSummary
}

func (s *Summary) AddPackage(pkg string, status string) {
	ps, ok := s.packages[pkg]
	if !ok {
		ps = TestSummary{}
	}

	switch status {
	case "pass":
		s.total.Passed++
		ps.Passed++
	case "fail":
		s.total.Failed++
		ps.Failed++
	case "run":
		s.total.Running++
		ps.Running++
	}

	s.packages[pkg] = ps
}

func (s *Summary) Total() TestSummary {
	return s.total
}

func (s *Summary) PackageSummary() TestSummary {
	ps := TestSummary{}
	for _, p := range s.packages {
		ps.Passed += p.Passed
		ps.Failed += p.Failed
		ps.Running += p.Running
	}
	return ps
}

func (m *siftModel) View() string {
	s := ""

	summary := Summary{
		packages: make(map[string]TestSummary),
	}

	// Iterate over our choices
	for i, test := range m.tests {

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
		if test.Toggled {
			m.testLock.Lock()

			log, ok := m.testLogs[test.Ref]

			m.testLock.Unlock()

			if !highlighted {
				log = dimmed.Render(log)
			}

			if ok {
				s += logStyle.Render(log) + "\n"
			}
		}
	}

	// print summary

	s += "\n"

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
	s += "\n"

	s += "\n"

	// Send the UI for rendering
	return s
}
