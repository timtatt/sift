package sift

import (
	"testing"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/timtatt/sift/internal/tests"
)

func TestGetIndentLevel(t *testing.T) {
	tests := []struct {
		name     string
		testName string
		want     int
	}{
		{
			name:     "no slashes",
			testName: "TestSimple",
			want:     0,
		},
		{
			name:     "one slash",
			testName: "TestParent/TestChild",
			want:     1,
		},
		{
			name:     "two slashes",
			testName: "TestGrandparent/TestParent/TestChild",
			want:     2,
		},
		{
			name:     "three slashes",
			testName: "TestRoot/TestLevel1/TestLevel2/TestLevel3",
			want:     3,
		},
		{
			name:     "empty string",
			testName: "",
			want:     0,
		},
		{
			name:     "trailing slash",
			testName: "TestParent/TestChild/",
			want:     2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getIndentLevel(tt.testName)
			if got != tt.want {
				t.Errorf("getIndentLevel(%q) = %d, want %d", tt.testName, got, tt.want)
			}
		})
	}
}

func TestGetDisplayName(t *testing.T) {
	tests := []struct {
		name     string
		testName string
		want     string
	}{
		{
			name:     "no slashes",
			testName: "TestSimple",
			want:     "TestSimple",
		},
		{
			name:     "one slash",
			testName: "TestParent/TestChild",
			want:     "TestChild",
		},
		{
			name:     "two slashes",
			testName: "TestGrandparent/TestParent/TestChild",
			want:     "TestChild",
		},
		{
			name:     "empty string",
			testName: "",
			want:     "",
		},
		{
			name:     "trailing slash",
			testName: "TestParent/TestChild/",
			want:     "",
		},
		{
			name:     "starts with slash",
			testName: "/TestChild",
			want:     "TestChild",
		},
		{
			name:     "multiple levels deep",
			testName: "A/B/C/D/E/F",
			want:     "F",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getDisplayName(tt.testName)
			if got != tt.want {
				t.Errorf("getDisplayName(%q) = %q, want %q", tt.testName, got, tt.want)
			}
		})
	}
}

func TestGetIndentWithLines(t *testing.T) {
	tests := []struct {
		name        string
		indentLevel int
		wantEmpty   bool
		wantCount   int
	}{
		{
			name:        "zero indent",
			indentLevel: 0,
			wantEmpty:   true,
			wantCount:   0,
		},
		{
			name:        "one indent",
			indentLevel: 1,
			wantEmpty:   false,
			wantCount:   1,
		},
		{
			name:        "two indents",
			indentLevel: 2,
			wantEmpty:   false,
			wantCount:   2,
		},
		{
			name:        "five indents",
			indentLevel: 5,
			wantEmpty:   false,
			wantCount:   5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getIndentWithLines(tt.indentLevel)

			if tt.wantEmpty {
				if got != "" {
					t.Errorf("getIndentWithLines(%d) = %q, want empty string", tt.indentLevel, got)
				}
				return
			}

			if len(got) == 0 {
				t.Errorf("getIndentWithLines(%d) returned empty string, want non-empty", tt.indentLevel)
			}
		})
	}
}

func TestLastKeysMatch(t *testing.T) {
	tests := []struct {
		name      string
		keyBuffer []string
		binding   key.Binding
		want      bool
	}{
		{
			name:      "match two-char sequence zA",
			keyBuffer: []string{"z", "A"},
			binding: key.NewBinding(
				key.WithKeys("zA"),
			),
			want: true,
		},
		{
			name:      "match two-char sequence zR",
			keyBuffer: []string{"z", "R"},
			binding: key.NewBinding(
				key.WithKeys("zR"),
			),
			want: true,
		},
		{
			name:      "no match wrong sequence",
			keyBuffer: []string{"z", "X"},
			binding: key.NewBinding(
				key.WithKeys("zA"),
			),
			want: false,
		},
		{
			name:      "match with history before",
			keyBuffer: []string{"k", "j", "z", "a"},
			binding: key.NewBinding(
				key.WithKeys("za"),
			),
			want: true,
		},
		{
			name:      "match single char at end",
			keyBuffer: []string{"z", "a"},
			binding: key.NewBinding(
				key.WithKeys("a"),
			),
			want: true,
		},
		{
			name:      "no match buffer too short",
			keyBuffer: []string{"z"},
			binding: key.NewBinding(
				key.WithKeys("zA"),
			),
			want: false,
		},
		{
			name:      "match multiple possible keys",
			keyBuffer: []string{"z", "a"},
			binding: key.NewBinding(
				key.WithKeys("za", "enter", " "),
			),
			want: true,
		},
		{
			name:      "empty buffer no match",
			keyBuffer: []string{"", ""},
			binding: key.NewBinding(
				key.WithKeys("zA"),
			),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &siftModel{
				keyBuffer: tt.keyBuffer,
			}
			got := m.LastKeysMatch(tt.binding)
			if got != tt.want {
				t.Errorf("LastKeysMatch() with buffer %v and binding keys %v = %v, want %v",
					tt.keyBuffer, tt.binding.Keys(), got, tt.want)
			}
		})
	}
}

