package tests

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/timtatt/sift/pkg/logparse"
)

func TestShouldSkipLogLine(t *testing.T) {
	tests := []struct {
		name string
		line string
		want bool
	}{
		{
			name: "skip RUN line",
			line: "=== RUN   TestExample",
			want: true,
		},
		{
			name: "skip RUN line with spaces",
			line: "  === RUN   TestExample",
			want: true,
		},
		{
			name: "skip PASS line",
			line: "--- PASS: TestExample (0.00s)",
			want: true,
		},
		{
			name: "skip FAIL line",
			line: "--- FAIL: TestExample (0.00s)",
			want: true,
		},
		{
			name: "skip SKIP line",
			line: "--- SKIP: TestExample (0.00s)",
			want: true,
		},
		{
			name: "keep regular log line",
			line: "this is a regular log line",
			want: false,
		},
		{
			name: "keep line with RUN in middle",
			line: "test RUN in the middle",
			want: false,
		},
		{
			name: "empty line",
			line: "",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, shouldSkipLogLine(tt.line))
		})
	}
}

func TestAddTestOutput(t *testing.T) {
	t.Run("status transitions", func(t *testing.T) {
		tests := []struct {
			action      string
			elapsed     float64
			wantElapsed time.Duration
		}{
			{
				action:      "run",
				elapsed:     0,
				wantElapsed: 0,
			},
			{
				action:      "pass",
				elapsed:     1.5,
				wantElapsed: time.Duration(1.5 * float64(time.Second)),
			},
			{
				action:      "fail",
				elapsed:     0.5,
				wantElapsed: time.Duration(0.5 * float64(time.Second)),
			},
			{
				action:      "skip",
				elapsed:     0.0,
				wantElapsed: 0,
			},
		}

		for _, tt := range tests {
			t.Run(tt.action, func(t *testing.T) {
				tm := NewTestManager(TestManagerOpts{})

				if tt.action != "run" {
					tm.AddTestOutput(TestOutputLine{
						Action:  "run",
						Package: "pkg",
						Test:    "Test",
					})
				}

				tm.AddTestOutput(TestOutputLine{
					Action:  tt.action,
					Package: "pkg",
					Test:    "Test",
					Elapsed: tt.elapsed,
				})

				test := tm.GetTest(0)
				require.NotNil(t, test)
				assert.Equal(t, tt.action, test.Status)
				if tt.elapsed > 0 {
					assert.Equal(t, tt.wantElapsed, test.Elapsed)
				}
			})
		}
	})

	t.Run("output handling", func(t *testing.T) {
		tm := NewTestManager(TestManagerOpts{ParseLogs: false})
		testRef := TestReference{Package: "pkg", Test: "Test"}

		tm.AddTestOutput(TestOutputLine{
			Action:  "run",
			Package: testRef.Package,
			Test:    testRef.Test,
		})
		tm.AddTestOutput(TestOutputLine{
			Action:  "output",
			Package: testRef.Package,
			Test:    testRef.Test,
			Output:  "log line 1\n",
			Time:    time.Now(),
		})
		tm.AddTestOutput(TestOutputLine{
			Action:  "output",
			Package: testRef.Package,
			Test:    testRef.Test,
			Output:  "log line 2\n",
			Time:    time.Now(),
		})

		assert.Equal(t, 2, tm.GetLogCount(testRef))
		logs := tm.GetLogs(testRef)
		require.Len(t, logs, 2)
		assert.Equal(t, "log line 1", logs[0].Message)
		assert.Equal(t, "log line 2", logs[1].Message)
	})

	t.Run("output skips test lines", func(t *testing.T) {
		tm := NewTestManager(TestManagerOpts{ParseLogs: false})
		testRef := TestReference{Package: "pkg", Test: "Test"}

		tm.AddTestOutput(TestOutputLine{
			Action:  "run",
			Package: testRef.Package,
			Test:    testRef.Test,
		})
		tm.AddTestOutput(TestOutputLine{
			Action:  "output",
			Package: testRef.Package,
			Test:    testRef.Test,
			Output:  "=== RUN   TestExample\n",
			Time:    time.Now(),
		})
		tm.AddTestOutput(TestOutputLine{
			Action:  "output",
			Package: testRef.Package,
			Test:    testRef.Test,
			Output:  "real log line\n",
			Time:    time.Now(),
		})
		tm.AddTestOutput(TestOutputLine{
			Action:  "output",
			Package: testRef.Package,
			Test:    testRef.Test,
			Output:  "--- PASS: TestExample (0.00s)\n",
			Time:    time.Now(),
		})

		assert.Equal(t, 1, tm.GetLogCount(testRef))
		logs := tm.GetLogs(testRef)
		require.Len(t, logs, 1)
		assert.Equal(t, "real log line", logs[0].Message)
	})

	t.Run("with parse logs enabled", func(t *testing.T) {
		tm := NewTestManager(TestManagerOpts{ParseLogs: true})
		testRef := TestReference{Package: "pkg", Test: "Test"}

		tm.AddTestOutput(TestOutputLine{
			Action:  "run",
			Package: testRef.Package,
			Test:    testRef.Test,
		})
		tm.AddTestOutput(TestOutputLine{
			Action:  "output",
			Package: testRef.Package,
			Test:    testRef.Test,
			Output:  `time=2025-10-05T09:52:58.046+11:00 level=INFO msg="This is an info message" key1=value1`,
			Time:    time.Now(),
		})

		logs := tm.GetLogs(testRef)
		require.Len(t, logs, 1)
		assert.Equal(t, "INFO", logs[0].Level)
		assert.Equal(t, "2025-10-05T09:52:58+11:00", logs[0].Time.Format(time.RFC3339))
		assert.Equal(t, "This is an info message", logs[0].Message)
		assert.Contains(t, logs[0].Additional, logparse.LogEntryAdditionalProp{Key: "key1", Value: "value1"})
	})
}

