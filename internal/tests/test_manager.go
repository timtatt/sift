package tests

import (
	"slices"
	"sync"
	"time"
)

type TestManager struct {
	tests    []*TestNode
	testLock sync.RWMutex

	testLogs    map[TestReference]string
	testLogLock sync.RWMutex
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
		tm.testLogLock.Lock()
		defer tm.testLogLock.Unlock()

		_, ok := tm.testLogs[testRef]

		if !ok {
			tm.testLogs[testRef] = testOutput.Output
		} else {
			tm.testLogs[testRef] += testOutput.Output
		}

	case "run":
		tm.testLock.Lock()
		defer tm.testLock.Unlock()

		tm.tests = append(tm.tests, &TestNode{
			Ref:    testRef,
			Status: "run",
		})
	case "pass", "fail":
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

func (tm *TestManager) GetLogs(testRef TestReference) (string, bool) {
	tm.testLogLock.RLock()
	defer tm.testLogLock.RUnlock()

	log, ok := tm.testLogs[testRef]

	return log, ok
}

func NewTestManager() *TestManager {
	return &TestManager{
		tests:    make([]*TestNode, 0),
		testLogs: make(map[TestReference]string),
	}

}