func TestBufferKey(t *testing.T) {
	tests := []struct {
		name           string
		initialBuffer  []string
		keyToAdd       string
		expectedBuffer []string
	}{
		{
			name:           "add to empty buffer",
			initialBuffer:  []string{"", ""},
			keyToAdd:       "a",
			expectedBuffer: []string{"", "a"},
		},
		{
			name:           "add to partially filled buffer",
			initialBuffer:  []string{"", "z"},
			keyToAdd:       "a",
			expectedBuffer: []string{"z", "a"},
		},
		{
			name:           "add to full buffer",
			initialBuffer:  []string{"z", "a"},
			keyToAdd:       "b",
			expectedBuffer: []string{"a", "b"},
		},
		{
			name:           "add multiple times",
			initialBuffer:  []string{"", ""},
			keyToAdd:       "x",
			expectedBuffer: []string{"", "x"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &siftModel{
				keyBuffer: make([]string, len(tt.initialBuffer)),
			}
			copy(m.keyBuffer, tt.initialBuffer)

			msg := tea.KeyMsg{
				Type:  tea.KeyRunes,
				Runes: []rune(tt.keyToAdd),
			}

			m.BufferKey(msg)

			if len(m.keyBuffer) != len(tt.expectedBuffer) {
				t.Errorf("Buffer length = %d, want %d", len(m.keyBuffer), len(tt.expectedBuffer))
				return
			}

			for i := range m.keyBuffer {
				if m.keyBuffer[i] != tt.expectedBuffer[i] {
					t.Errorf("Buffer[%d] = %q, want %q", i, m.keyBuffer[i], tt.expectedBuffer[i])
				}
			}
		})
	}
}

func TestBufferKeySequence(t *testing.T) {
	m := &siftModel{
		keyBuffer: make([]string, 2),
	}

	keys := []struct {
		key      string
		expected []string
	}{
		{"z", []string{"", "z"}},
		{"A", []string{"z", "A"}},
		{"k", []string{"A", "k"}},
	}

	for _, k := range keys {
		msg := tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune(k.key),
		}
		m.BufferKey(msg)

		for i := range m.keyBuffer {
			if m.keyBuffer[i] != k.expected[i] {
				t.Errorf("After adding %q, buffer[%d] = %q, want %q",
					k.key, i, m.keyBuffer[i], k.expected[i])
			}
		}
	}
}

func TestBufferKeyAndLastKeysMatchIntegration(t *testing.T) {
	m := &siftModel{
		keyBuffer: make([]string, 2),
	}

	binding := key.NewBinding(key.WithKeys("zA"))

	if m.LastKeysMatch(binding) {
		t.Error("Should not match with empty buffer")
	}

	m.BufferKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("z")})
	if m.LastKeysMatch(binding) {
		t.Error("Should not match with only 'z' in buffer")
	}

	m.BufferKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("A")})
	if !m.LastKeysMatch(binding) {
		t.Error("Should match with 'zA' in buffer")
	}

	m.BufferKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	if m.LastKeysMatch(binding) {
		t.Error("Should not match after buffer shifted")
	}
}