func TestGetTest(t *testing.T) {
	tm := NewTestManager(TestManagerOpts{})
	tm.AddTestOutput(TestOutputLine{
		Action:  "run",
		Package: "pkg1",
		Test:    "Test1",
	})
	tm.AddTestOutput(TestOutputLine{
		Action:  "run",
		Package: "pkg2",
		Test:    "Test2",
	})

	tests := []struct {
		name     string
		index    int
		wantNil  bool
		wantTest string
		wantPkg  string
	}{
		{
			name:     "first test",
			index:    0,
			wantNil:  false,
			wantTest: "Test1",
			wantPkg:  "pkg1",
		},
		{
			name:     "second test",
			index:    1,
			wantNil:  false,
			wantTest: "Test2",
			wantPkg:  "pkg2",
		},
		{
			name:    "negative index",
			index:   -1,
			wantNil: true,
		},
		{
			name:    "out of bounds",
			index:   10,
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			test := tm.GetTest(tt.index)

			if tt.wantNil {
				assert.Nil(t, test)
				return
			}

			require.NotNil(t, test)
			assert.Equal(t, tt.wantTest, test.Ref.Test)
			assert.Equal(t, tt.wantPkg, test.Ref.Package)
		})
	}
}

func TestGetTestCount(t *testing.T) {
	tm := NewTestManager(TestManagerOpts{})
	assert.Equal(t, 0, tm.GetTestCount())

	for i := 1; i <= 5; i++ {
		tm.AddTestOutput(TestOutputLine{
			Action:  "run",
			Package: "pkg",
			Test:    "Test",
		})
		assert.Equal(t, i, tm.GetTestCount())
	}
}

func TestGetTests(t *testing.T) {
	tm := NewTestManager(TestManagerOpts{})
	tm.AddTestOutput(TestOutputLine{Action: "run", Package: "pkg", Test: "Test1"})
	tm.AddTestOutput(TestOutputLine{Action: "run", Package: "pkg", Test: "Test2"})
	tm.AddTestOutput(TestOutputLine{Action: "run", Package: "pkg", Test: "Test3"})

	t.Run("iterate all", func(t *testing.T) {
		var collected []*TestNode
		tm.GetTests(func(i int, test *TestNode) bool {
			collected = append(collected, test)
			return true
		})
		assert.Len(t, collected, 3)
	})

	t.Run("early exit", func(t *testing.T) {
		var collected []*TestNode
		tm.GetTests(func(i int, test *TestNode) bool {
			collected = append(collected, test)
			return i < 1
		})
		assert.Len(t, collected, 2)
	})
}

