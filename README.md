# sift
*A lightweight terminal UI for displaying Go tests.*

![Screenshot of the Sift UI](/assets/screenshot.png "v0.1.0 UI")

Sift solves the problem of verbose Go test logs by allowing you to show/hide the logs at an individual tests level.


## Installation

```bash
go install github.com/timtatt/sift@v0.1.0
```

## Usage

`sift` works by consuming the verbose json output from the `go test` command. The easiest way to use it is to pipe `|` the output straight into `sift` 

```bash
go test {your-go-package} -v -json | sift

# eg. 
go test ./... -v -json | sift
```

### Keymaps

To see the available keymaps, press `?`. The keymaps are based on vim motion standard keymaps for scrolling and managing folds

## Feature Roadmap

- [ ] Add test status to summary
- [ ] Add header
- [ ] Accordion levels based on '/'
- [ ] Nesting the child tests
- [ ] Search for tests with '/'
- [ ] Filter tests by status (pass/fail/skip)
- [ ] Support for light mode
- [ ] Add inline mode to show test summary
- [ ] Add animated chars

## Bug Fixes
- [ ] Add responsive wrapping for the help
- [ ] When items collapsed and viewport is too large, rerender to remove whitespace


## Credits

The UI design of `sift` is heavily inspired by the [vitest cli](https://github.com/vitest-dev/vitest)