func TestNextTest(t *testing.T) {
	tests := []struct {
		name           string
		initialCursor  int
		testCount      int
		autoToggleMode bool
		searchFilter   string
		expectedCursor int
		expectedLog    int
	}{
		{
			name:           "move to next test",
			initialCursor:  0,
			testCount:      3,
			autoToggleMode: false,
			expectedCursor: 1,
			expectedLog:    0,
		},
		{
			name:           "move to next test with auto toggle",
			initialCursor:  0,
			testCount:      3,
			autoToggleMode: true,
			expectedCursor: 1,
			expectedLog:    0,
		},
		{
			name:           "at last test should not move",
			initialCursor:  2,
			testCount:      3,
			autoToggleMode: false,
			expectedCursor: 2,
			expectedLog:    0,
		},
		{
			name:           "at boundary condition",
			initialCursor:  1,
			testCount:      2,
			autoToggleMode: false,
			expectedCursor: 1,
			expectedLog:    0,
		},
		{
			name:           "single test should not move",
			initialCursor:  0,
			testCount:      1,
			autoToggleMode: false,
			expectedCursor: 0,
			expectedLog:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := createTestModel(testModelOpts{
				testCount:      tt.testCount,
				autoToggleMode: tt.autoToggleMode,
			})
			m.cursor.test = tt.initialCursor

			m.NextTest()

			if m.cursor.test != tt.expectedCursor {
				t.Errorf("cursor.test = %d, want %d", m.cursor.test, tt.expectedCursor)
			}
			if m.cursor.log != tt.expectedLog {
				t.Errorf("cursor.log = %d, want %d", m.cursor.log, tt.expectedLog)
			}
		})
	}
}

func TestPrevTest(t *testing.T) {
	tests := []struct {
		name           string
		initialCursor  int
		testCount      int
		autoToggleMode bool
		expectedCursor int
		expectedLog    int
	}{
		{
			name:           "move to previous test",
			initialCursor:  2,
			testCount:      3,
			autoToggleMode: false,
			expectedCursor: 1,
			expectedLog:    0,
		},
		{
			name:           "move to previous test with auto toggle",
			initialCursor:  2,
			testCount:      3,
			autoToggleMode: true,
			expectedCursor: 1,
			expectedLog:    0,
		},
		{
			name:           "at first test should not move",
			initialCursor:  0,
			testCount:      3,
			autoToggleMode: false,
			expectedCursor: 0,
			expectedLog:    0,
		},
		{
			name:           "at boundary condition",
			initialCursor:  1,
			testCount:      3,
			autoToggleMode: false,
			expectedCursor: 0,
			expectedLog:    0,
		},
		{
			name:           "single test should not move",
			initialCursor:  0,
			testCount:      1,
			autoToggleMode: false,
			expectedCursor: 0,
			expectedLog:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := createTestModel(testModelOpts{
				testCount:      tt.testCount,
				autoToggleMode: tt.autoToggleMode,
			})
			m.cursor.test = tt.initialCursor

			m.PrevTest()

			if m.cursor.test != tt.expectedCursor {
				t.Errorf("cursor.test = %d, want %d", m.cursor.test, tt.expectedCursor)
			}
			if m.cursor.log != tt.expectedLog {
				t.Errorf("cursor.log = %d, want %d", m.cursor.log, tt.expectedLog)
			}
		})
	}
}

func TestCursorDown(t *testing.T) {
	tests := []struct {
		name            string
		initialTestIdx  int
		initialLogIdx   int
		testCount       int
		toggled         bool
		logCount        int
		autoToggleMode  bool
		expectedTestIdx int
		expectedLogIdx  int
	}{
		{
			name:            "move within logs when toggled",
			initialTestIdx:  0,
			initialLogIdx:   0,
			testCount:       3,
			toggled:         true,
			logCount:        5,
			autoToggleMode:  false,
			expectedTestIdx: 0,
			expectedLogIdx:  1,
		},
		{
			name:            "move to next test when at last log",
			initialTestIdx:  0,
			initialLogIdx:   4,
			testCount:       3,
			toggled:         true,
			logCount:        5,
			autoToggleMode:  false,
			expectedTestIdx: 1,
			expectedLogIdx:  0,
		},
		{
			name:            "move to next test when not toggled",
			initialTestIdx:  0,
			initialLogIdx:   0,
			testCount:       3,
			toggled:         false,
			logCount:        0,
			autoToggleMode:  false,
			expectedTestIdx: 1,
			expectedLogIdx:  0,
		},
		{
			name:            "at last test should not move",
			initialTestIdx:  2,
			initialLogIdx:   0,
			testCount:       3,
			toggled:         false,
			logCount:        0,
			autoToggleMode:  false,
			expectedTestIdx: 2,
			expectedLogIdx:  0,
		},
		{
			name:            "at last test last log should not move",
			initialTestIdx:  2,
			initialLogIdx:   3,
			testCount:       3,
			toggled:         true,
			logCount:        4,
			autoToggleMode:  false,
			expectedTestIdx: 2,
			expectedLogIdx:  3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := createTestModel(testModelOpts{
				testCount:      tt.testCount,
				autoToggleMode: tt.autoToggleMode,
				logCount:       tt.logCount,
			})
			m.cursor.test = tt.initialTestIdx
			m.cursor.log = tt.initialLogIdx

			test := m.testManager.GetTest(tt.initialTestIdx)
			if test != nil {
				m.testState[test.Ref].toggled = tt.toggled
			}

			m.CursorDown()

			if m.cursor.test != tt.expectedTestIdx {
				t.Errorf("cursor.test = %d, want %d", m.cursor.test, tt.expectedTestIdx)
			}
			if m.cursor.log != tt.expectedLogIdx {
				t.Errorf("cursor.log = %d, want %d", m.cursor.log, tt.expectedLogIdx)
			}
		})
	}
}

