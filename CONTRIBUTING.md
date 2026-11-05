# Contributing

## Licence header

Every Go source file, including tests, starts with the FSL-1.1 header, followed by a blank line before `package`. Copy the
header from the top of `src/distributor/distributor.go` when you create a new `.go` file, and do not change it.

## Commands

Go module: `github.com/chippyash/gamedistributor`. `go.mod` says `go 1.22.0` because Fyne v2.8.2 needs it. Use a Go
toolchain that satisfies that, or leave `GOTOOLCHAIN=auto` (the default) so Go fetches the right one.

- Build the CLI only: `go build -o /dev/null ./src/cli/`
- Run the CLI: `go run ./src/cli 24 7 3 -c`
- Run the desktop app from source: `go run -tags desktop ./src/desktop/`
- Test the CLI packages: `go test ./src/...`
- Test the desktop app: `go test -tags desktop ./src/desktop/...`
- Run one test: `go test -tags desktop ./src/desktop -run <TestName> -v`
- Lint: `make lint`. Desktop code is behind a build tag, so also run `golangci-lint run --build-tags desktop ./src/desktop/...`

Do not run `make test` for routine checks. It runs `go get -u`, which rewrites `go.mod`. `make build` also runs `clean`,
which runs `go mod tidy` and can rewrite `go.mod` and `go.sum`.

Tests use the standard `testing` package and exercise exported APIs only (external `_test` packages). `src/cli` is
`package main`, so its tests build the binary and run it. Desktop UI tests use Fyne's `test` package (`test.NewApp()`).

## Versions

The version comes from the root `VERSION` file only. The Makefile injects it into the CLI and the desktop app with
`-ldflags "-X main.version=..."`. The VERSION file and package tag will be updated by the repo maintainer following a successful
merge of a pull request to the master branch.

## Desktop app development notes

Fyne gotchas found while building the desktop app. Each one cost time to find.

- **Overlays take the keyboard.** A `widget.PopUp` is an overlay, and while one is shown typed characters go to it, not
  the inputs. The hover help is therefore a card in the window content (`helpTip` in `src/desktop/hover.go`), not a popup.
- **Owner widgets.** A Fyne widget's `BaseWidget` has one owner, set once. `widget.NewEntry`, `NewLabel` and `NewCheck`
  make the inner widget its own owner. When wrapping them, build the inner widget as a struct literal
  (`&widget.Entry{...}`, `&widget.Label{}`, `&widget.Check{...}`) and then call `wrapper.ExtendBaseWidget(wrapper)`.
  Otherwise focus, layout and refresh go to the inner widget, and the field stops accepting input or stops drawing its icon.
- **Hit testing.** Fyne sends pointer events to the deepest hoverable widget under the pointer. A wrapper that is not the
  widget under the pointer never gets hover events.
- **Spurious leave events.** Some systems, Wine among them, send a leave straight after an enter. The hover help therefore
  hides only after a 300 ms grace period (`hideGrace`), and a click hides it at once.
- **Layout after changes.** When a widget's visibility or size changes inside a container, the container may not re-lay
  out. Call `Refresh` on the container.
- **Deprecated theme calls.** `theme.LightTheme()` and `theme.DarkTheme()` are deprecated. Use
  `newVariantTheme(theme.VariantDark)` from `src/desktop/theme.go`.
- **Wayland pointer offset.** On a Wayland session the pointer does not line up with the widgets, so `main.go` selects the
  X11 (XWayland) GLFW platform when it is available.

## Releases

Release notes are written on the GitHub release. See [docs/releasing.md](docs/releasing.md).

## TODO

 - [ ] Find someone willing to build this app on the Mac platform, as the author does not have a Mac and never will.
