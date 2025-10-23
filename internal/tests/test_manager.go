package tests

import (
	"cmp"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/timtatt/sift/pkg/logparse"
)

type TestManager struct {
	tests    []*TestNode
	testLock sync.RWMutex

	testLogs    map[TestReference][]logparse.LogEntry
	testLogLock sync.RWMutex

	opts TestManagerOpts
}

type TestManagerOpts struct {
	ParseLogs bool
}

func NewTestManager(opts TestManagerOpts) *TestManager {
	return &TestManager{
		opts:     opts,
		tests:    make([]*TestNode, 0),
		testLogs: make(map[TestReference][]logparse.LogEntry),
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

func CompareTestNode(a, b *TestNode) int {
	if c := cmp.Compare(a.Ref.Package, b.Ref.Package); c != 0 {
		return c
	}
	return cmp.Compare(a.Ref.Test, b.Ref.Test)
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
	Time       time.Time `json:"Time"`
	Action     string    `json:"Action"`
	Package    string    `json:"Package"`
	ImportPath string    `json:"ImportPath,omitempty"`
	Test       string    `json:"Test,omitempty"`
	Elapsed    float64   `json:"Elapsed,omitempty"`
	Output     string    `json:"output,omitempty"`
}

func (tm *TestManager) AddTestOutput(testOutput TestOutputLine) {
	pkg := testOutput.Package
	if pkg == "" && testOutput.ImportPath != "" {
		pkg = testOutput.ImportPath
	}

	testRef := TestReference{
		Package: pkg,
		Test:    testOutput.Test,
	}

	switch testOutput.Action {
	case "output", "build-output":
		log := strings.TrimRight(testOutput.Output, "\n")

		var logEntry logparse.LogEntry
		if tm.opts.ParseLogs {
			logEntry = logparse.ParseLog(log)
		} else {
			logEntry = logparse.LogEntry{
				Message: log,
			}
		}

		// provide a time if one isn't present in the log entry
		if logEntry.Time.IsZero() {
			logEntry.Time = testOutput.Time
		}

		if shouldSkipLogLine(log) {
			return
		}

		tm.testLogLock.Lock()
		defer tm.testLogLock.Unlock()

		_, ok := tm.testLogs[testRef]

		if ok {
			tm.testLogs[testRef] = append(tm.testLogs[testRef], logEntry)
		} else {
			tm.testLogs[testRef] = []logparse.LogEntry{logEntry}
		}

	case "build-fail":
		tm.testLock.Lock()
		defer tm.testLock.Unlock()

		newTest := &TestNode{
			Ref:    testRef,
			Status: "error",
		}

		tm.tests = slices.Insert(tm.tests, 0, newTest)

	case "run":
		tm.testLock.Lock()
		defer tm.testLock.Unlock()

		newTest := &TestNode{
			Ref:    testRef,
			Status: "run",
		}

		insertIdx := len(tm.tests)
		for i, t := range tm.tests {
			if CompareTestNode(newTest, t) < 0 {
				insertIdx = i
				break
			}
		}

		tm.tests = slices.Insert(tm.tests, insertIdx, newTest)
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

func (tm *TestManager) GetLogs(testRef TestReference) []logparse.LogEntry {
	tm.testLogLock.RLock()
	defer tm.testLogLock.RUnlock()

	if log, ok := tm.testLogs[testRef]; ok {
		return log
	}

	return nil
}
