# Desktop App Requirements

## Overview

A desktop version of the Game Distributor. The current CLI (`gamedist`) takes a number of fixtures, players and players
per fixture, and produces a player-to-fixture distribution matrix that can be exported to a spreadsheet. The desktop app
should make the same capability available to users who do not want to use a terminal.

## Scope

- One distribution is created at a time.
- Settings are not saved or reloaded between sessions.
- The CLI is unchanged and remains part of the release.

## Functional Requirements

- Inputs: number of fixtures, number of players, players per fixture (see `README.md`)
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
- An About dialog shows the app version and the FSL-1.1 notice

## Non-Functional Requirements

- Supporting app framework: https://fyne.io/ Github: https://github.com/fyne-io/fyne. This adds a large dependency tree, which is accepted.
- Use the current latest Fyne that works. Update the Go version in `go.mod` to match what that Fyne version requires. This applies to the whole module, including the CLI.
- Target platforms: Linux and Windows
- Windows binary is cross-compiled from Linux with a MinGW-w64 cross-compiler, so that `make build` still produces both binaries
- The Windows app hides the console window
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
