package tests

import (
	"slices"
	"strings"
	"sync"
	"time"
)

type TestManager struct {
	tests    []*TestNode
	testLock sync.RWMutex

	testLogs    map[TestReference][]string
	testLogLock sync.RWMutex
}

func NewTestManager() *TestManager {
	return &TestManager{
		tests:    make([]*TestNode, 0),
		testLogs: make(map[TestReference][]string),
	}
}

type TestReference struct {
	Package string
	Test    string
}

type TestNode struct {
	Ref     TestReference
	Elapsed time.Duration
	Status  string // pass, fail, run
}

// Filters out redundant Go test output lines like "=== RUN" and "--- PASS/FAIL/SKIP"
func shouldSkipLogLine(line string) bool {
	trimmed := strings.TrimLeft(line, " \t")
	return strings.HasPrefix(trimmed, "=== RUN") ||
		strings.HasPrefix(trimmed, "--- PASS:") ||
		strings.HasPrefix(trimmed, "--- FAIL:") ||
		strings.HasPrefix(trimmed, "--- SKIP:")
}

// JSON output from `go test -json`
type TestOutputLine struct {
	Time    time.Time `json:"time"`
	Action  string    `json:"action"`
	Package string    `json:"package"`
	Test    string    `json:"test,omitempty"`
	Elapsed float64   `json:"elapsed,omitempty"`
	Output  string    `json:"output,omitempty"`
}

func (tm *TestManager) AddTestOutput(testOutput TestOutputLine) {
	testRef := TestReference{
		Package: testOutput.Package,
		Test:    testOutput.Test,
	}

	switch testOutput.Action {
	case "output":
		log := strings.TrimRight(testOutput.Output, "\n")

		if shouldSkipLogLine(log) {
			return
		}

		tm.testLogLock.Lock()
		defer tm.testLogLock.Unlock()

		_, ok := tm.testLogs[testRef]

		if ok {
			tm.testLogs[testRef] = append(tm.testLogs[testRef], log)
		} else {
			tm.testLogs[testRef] = []string{log}
		}

	case "run":
		tm.testLock.Lock()
		defer tm.testLock.Unlock()

		tm.tests = append(tm.tests, &TestNode{
			Ref:    testRef,
			Status: "run",
		})
	case "pass", "fail", "skip":
		tm.testLock.Lock()
		defer tm.testLock.Unlock()

		testIdx := slices.IndexFunc(tm.tests, func(t *TestNode) bool {
			return t.Ref == testRef
		})
		if testIdx > -1 {
			test := tm.tests[testIdx]

			test.Status = testOutput.Action
			test.Elapsed = time.Duration(float64(time.Second) * testOutput.Elapsed)
		}
	}
}

func (tm *TestManager) GetTests(yield func(int, *TestNode) bool) {

	tm.testLock.RLock()
	defer tm.testLock.RUnlock()

	for i, test := range tm.tests {
		if !yield(i, test) {
			return
		}
	}
}

func (tm *TestManager) GetTest(index int) *TestNode {
	tm.testLock.RLock()
	defer tm.testLock.RUnlock()

	if index < 0 || index >= len(tm.tests) {
		return nil
	}

	return tm.tests[index]
}

func (tm *TestManager) GetTestCount() int {
	tm.testLock.RLock()
	defer tm.testLock.RUnlock()

	return len(tm.tests)
}

func (tm *TestManager) GetLogCount(testRef TestReference) int {
	tm.testLogLock.RLock()
	defer tm.testLogLock.RUnlock()

	if log, ok := tm.testLogs[testRef]; ok {
		return len(log)
	}

	return 0
}

func (tm *TestManager) GetLogs(testRef TestReference) []string {
	tm.testLogLock.RLock()
	defer tm.testLogLock.RUnlock()

	if log, ok := tm.testLogs[testRef]; ok {
		return log
	}

	return nil
}
