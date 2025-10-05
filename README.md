# sift

> _A lightweight terminal UI for displaying Go tests._

sift is a lightweight terminal UI for displaying Go test results. It allows developers to traverse verbose Go test logs in their terminal. Each test is able to be expanded and collapsed to only show the logs that matter.

## Screenshot

<img width="60%" alt="screenshot" src="https://github.com/user-attachments/assets/604c9bf1-d0a2-4e3a-a34d-eb87c7e93a42" />

## Installation

```bash
go install github.com/timtatt/sift@v0.7.0
```

## Try it out!

You can try a demo of sift with the sample tests provided in the `samples` folder.

```bash
# Clone the repo
git clone github.com/timtatt/sift.git

# Run sift
go test ./samples/... -v -json | sift
```

## Usage

`sift` works by consuming the verbose json output from the `go test` command. The easiest way to use it is to pipe `|` the output straight into `sift`

```bash
go test {your-go-package} -v -json | sift

# eg.
go test ./... -v -json | sift
```

## Demo

<video width="60%" src="https://github.com/user-attachments/assets/44b23d46-739b-4956-8894-25ed6d7ae5e9"></video>

### CLI Flags

| Flag                | Shorthand | Description                                     |
| ------------------- | --------- | ----------------------------------------------- |
| `--debug`           | `-d`      | Enable debug view                               |
| `--raw`             | `-r`      | Disable prettified logs                         |
| `--non-interactive` | `-n`      | Skip alternate screen and show inline view only |

**Example:**

```bash
# Run in non-interactive mode (inline output)
go test ./... -v -json | sift -n

# Enable debug view
go test ./... -v -json | sift --debug

# Disable log prettification
go test ./... -v -json | sift --raw
```

### Keymaps

The keymaps are based on vim motion standard keymaps for scrolling and managing folds. Press `?` to toggle the help menu.

#### Navigation

| Key       | Action                       |
| --------- | ---------------------------- |
| `↑` / `k` | Move up                      |
| `↓` / `j` | Move down                    |
| `{`       | Jump to previous test        |
| `}`       | Jump to next test            |
| `[`       | Jump to previous failed test |
| `]`       | Jump to next failed test     |

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
| `m`            | Change mode      |
| `q` / `ctrl+c` | Quit             |

## Credits

The UI design of `sift` is heavily inspired by the [vitest cli](https://github.com/vitest-dev/vitest)
