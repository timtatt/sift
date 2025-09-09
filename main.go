package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"slices"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type TestOutputLine struct {
	Time    time.Time `json:"time"`
	Action  string    `json:"action"`
	Package string    `json:"package"`
	Test    string    `json:"test,omitempty"`
	Elapsed float64   `json:"elapsed,omitempty"`
	Output  string    `json:"output,omitempty"`
}

func main() {

	scanner := bufio.NewScanner(os.Stdin)

	// ctx := context.Background()

	// TODO: setup logger

	m := initialModel()

	p := tea.NewProgram(&m)

	go func() {
		for scanner.Scan() {
			var line TestOutputLine

			err := json.Unmarshal(scanner.Bytes(), &line)
			if err != nil {
				// TODO: write to a temp dir log
				continue
			}

			m.AddTest(line)
			if line.Action != "output" {
				p.Send(TestsUpdatedMsg{})
			}

		}

		if err := scanner.Err(); err != nil {
			// return fmt.Errorf("error reading stdin: %w", err)
			p.Quit()
		}
	}()

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}

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

type model struct {
	tests    []TestNode
	testLogs map[TestReference]string
	testLock sync.Mutex

	cursor    int
	startTime time.Time
}

func initialModel() model {
	return model{
		tests:     make([]TestNode, 0),
		testLogs:  make(map[TestReference]string),
		startTime: time.Now(),
	}
}

func (m *model) AddTest(testOutput TestOutputLine) {
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

		m.tests = append(m.tests, TestNode{
			Ref:     testRef,
			Status:  "run",
			Toggled: false,
		})

		m.testLock.Unlock()
	case "pass", "fail":
		m.testLock.Lock()

		testIdx := slices.IndexFunc(m.tests, func(t TestNode) bool {
			return t.Ref == testRef
		})
		if testIdx > -1 {
			m.tests[testIdx].Status = testOutput.Action
		}

		m.testLock.Unlock()
	}

}

func (m *model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

// TODO: fix accordion update when bottom one is closed
func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// Is it a key press?
	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {

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

func (m *model) View() string {
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
				s += fmt.Sprintf("   %s\n", log)
			}
		}
	}

	// print summary

	s += "\n"

	ps := summary.PackageSummary()

	s += dimmed.Render(" Packages ")
	if ps.Passed > 0 {
		s += greenText.Bold(true).Render(fmt.Sprintf("%d passed ", ps.Passed))
	}

	if ps.Failed > 0 {
		s += redText.Bold(true).Render(fmt.Sprintf("%d failed ", ps.Failed))
	}

	if ps.Running > 0 {
		s += dimmed.Render(fmt.Sprintf("%d running ", ps.Running))
	}
	s += dimmed.Render(fmt.Sprintf("(%d)\n", ps.Passed+ps.Failed+ps.Running))

	s += dimmed.Render("    Tests ")
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

	s += dimmed.Render(fmt.Sprintf("(%d)\n", total.Passed+total.Failed+total.Running))

	s += dimmed.Render(" Start At ")
	s += m.startTime.Format(time.TimeOnly)
	s += "\n"

	s += "\n"

	// Send the UI for rendering
	return s
}
