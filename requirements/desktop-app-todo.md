# Desktop App Implementation To-Do

This is the implementation plan for `requirements/desktop-app.md`. Each task is marked with who carries it out:

- **[Claude]**: I can do this in the repository.
- **[You]**: needs an action from you, such as installing software, using a Windows machine, or pushing and releasing.
- **[Both]**: I do the work, and you check or decide something.

Status: Phases 0 to 2 are complete. Phase 3 is complete except the comparison with the CLI output. Phase 4 has its code
compiled, and awaits a run on a Linux desktop. Phases 5 to 8 are not started.

## Environment check

- gcc and the GL and X11 development headers are installed.
- The MinGW-w64 cross-compiler is installed.
- Go is 1.21.0 locally. `GOTOOLCHAIN=auto` is in use, and Go fetches the toolchain that the Fyne version needs. `go.mod` now says `go 1.22.0`.
- The Wayland client headers are installed. The Fyne code compiles and vets with the `desktop` tag.

## Phase 0: Prerequisites

- [x] **[You]** Install the MinGW-w64 cross-compiler: `sudo apt install gcc-mingw-w64-x86-64`
- [x] **[Claude]** Use `GOTOOLCHAIN=auto` as required. Go has fetched Go 1.22 for the Fyne dependency.
- [x] **[You]** Install the Wayland development headers: `sudo apt install libwayland-dev libxkbcommon-dev`. This is needed to compile the Fyne code.
- [x] **[You]** Access to a Windows machine is available.
- [ ] **[You]** Test `gamedistapp.exe` on the Windows machine once it is built (Phase 5).

## Phase 1: Distributor changes (`src/distributor`)

Decision: `Distribute` takes the context as its first argument, so the call becomes `Distribute(ctx)`. The CLI passes `context.Background()`.

- [x] **[Claude]** Change `Distribute` to `Distribute(ctx context.Context) (dist *Distributor, err error)`.
- [x] **[Claude]** Check `ctx.Err()` at the start of each attempt, and return it as the error when cancelled.
- [x] **[Claude]** Add an attempt counter that is safe to read while a run is in progress. It uses `atomic.Int64`, exposed as `AttemptsMade() int`, and increments once per attempt.
- [x] **[Claude]** Make the panic recovery return an error, so the result is never `(nil, nil)`.
- [x] **[Claude]** Update the CLI call site and the existing distributor tests for the new signature.
- [x] **[Claude]** Add tests: a cancelled context returns `context.Canceled`; `AttemptsMade` counts every attempt; a recovered panic returns a non-nil error.

## Phase 2: Input validation (`src/desktop/model`, no Fyne import)

- [x] **[Claude]** Implement validation for: attempts 1000–100000; fixtures, players, players per fixture and maximum players together at least 1; players per fixture not greater than players; fixtures at most 200; players at most 100. See `settings.go`.
- [x] **[Claude]** Return one error per invalid field, so the UI can show it inline next to that field.
- [x] **[Claude]** Write table-driven tests for every rule, including boundary values. See `settings_test.go`.

## Phase 3: Result model (`src/desktop/model`, no Fyne import)

- [x] **[Claude]** Build the secondary panel text from the same wording the CLI uses. Covers the summary, the effective Max Play Together note, the spacing note, "Attempt N of M" and "Cancelled after N attempts". See `summary.go`. "Too many attempts" is the distributor's error text, shown as-is.
- [x] **[Claude]** Build the CSV content in the CLI layout, transposed when the transpose option is on. See `output.go`.
- [x] **[Claude]** Build the matrix cell text: `X` for a player in a fixture and blank otherwise, with `P1…` and `F1…` labels. See `output.go`.
- [ ] **[Claude]** Compare the CSV and the summary with the CLI output for the same inputs. The tests check the expected layout and wording, but not a byte-for-byte comparison with a CLI run.

## Phase 4: Fyne application (`src/desktop`, behind the `desktop` build tag)