func TestCursorUp(t *testing.T) {
	tests := []struct {
		name            string
		initialTestIdx  int
		initialLogIdx   int
		testCount       int
		toggled         bool
		logCount        int
		autoToggleMode  bool
		expectedTestIdx int
		expectedLogIdx  int
	}{
		{
			name:            "move within logs when at higher log index",
			initialTestIdx:  1,
			initialLogIdx:   2,
			testCount:       3,
			toggled:         true,
			logCount:        5,
			autoToggleMode:  false,
			expectedTestIdx: 1,
			expectedLogIdx:  1,
		},
		{
			name:            "move to previous test when at first log",
			initialTestIdx:  1,
			initialLogIdx:   0,
			testCount:       3,
			toggled:         false,
			logCount:        5,
			autoToggleMode:  false,
			expectedTestIdx: 0,
			expectedLogIdx:  0,
		},
		{
			name:            "at first test should not move",
			initialTestIdx:  0,
			initialLogIdx:   0,
			testCount:       3,
			toggled:         false,
			logCount:        0,
			autoToggleMode:  false,
			expectedTestIdx: 0,
			expectedLogIdx:  0,
		},
		{
			name:            "move to previous test goes to last log if toggled",
			initialTestIdx:  1,
			initialLogIdx:   0,
			testCount:       3,
			toggled:         false,
			logCount:        5,
			autoToggleMode:  true,
			expectedTestIdx: 0,
			expectedLogIdx:  4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := createTestModel(testModelOpts{
				testCount:      tt.testCount,
				autoToggleMode: tt.autoToggleMode,
				logCount:       tt.logCount,
			})
			m.cursor.test = tt.initialTestIdx
			m.cursor.log = tt.initialLogIdx

			// Set up the toggled state for the previous test if we're testing auto-toggle
			if tt.autoToggleMode && tt.initialTestIdx > 0 {
				prevTest := m.testManager.GetTest(tt.initialTestIdx - 1)
				if prevTest != nil {
					m.testState[prevTest.Ref].toggled = true
				}
			}

			m.CursorUp()

			if m.cursor.test != tt.expectedTestIdx {
				t.Errorf("cursor.test = %d, want %d", m.cursor.test, tt.expectedTestIdx)
			}
			if m.cursor.log != tt.expectedLogIdx {
				t.Errorf("cursor.log = %d, want %d", m.cursor.log, tt.expectedLogIdx)
			}
		})
	}
}

func TestNextFailingTest(t *testing.T) {
	tests := []struct {
		name           string
		initialCursor  int
		testStatuses   []string
		autoToggleMode bool
		expectedCursor int
		expectedLog    int
	}{
		{
			name:           "move to next failing test",
			initialCursor:  0,
			testStatuses:   []string{"pass", "fail", "pass"},
			autoToggleMode: false,
			expectedCursor: 1,
			expectedLog:    0,
		},
		{
			name:           "skip passing tests",
			initialCursor:  0,
			testStatuses:   []string{"pass", "pass", "fail", "pass"},
			autoToggleMode: false,
			expectedCursor: 2,
			expectedLog:    0,
		},
		{
			name:           "no failing tests ahead",
			initialCursor:  0,
			testStatuses:   []string{"pass", "pass", "pass"},
			autoToggleMode: false,
			expectedCursor: 0,
			expectedLog:    0,
		},
		{
			name:           "at last test no move",
			initialCursor:  2,
			testStatuses:   []string{"fail", "fail", "fail"},
			autoToggleMode: false,
			expectedCursor: 2,
			expectedLog:    0,
		},
		{
			name:           "with auto toggle mode",
			initialCursor:  0,
			testStatuses:   []string{"pass", "fail", "pass"},
			autoToggleMode: true,
			expectedCursor: 1,
			expectedLog:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := createTestModel(testModelOpts{
				autoToggleMode: tt.autoToggleMode,
				testStatuses:   tt.testStatuses,
			})
			m.cursor.test = tt.initialCursor

			m.NextFailingTest()

			if m.cursor.test != tt.expectedCursor {
				t.Errorf("cursor.test = %d, want %d", m.cursor.test, tt.expectedCursor)
			}
			if m.cursor.log != tt.expectedLog {
				t.Errorf("cursor.log = %d, want %d", m.cursor.log, tt.expectedLog)
			}
		})
	}
}

