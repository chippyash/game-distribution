# Desktop App Requirements

## Overview

A desktop version of the Game Distributor. The current CLI (`gamedist`) takes a number of fixtures, players and players
per fixture, and produces a player-to-fixture distribution matrix that can be exported to a spreadsheet. The desktop app
should make the same capability available to users who do not want to use a terminal.

## Scope

- One distribution is created at a time.
- The theme choice is remembered and reloaded between sessions, using Fyne's native preferences. No other settings are saved or reloaded.
- The CLI is unchanged and remains part of the release.

## Functional Requirements

- Inputs: number of fixtures, number of players, players per fixture (see `README.md`). The input labels are "Num. of Fixture", "Num. of Players", "Players per fixture", "Max players together" and "Max. Attempts".
- Options matching the CLI: maximum players together, attempts, fixture spacing, transpose
- The options start at the CLI defaults: attempts 1000, maximum players together 1, fixture spacing off, transpose off
- Inputs are validated:
  - attempts between 1000 and 100000, as in the CLI
  - fixtures, players, players per fixture and maximum players together are at least 1
  - players per fixture does not exceed players
  - fixtures have an upper bound of 200, and players have an upper bound of 100
  - an out-of-range value shows an inline error and blocks Start
- Output: Draw the matrix on screen with an option to download the result as a CSV file
- The matrix shows `X` for a player in a fixture and a blank cell otherwise, with rows labelled `P1…` and columns labelled
  `F1…` as in the CSV
- The input fields are spaced closely, and an error message is shown only when it has something to say, so that the Save CSV button stays on screen.
- The on-screen grid scrolls in both directions. Players are the rows by default, as in the CLI.
- The grid is tested at 100 fixtures and 50 players
- A run executes in the background, because high attempt counts can take many seconds
- A spinner shows that a run is in progress, and a cancel button stops it
- A cancelled run shows no matrix. The secondary panel reports "Cancelled after N attempts".
- The main results panel shows only the generated matrix
- A secondary results panel shows the CLI's summary (players, players that must play, extra-play players, attempt count)
  and all other feedback and errors, including "too many attempts", the effective Max Play Together, and the spacing note,
  using the same wording as the CLI. While a run is in progress, the panel shows "Attempt N of M" using the CLI's wording.
- Save is disabled until a matrix exists
- Start is disabled while a run is running or cancelling
- Downloading the result opens a save dialog with a default name of `distribution.csv`, in the same format as the CLI
- When the transpose option is on, the CSV is transposed too, so it matches what is on screen
- Toggling transpose re-draws the current matrix immediately
- Changing fixtures, players, players per fixture, maximum players together, attempts or fixture spacing clears the matrix and
  disables Save until the next run, so a saved file never disagrees with the inputs on screen
- An application menu in the window, with three entries: Theme, About and Help
- Theme opens a dialog with three choices: light, dark, and following the system setting. The dialog has OK and Cancel buttons. OK applies the chosen theme and remembers it between sessions, using Fyne's native preferences. Cancel changes nothing.
- The app starts in the light theme until a choice is made.
- The splash screen is not affected by the theme.
- Menu entries have no keyboard shortcuts.
- Help opens `https://github.com/chippyash/game-distribution/blob/<version>/README.md` in the user's default web browser, where `<version>` is the value in the `VERSION` file, for example `v1.0.0`.
- If the browser can't be opened for Help, a dialog shows the error
- About opens a dialog that shows the app version and the FSL-1.1 notice. The last line of the dialog is a link to `https://github.com/chippyash/game-distribution/releases/tag/<version>`, where `<version>` is the value in the `VERSION` file. The link replaces the reference to the LICENSE file. The menu's About entry replaces the About button, which is removed from the window.
- The window shows `src/desktop/resources/icon-linux.jpeg` as its icon
- When the app opens, `src/desktop/resources/splash.jpeg` is shown in the window, scaled to fit and centred. After 2 seconds it fades out over 3 seconds to reveal the main window. Input is blocked while the splash is showing, including during the fade. This happens on every launch.

## Non-Functional Requirements

- Supporting app framework: https://fyne.io/ Github: https://github.com/fyne-io/fyne. This adds a large dependency tree, which is accepted.
- Use the current latest Fyne that works. Update the Go version in `go.mod` to match what that Fyne version requires. This applies to the whole module, including the CLI.
- Target platforms: Linux and Windows
- Windows binary is cross-compiled from Linux with a MinGW-w64 cross-compiler, so that `make build` still produces both binaries
- The Windows app hides the console window
- The preferences app ID is `com.chippyash.gamedistributor`
- The Help link uses the version injected at build time from the `VERSION` file, so each build links to the README for that version
- The icon image is embedded in the binary at build time, so the app does not depend on the file's location when it runs. JPEG is accepted; no conversion is needed.
- The splash image is embedded in the binary at build time, as the window icon is.
- The Windows `.exe` has `src/desktop/resources/icon-windows.ico` as its file icon. The icon is compiled into the executable as a Windows resource, so the binary is self-contained.
- Each binary embeds its own icon. The Linux binary embeds `icon-linux.jpeg` for its window icon. The Windows binary embeds `icon-windows.ico` for its file icon, and `icon-linux.jpeg` for its window icon.
- Distribution method: binaries packaged into dist/gamedist.zip as gamedistapp (Linux) and gamedistapp.exe (Windows), alongside the CLI binaries
- `dist/` is no longer tracked in git, and is listed in `.gitignore` again. The release workflow builds the zip and attaches it to the draft release.
- The release workflow installs the Linux GUI development packages (OpenGL and X11 headers) and MinGW-w64, so that tagged releases include the desktop app
- The app version shown in the About dialog is injected at build time with `-ldflags "-X"`, read from the root `VERSION` file in the Makefile. The desktop app has no second copy of the version. The CLI is unchanged.
- Desktop apps reuse `src/distributor` directly.
- `Distribute` accepts a `context.Context`, and checks `ctx.Err()` at the start of each attempt, so that a run can be cancelled from another goroutine
- The distributor exposes a read-only attempt counter, updated atomically once per attempt, that the app can read while a run is in progress
- The distributor's panic recovery returns an error instead of `(nil, nil)`, so the app can show the failure in the secondary results panel
- App code lives in `src/desktop/`. Logic that does not depend on Fyne is in a sub-package with no Fyne import.
- The Fyne entry point is behind a build tag (`desktop`). `make build` and the release workflow pass the tag.
- `make test`, `go vet` and `make lint` run without the GUI packages installed, because the Fyne code is excluded without the tag.
- The desktop package has automated tests, run by `make test`. Layout and logic are kept in functions that can be tested without the UI.
- Fixture spacing and play-together behaviour are same as the CLI. On screen toggle switches reflect their state.

## Open Questions

1. **Linux runtime libraries.** Should the release notes or README say that Linux users need OpenGL and X11 libraries installed? *Left open. Revisit once the Fyne code is in the code base, to see whether it is necessary.*