- [x] **[Claude]** Add the Fyne dependency (v2.8.2) and update `go.mod` to the Go version it requires (`go 1.22.0`). This also changes the Go version for the CLI.
- [x] **[Claude]** Build the window (`ui.go`). Inputs: fixtures, players, players per fixture, and the four options. Start, Cancel and Save buttons. A spinner and a status line. A matrix grid that scrolls in both directions. A secondary results panel.
- [x] **[Claude]** Enable and disable buttons as specified: Save only when a matrix exists, and Start disabled while a run is running or cancelling.
- [x] **[Claude]** Run each distribution in a goroutine, with a cancellable context. Poll `AttemptsMade` to update the status line. Widgets are only updated through `fyne.Do`.
- [x] **[Claude]** Clear the matrix and disable Save when a non-transpose input changes. Redraw immediately when transpose changes.
- [x] **[Claude]** Save dialog with the default name `distribution.csv`.
- [x] **[Claude]** About dialog showing the version, and the FSL-1.1 notice. The version is set by `main.version` (`main.go`).
- [x] **[Claude]** Compile the Fyne code. `go vet -tags desktop ./src/desktop/` passes, and `go build -tags desktop` produces a Linux binary.
- [x] **[Claude]** Test the grid at 100 fixtures by 50 players, using Fyne's test driver. `src/desktop/ui_test.go` runs a real 100 by 50 distribution, checks the grid size in both orientations, and reads every cell through the table callbacks. It needs the `desktop` tag.
- [x] **[Claude]** Fix the pointer offset reported on the Wayland session: `main.go` asks GLFW for X11 (XWayland) before creating the window. A probe confirmed the X11 platform is selected.
- [x] **[Claude]** Embed `resources/icon-linux.jpeg` with `go:embed` in `main.go`, and set it as the application icon with `a.SetIcon` before the window is created. JPEG is decoded by Fyne, so no conversion is needed.
- [ ] **[Claude]** Test that the embedded icon is non-empty and decodes as an image.
- [X] **[You]** Check that the icon appears in the window title bar on Linux.
- [x] **[Claude]** Create `src/desktop/resources/icon-windows.ico` from `icon-linux.jpeg`, with 16, 24, 32, 48, 64, 128 and 256 pixel sizes.
- [x] **[Claude]** Embed `icon-windows.ico` in the Windows `.exe`. `src/desktop/resources/icon-windows.rc` is compiled to `src/desktop/icon_windows_amd64.syso` by `make desktop-icon`, which is ignored by git. The Windows build checked: the `.exe` has a `.rsrc` section.
- [X] **[You]** Check the icon on the Windows `.exe` in Explorer, and the window icon on Windows.
- [x] **[You]** Decide whether the Windows `.exe` needs an icon as well. Decided: yes.
- [x] **[Claude]** Splash screen: embed `resources/splash.jpeg` with `go:embed`. Show it scaled to fit and centred over the main window, then after 2 seconds fade it out over 3 seconds with a `canvas.Image` `Translucency` animation (`fyne.NewAnimation`), and remove it when the fade finishes. The JPEG is accepted; Fyne decodes it. Input is blocked while the splash is showing, including during the fade.
- [x] **[Claude]** Test that the splash is removed after the fade, and that the controls are usable afterwards. `src/desktop/splash_test.go` checks the timings, the input lock, the fade values and the release of the controls. The splash code is in `src/desktop/splash.go`.
- [X] **[You]** Check the splash on Linux and Windows.

Application menu (in the window: Theme, About, Help; no keyboard shortcuts):

- [x] **[Claude]** Use the preferences app ID `com.chippyash.gamedistributor`: create the app with `app.NewWithID` in `main.go`.
- [x] **[Claude]** Add the in-window menu with three entries: Theme, About and Help. Remove the About button from the window, so the menu's About entry replaces it. The About dialog itself is unchanged.
- [x] **[Claude]** Theme dialog: three choices, light, dark and following the system setting, with OK and Cancel. Show the current choice as selected when the dialog opens. OK applies the theme to the whole window and saves it with Fyne's preferences. Cancel changes nothing.
- [x] **[Claude]** Start in the light theme until a choice has been saved. Read the saved choice when the app opens. The splash is not affected by the theme.
- [x] **[Claude]** Help: open `https://github.com/chippyash/game-distribution/blob/<version>/README.md` in the default browser, where `<version>` is the injected `VERSION` value. Put the URL builder in `src/desktop/model`, so it can be tested without the UI. If the browser can't be opened, show the error in a dialog.
- [x] **[Claude]** Test the Help URL builder, the theme preference round trip (save, then read back), and the menu entries.
- [x] **[Claude]** About dialog: replace the last line (the LICENSE reference) with a link to `https://github.com/chippyash/game-distribution/releases/tag/<version>`, where `<version>` is the `VERSION` value. The URL is built in `src/desktop/model/links.go`, with a test. The link is a Fyne hyperlink in the dialog.
- [x] **[Claude]** Input labels: change "Fixtures" to "Num. of Fixture", "Players" to "Num. of Players" and "Attempts" to "Max. Attempts" in `src/desktop/ui.go`.
- [x] **[Claude]** Reduce the space between the input fields, and hide each field's error message when it is empty, so the Save CSV button stays on screen.
- [x] **[Claude]** Hover help text for each input in `src/desktop/model/help.go`, with a test that every input has one.
- [x] **[Claude]** Hover help widgets in `src/desktop/hover.go`: a 1-second hover delay, a help card inside the window (not an overlay, which would take the keyboard from the inputs), placed beside the input and hidden when the pointer leaves or the input is clicked, and no popup while the splash is showing. Labels and entries are wrapped directly. Check boxes are held in a holder widget, because wrapping `widget.Check` directly loses its icon.
- [x] **[Claude]** Tests for the hover delay, hiding on leave, the splash lock and the help text on every input (`src/desktop/hover_test.go`).
- [X] **[You]** Check the hover help on Linux and Windows: each input and label, the 1-second delay, and that it hides when the pointer leaves.
- [X] **[You]** Check the menu, the theme dialog, the theme change and the Help link on Linux and Windows.
- [x] **[You]** Run the app on Linux and check the pointer position, layout, spinner, cancel, and the save dialog. I can't see the GUI from here.

