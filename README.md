# sift - _A lightweight terminal UI for displaying Go tests._

<img src="/assets/screenshot.png" width="60%" alt="Screenshot of the Sift UI">

Sift solves the problem of verbose Go test logs by allowing you to show/hide the logs at an individual tests level.

## Installation

```bash
go install github.com/timtatt/sift@v0.5.0
```

## Usage

`sift` works by consuming the verbose json output from the `go test` command. The easiest way to use it is to pipe `|` the output straight into `sift`

```bash
go test {your-go-package} -v -json | sift

# eg.
go test ./... -v -json | sift
```

### CLI Flags

| Flag                | Shorthand | Description                                     |
| ------------------- | --------- | ----------------------------------------------- |
| `--debug`           |           | Enable debug view                               |
| `--non-interactive` | `-n`      | Skip alternate screen and show inline view only |

**Example:**

```bash
# Run in non-interactive mode (inline output)
go test ./... -v -json | sift -n

# Enable debug view
go test ./... -v -json | sift --debug
```

### Keymaps

The keymaps are based on vim motion standard keymaps for scrolling and managing folds. Press `?` to toggle the help menu.

#### Navigation

| Key       | Action                |
| --------- | --------------------- |
| `↑` / `k` | Move up               |
| `↓` / `j` | Move down             |
| `{`       | Jump to previous test |
| `}`       | Jump to next test     |

#### Viewport Scrolling

| Key      | Action                |
| -------- | --------------------- |
| `ctrl+y` | Scroll viewport up    |
| `ctrl+e` | Scroll viewport down  |
| `ctrl+u` | Scroll half page up   |
| `ctrl+d` | Scroll half page down |

#### Toggle/Expand/Collapse Tests

| Key               | Action                                      |
| ----------------- | ------------------------------------------- |
| `enter` / `space` | Toggle test output                          |
| `za`              | Toggle test output (vim-style)              |
| `zo`              | Expand test output                          |
| `zc`              | Collapse test output                        |
| `zA`              | Toggle test recursively (includes subtests) |
| `zR`              | Expand all tests                            |
| `zM`              | Collapse all tests                          |

#### Search

| Key   | Action                                 |
| ----- | -------------------------------------- |
| `/`   | Enter search mode                      |
| `esc` | Clear search filter and show all tests |

**Search Tips:**

- Type to filter tests using fuzzy matching (case-insensitive)
- Press `enter` to exit search mode while keeping the filter active
- Press `esc` to clear the search filter and show all tests

#### Other

| Key            | Action           |
| -------------- | ---------------- |
| `?`            | Toggle help menu |
| `q` / `ctrl+c` | Quit             |

## Feature Roadmap

- [ ] Filter tests by status (pass/fail/skip)
- [ ] Support for light mode
- [ ] Add animated chars

## Bug Fixes

- [ ] Add responsive wrapping for the help
- [ ] When items collapsed and viewport is too large, rerender to remove whitespace

## Credits

The UI design of `sift` is heavily inspired by the [vitest cli](https://github.com/vitest-dev/vitest)
