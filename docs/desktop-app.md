# Desktop app guide

Audience: people who want to create a distribution without using a terminal. This guide describes the desktop app as
released in v1.1.0. For the command-line tool, see the [README](../README.md).

The desktop app does the same job as the CLI, in a window. It is `gamedistapp` on Linux and `gamedistapp.exe` on Windows.
Both are in the release archive with the CLI.

![The desktop app main window at start](images/desktop-main-window.png)

*Linux, v1.2.1, dark theme. The Windows appearance has not been checked on a real Windows machine.*

## Starting the app

Run `gamedistapp` on Linux, or double-click `gamedistapp.exe` on Windows.

A splash screen shows for about five seconds. The controls are disabled until it finishes.

## Setting up a distribution

The inputs are on the left, with the labels used in the app:

- **Num. of Fixture:** the number of fixtures, from 1 to 200.
- **Num. of Players:** the number of players, from 1 to 100.
- **Players per fixture:** how many players take part in each fixture. It must be at least 1 and no more than the number of
  players.
- **Max players together:** the most fixtures any two players may share. The default is 1. If that is impossible for the
  other settings, it is raised automatically, and the summary says so.
- **Max. Attempts:** how many random attempts the program makes before giving up, from 1000 to 100000. The default is 1000.
  A run can fail at 1000 attempts even when a solution exists, so raise this if you see "too many attempts". For 24
  fixtures, 7 players and 3 per fixture, 5000 attempts is reliable.
- **Fixture spacing (-s):** tries to stop players from playing in back-to-back fixtures. It is off by default, and it is a
  preference rather than a guarantee.
- **Transpose (-t):** swaps rows and columns, so fixtures are the rows and players are the columns.

An invalid value shows a message in red under its input, and **Start** stays disabled until every input is valid. The
options start at the CLI defaults.

![The inputs filled in for 24 fixtures, 7 players and 3 per fixture](images/desktop-inputs.png)

*Linux, v1.2.1, dark theme.*

## Running a distribution

Press **Start**. The run happens in the background. A spinner shows that it is working, and the status line shows the attempt
number, for example "Attempt 12 of 1000".

Press **Cancel** to stop a run. A cancelled run shows no matrix, and the Summary tab says "Cancelled after N attempts".
The inputs cannot be changed while a run is going.

## Reading the results

- The **Matrix** tab shows the matrix. An `X` marks a player taking part in a fixture, and a blank cell means they are not.
  Rows and columns are labelled `P1` to `Pn` for players and `F1` to `Fn` for fixtures. Players are the rows by default.
  Scroll in either direction for large matrices.
- The **Summary** tab shows the summary, in the same wording as the CLI: the players and fixtures, how many fixtures each
  player must play, how many players play an extra fixture, whether fixture spacing applied, the maximum play-together value
  in use, and the attempt number. Errors and notes also appear here.

If a run fails, the **Summary** tab opens by itself and shows the error. For example, "too many attempts, exiting. Tried
1000 of 1000" means the attempts limit was reached before a valid distribution was found. A cancelled run also opens the
Summary tab. A successful run opens the **Matrix** tab.

Changing **Transpose** redraws the matrix straight away. Changing any other input clears the matrix and disables **Save
CSV**, so you always save the result that matches the inputs.

![The matrix and summary for 24 fixtures, 7 players and 3 per fixture](images/desktop-results.png)

*Linux, v1.2.1, dark theme. The matrix has players as rows and fixtures as columns, and the summary is on the Summary tab.*

## Saving the result

**Save CSV** is enabled once a matrix has been generated. It opens a save dialog with the default name `distribution.csv`.
The file is described in [Saving the distribution as CSV](../README.md#saving-the-distribution-as-csv).

## Menu

The **Menu** in the window has these entries:

- **Theme:** opens a dialog with **Light**, **Dark** and **System**. **OK** applies the theme and remembers it for next time.
  **Cancel** changes nothing. The app starts in the light theme until you choose another.
- **About:** shows the version and the licence, with a link to the release page for that version.
- **Help:** opens the desktop app guide for the version you are running, in your default web browser. If the browser cannot be opened,
  an error dialog says so.
- **Quit:** closes the app.

The menu entries have no keyboard shortcuts.

![The Menu open, showing Theme, About, Help and Quit](images/desktop-menu.png)

*Linux, v1.2.1, dark theme.*

## Hover help

Rest the pointer on an input or its label for about one second, and a short explanation appears beside it. The explanation
disappears when the pointer leaves, or when you click the input.
