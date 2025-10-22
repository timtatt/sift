package sift

import (
	"log/slog"
	"strings"
)

type testStack struct {
	stack       []string
	lastElement int
}

func newTestStack() *testStack {
	return &testStack{
		stack:       make([]string, 10),
		lastElement: -1,
	}
}

func (ts *testStack) Len() int {
	return ts.lastElement + 1
}

func (ts *testStack) Push(testName string) {
	if len(ts.stack) == ts.lastElement+1 {
		ts.stack = append(ts.stack, testName)
	} else {
		ts.stack[ts.lastElement+1] = testName
	}
	ts.lastElement++
}

func (ts *testStack) PopUntilPrefix(testName string) string {
	for ts.lastElement > -1 {
		if strings.HasPrefix(testName, ts.stack[ts.lastElement]+"/") {
			slog.Debug("popUntilPrefix", "testName", testName, "stack", ts.stack[:ts.lastElement+1])
			return ts.stack[ts.lastElement] + "/"
		}
		ts.stack[ts.lastElement] = ""
		ts.lastElement--
	}

	return ""
}
