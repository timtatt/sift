package tests

import (
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

// addLogEntry is a helper method to add a log entry to the test logs
// Caller must hold testLogLock
func (tm *TestManager) addLogEntry(testRef TestReference, logEntry logparse.LogEntry) {
	if _, ok := tm.testLogs[testRef]; ok {
		tm.testLogs[testRef] = append(tm.testLogs[testRef], logEntry)
	} else {
		tm.testLogs[testRef] = []logparse.LogEntry{logEntry}
	}
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
	Time       time.Time `json:"time"`
	Action     string    `json:"action"`
	Package    string    `json:"package"`
	ImportPath string    `json:"ImportPath,omitempty"` // Used for build-output and build-fail actions
	Test       string    `json:"test,omitempty"`
	Elapsed    float64   `json:"elapsed,omitempty"`
	Output     string    `json:"output,omitempty"`
}

func (tm *TestManager) AddTestOutput(testOutput TestOutputLine) {
	// For build-output and build-fail, use ImportPath instead of Package
	pkg := testOutput.Package
	if pkg == "" && testOutput.ImportPath != "" {
		pkg = testOutput.ImportPath
	}

	testRef := TestReference{
		Package: pkg,
		Test:    testOutput.Test,
	}

	switch testOutput.Action {
	case "build-output":
		// Handle build errors (compilation failures)
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

		tm.testLogLock.Lock()
		defer tm.testLogLock.Unlock()

		tm.addLogEntry(testRef, logEntry)

	case "build-fail":
		// Create a test node for the build failure if it doesn't exist
		tm.testLock.Lock()
		defer tm.testLock.Unlock()

		testIdx := slices.IndexFunc(tm.tests, func(t *TestNode) bool {
			return t.Ref == testRef
		})
		if testIdx == -1 {
			// Create a new test node for the build failure
			tm.tests = append(tm.tests, &TestNode{
				Ref:    testRef,
				Status: "fail",
			})
		} else {
			// Update existing test node to failed
			tm.tests[testIdx].Status = "fail"
		}

	case "output":
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

		tm.addLogEntry(testRef, logEntry)

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

func (tm *TestManager) GetLogs(testRef TestReference) []logparse.LogEntry {
	tm.testLogLock.RLock()
	defer tm.testLogLock.RUnlock()

	if log, ok := tm.testLogs[testRef]; ok {
		return log
	}

	return nil
}