func TestGetLogCount(t *testing.T) {
	tm := NewTestManager(TestManagerOpts{ParseLogs: false})
	testRef := TestReference{Package: "pkg", Test: "Test"}

	t.Run("increments with output", func(t *testing.T) {
		assert.Equal(t, 0, tm.GetLogCount(testRef))
		tm.AddTestOutput(TestOutputLine{
			Action:  "run",
			Package: testRef.Package,
			Test:    testRef.Test,
		})

		for i := 1; i <= 3; i++ {
			tm.AddTestOutput(TestOutputLine{
				Action:  "output",
				Package: testRef.Package,
				Test:    testRef.Test,
				Output:  "log\n",
				Time:    time.Now(),
			})
			assert.Equal(t, i, tm.GetLogCount(testRef))
		}
	})

	t.Run("non-existent test", func(t *testing.T) {
		nonExistentRef := TestReference{Package: "nonexistent", Test: "Test"}
		assert.Equal(t, 0, tm.GetLogCount(nonExistentRef))
	})
}

func TestGetLogs(t *testing.T) {
	tm := NewTestManager(TestManagerOpts{ParseLogs: false})
	testRef := TestReference{Package: "pkg", Test: "Test"}

	t.Run("returns logs", func(t *testing.T) {
		tm.AddTestOutput(TestOutputLine{
			Action:  "run",
			Package: testRef.Package,
			Test:    testRef.Test,
		})
		tm.AddTestOutput(TestOutputLine{
			Action:  "output",
			Package: testRef.Package,
			Test:    testRef.Test,
			Output:  "log 1\n",
			Time:    time.Now(),
		})
		tm.AddTestOutput(TestOutputLine{
			Action:  "output",
			Package: testRef.Package,
			Test:    testRef.Test,
			Output:  "log 2\n",
			Time:    time.Now(),
		})

		logs := tm.GetLogs(testRef)
		require.Len(t, logs, 2)
		assert.Equal(t, "log 1", logs[0].Message)
		assert.Equal(t, "log 2", logs[1].Message)
	})

	t.Run("non-existent test", func(t *testing.T) {
		nonExistentRef := TestReference{Package: "nonexistent", Test: "Test"}
		assert.Nil(t, tm.GetLogs(nonExistentRef))
	})
}

func TestBuildErrorHandling(t *testing.T) {
	tm := NewTestManager(TestManagerOpts{ParseLogs: false})

	tm.AddTestOutput(TestOutputLine{
		Action:     "build-output",
		ImportPath: "github.com/example/pkg",
		Output:     "# github.com/example/pkg\n",
		Time:       time.Now(),
	})
	tm.AddTestOutput(TestOutputLine{
		Action:  "build-output",
		Package: "github.com/example/pkg",
		Output:  "pkg/file.go:10:5: undefined: someFunction\n",
		Time:    time.Now(),
	})
	tm.AddTestOutput(TestOutputLine{
		Action:  "build-fail",
		Package: "github.com/example/pkg",
	})

	assert.Equal(t, 1, tm.GetTestCount())

	test := tm.GetTest(0)
	require.NotNil(t, test)
	assert.Equal(t, "github.com/example/pkg", test.Ref.Package)
	assert.Equal(t, "", test.Ref.Test)
	assert.Equal(t, "error", test.Status)

	buildErrRef := TestReference{Package: "github.com/example/pkg", Test: ""}
	assert.Equal(t, 1, tm.GetLogCount(buildErrRef))

	logs := tm.GetLogs(buildErrRef)
	require.Len(t, logs, 1)
	assert.Equal(t, "pkg/file.go:10:5: undefined: someFunction", logs[0].Message)
}