func TestPrevFailingTest(t *testing.T) {
	tests := []struct {
		name           string
		initialCursor  int
		testStatuses   []string
		autoToggleMode bool
		expectedCursor int
		expectedLog    int
	}{
		{
			name:           "move to previous failing test",
			initialCursor:  2,
			testStatuses:   []string{"fail", "pass", "pass"},
			autoToggleMode: false,
			expectedCursor: 0,
			expectedLog:    0,
		},
		{
			name:           "skip passing tests",
			initialCursor:  3,
			testStatuses:   []string{"fail", "pass", "pass", "pass"},
			autoToggleMode: false,
			expectedCursor: 0,
			expectedLog:    0,
		},
		{
			name:           "no failing tests before",
			initialCursor:  2,
			testStatuses:   []string{"pass", "pass", "pass"},
			autoToggleMode: false,
			expectedCursor: 2,
			expectedLog:    0,
		},
		{
			name:           "at first test no move",
			initialCursor:  0,
			testStatuses:   []string{"fail", "fail", "fail"},
			autoToggleMode: false,
			expectedCursor: 0,
			expectedLog:    0,
		},
		{
			name:           "with auto toggle mode",
			initialCursor:  2,
			testStatuses:   []string{"fail", "pass", "pass"},
			autoToggleMode: true,
			expectedCursor: 0,
			expectedLog:    0,
		},
		{
			name:           "multiple failing tests find closest",
			initialCursor:  3,
			testStatuses:   []string{"fail", "fail", "pass", "pass"},
			autoToggleMode: false,
			expectedCursor: 1,
			expectedLog:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := createTestModel(testModelOpts{
				autoToggleMode: tt.autoToggleMode,
				testStatuses:   tt.testStatuses,
			})
			m.cursor.test = tt.initialCursor

			m.PrevFailingTest()

			if m.cursor.test != tt.expectedCursor {
				t.Errorf("cursor.test = %d, want %d", m.cursor.test, tt.expectedCursor)
			}
			if m.cursor.log != tt.expectedLog {
				t.Errorf("cursor.log = %d, want %d", m.cursor.log, tt.expectedLog)
			}
		})
	}
}

type testModelOpts struct {
	testCount      int
	autoToggleMode bool
	logCount       int
	testStatuses   []string
}

func createTestModel(opts testModelOpts) *siftModel {
	m := NewSiftModel(SiftOptions{})
	m.autoToggleMode = opts.autoToggleMode

	testCount := opts.testCount
	if len(opts.testStatuses) > 0 {
		testCount = len(opts.testStatuses)
	}

	for i := 0; i < testCount; i++ {
		testRef := tests.TestReference{
			Package: "test/package",
			Test:    "Test" + string(rune('A'+i)),
		}

		m.testManager.AddTestOutput(tests.TestOutputLine{
			Action:  "run",
			Package: testRef.Package,
			Test:    testRef.Test,
		})

		for j := 0; j < opts.logCount; j++ {
			m.testManager.AddTestOutput(tests.TestOutputLine{
				Action:  "output",
				Package: testRef.Package,
				Test:    testRef.Test,
				Output:  "log line " + string(rune('0'+j)),
			})
		}

		status := "pass"
		if len(opts.testStatuses) > 0 && i < len(opts.testStatuses) {
			status = opts.testStatuses[i]
		}
		m.testManager.AddTestOutput(tests.TestOutputLine{
			Action:  status,
			Package: testRef.Package,
			Test:    testRef.Test,
		})

		m.testState[testRef] = &testState{
			toggled:     false,
			viewportPos: i,
		}
	}

	return m
}