Decision made while writing `ui.go`: the input fields are disabled while a run is going, so the finished matrix always matches the settings shown on screen. The requirements didn't cover this. Tell me if you want the inputs to stay editable during a run.

## Phase 5: Build (`Makefile`)

- [x] **[Claude]** Add build targets for `gamedistapp` (Linux) and `gamedistapp.exe` (Windows). The Windows target uses `CC=x86_64-w64-mingw32-gcc` and `CGO_ENABLED=1`.
- [x] **[Claude]** Pass `-tags desktop` to the desktop builds only. The CLI build is unchanged.
- [x] **[Claude]** Inject the version with `-ldflags "-X main.version=$(VERSION)"`, read from the root `VERSION` file. Do not add a second copy.
- [x] **[Claude]** Hide the console window on the Windows build with `-ldflags "-H=windowsgui"`.
- [x] **[Claude]** Add `gamedistapp`, `gamedistapp.exe` and the existing CLI binaries to `dist/gamedist.zip`. `build-desktop` appends the two desktop binaries.
- [ ] **[Claude]** Check that `make test`, `go vet ./...` and `make lint` run without the `desktop` tag, and without the GUI packages.
- [ ] **[Claude]** Check that `make build` produces all four binaries.
- [ ] **[You]** Test `gamedistapp.exe` on the Windows machine.

## Phase 6: Release workflow (`.github/workflows/fsl-release.yml`)

- [ ] **[Claude]** Install `libgl1-mesa-dev`, `xorg-dev`, `libwayland-dev` and `gcc-mingw-w64-x86-64` on the Ubuntu runner.
- [ ] **[Claude]** Make sure `make build` runs with the `desktop` tag, so that tagged releases include the desktop app.
- [ ] **[You]** Decide whether the release should also include a macOS build. The requirements exclude macOS, so this is currently out of scope.
- [ ] **[You]** Verify the workflow: push a test tag, check the draft release contents, then delete the test tag and release.

## Phase 7: Documentation

- [ ] **[Claude]** README: describe the desktop app and its binaries, add the new binaries to the release package list, and document build prerequisites for the desktop app.
- [ ] **[Claude]** Update `CLAUDE.md` for the new packages, the `desktop` build tag, and the test commands.
- [ ] **[Both]** Open question: Linux runtime libraries. Decide whether the README says users need OpenGL and X11 libraries installed. I can check this once the code is written, by running the binary on a minimal system.

## Phase 8: Git and release

- [ ] **[Claude]** Commit each phase to `feature/desktop-app`, when you ask.
- [ ] **[You]** Push the branch and open a pull request, or ask me to push once the network is reachable.
- [ ] **[You]** Merge to `master`, and create the release tag when you're ready.

## Summary: what needs you

1. Run the Linux app and check the GUI (Phase 4). Build it with `go build -tags desktop -o bin/gamedistapp ./src/desktop/`.
2. Run the Linux app and check the GUI (Phase 4).
3. Test the Windows app on your Windows machine (Phases 0, 4 and 5).
4. Decide on macOS, and on the Linux runtime libraries (Phases 6 and 7).
5. Verify the release workflow with a test tag (Phase 6).
6. Push, open the pull request, merge, and tag (Phase 8).
7. Decide whether inputs should stay editable during a run (Phase 4).
