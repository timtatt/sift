# AGENTS.md

This document provides context for AI coding agents working on the sift project.

## Project Overview

**sift** is a lightweight terminal UI for displaying Go test results. It allows developers to traverse verbose Go test logs in their terminal, with the ability to expand and collapse individual tests to focus only on the logs that matter.

### Purpose

The tool was created to address the overwhelming nature of Go's test output, especially for large test suites. When running tests from the command line with `go test -v -json`, the verbose output can make it difficult to quickly identify and drill into failing or interesting tests. sift provides an interactive, vim-inspired interface for navigating test results without needing an IDE.

## Key Features

- **Interactive Test Browsing**: Collapse and expand test outputs to reduce visual clutter
- **Fuzzy Search/Filter**: Case-insensitive fuzzy search to quickly find specific tests
- **Vim-Inspired Navigation**: Familiar keymaps for developers comfortable with vim motions
- **Smart Log Parsing**: Automatically detects and prettifies structured logs (slog JSON/text, default logs)
- **Two View Modes**:
  - Interactive mode with alternate screen (default)
  - Non-interactive inline mode for CI/CD pipelines
- **Auto-Toggle Mode**: Automatically opens/closes tests as you navigate
- **Test Status Indicators**: Visual icons for pass/fail/skip/running states

## Architecture

### Technology Stack

- **Language**: Go 1.25.0
- **UI Framework**: [Bubble Tea](https://github.com/charmbracelet/bubbletea) (TUI framework based on The Elm Architecture)
- **Styling**: [Lipgloss](https://github.com/charmbracelet/lipgloss) for terminal styling
- **UI Components**: [Bubbles](https://github.com/charmbracelet/bubbles) for viewport and text input
- **CLI Parsing**: [Kong](https://github.com/alecthomas/kong)
- **Search**: [fuzzysearch](https://github.com/lithammer/fuzzysearch) for fuzzy matching
- **Testing**: [testify](https://github.com/stretchr/testify) for test assertions

### Project Structure

```
internal/
├── sift/           # Core bubbletea view
├── tests/          # Test manager and data structures
pkg/
├── logparse/          # Log parsing utilities
├── viewbuilder/       # View construction helper
└── helpview/          # Wrapping help view
```

### Core Components

#### 1. Main Application (`internal/sift/sift.go`)

The `sift` struct orchestrates three concurrent goroutines using `errgroup`:

- **Stdin Scanner**: Reads JSON test output line-by-line from stdin
- **Bubble Tea Program**: Handles UI rendering and user input
- **Frame Loop**: Sends periodic frame messages for smooth updates (120 FPS)

#### 2. Test Manager (`internal/tests/test_manager.go`)

Thread-safe manager that:

- Stores test nodes with status, elapsed time, and metadata
- Maintains separate map of test logs (parsed log entries)
- Uses RWMutex for concurrent access from scanner and UI goroutines
- Processes JSON output from `go test -json` command
- Filters out redundant Go test output lines (e.g., `=== RUN`, `--- PASS`)

#### 3. UI Model (`internal/sift/view.go`)

Implements the Bubble Tea architecture:

- **Model**: Contains test state, cursor position, viewport, search input
- **Update**: Handles all user input events and key combinations
- **View**: Delegates to interactive or inline view renderers

Key state management:

- `testState`: Map of test references to their toggle/viewport state
- `cursor`: Tracks currently selected test and log line
- `autoToggleMode`: Automatically expands/collapses tests during navigation
- `searchInput`: Bubble Tea text input for fuzzy search

#### 4. Log Parser (`pkg/logparse/`)

Attempts to parse logs in order:

1. Default structured log format
2. Slog JSON format
3. Slog text format
4. Falls back to raw message

Extracts timestamp, log level, message, and additional fields for prettification.

### Data Flow

```
stdin (go test -json)
    ↓
Scanner goroutine → TestManager.AddTestOutput()
    ↓
TestManager (thread-safe storage)
    ↓
Frame loop → FrameMsg → Model.Update()
    ↓
Model.View() → interactive/inline view
    ↓
Terminal output
```

### Key Concepts

#### Test References

Each test is uniquely identified by a `TestReference` struct containing:

- `Package`: The Go package path
- `Test`: The test name (including subtest hierarchy with `/` separators)

#### Cursor System

The cursor tracks two positions:

- `cursor.test`: Index of the currently selected test
- `cursor.log`: Index of the currently selected log line within that test

#### Viewport Management

- Automatically scrolls to keep cursor visible with a 5-line buffer
- Adjusts height dynamically based on content and footer size
- Syncs with test toggle states to update content

#### Search/Filter

- Uses fuzzy matching on test names (case-insensitive)
- Filters tests in real-time as user types
- Automatically adjusts cursor to nearest visible test

## Input Format

sift consumes the JSON output from Go's test command:

```bash
go test ./... -v -json | sift
```

Each line is a JSON object with fields:

- `time`: Timestamp
- `action`: `run`, `output`, `pass`, `fail`, `skip`
- `package`: Package path
- `test`: Test name (optional)
- `elapsed`: Duration in seconds (optional)
- `output`: Log line (optional)

## Development Notes

### Testing

- Unit tests exist for core components (test_manager, logparse, view)
- Uses testify for test assertions
- Sample tests provided in `samples/` directory for manual testing
- Run demo: `go test ./samples/... -v -json | go run .`

### Thread Safety

The TestManager uses read-write locks extensively since it's accessed concurrently by:

- The scanner goroutine (writes)
- The Bubble Tea update/view methods (reads)

### Performance

- Frame rate set to 120 FPS for smooth animations
- View builder tracks line counts for efficient viewport positioning
- Logs are stored separately from test nodes to optimize memory
