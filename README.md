# Game Distributor

This utility will attempt to work out a random selection of players for a set of fixtures.  Although originally inspired
by the need to distribute an even number of fixtures amongst a pool of players in a team across a set of fixtures in Lawn Bowls,
there is no reason it cannot be used in other games: e.g.

 - football/rugby where you want to give all the players a fair crack at the game
 - organising a tournament to set up the first round play-offs

You may want to tweak the actual team selection, this utility will give you a good starting point. Export the
result to a CSV file which can be imported using your spreadsheet program, Add player/team labels and fixture dates/names.
Manually move players around to get the shape you want. For instance, you may have couples that want to play together, or not!
You may know that certain players are on holiday, so you can swap them for another on that date.

It takes 3 main parameters:
 - number of fixtures
 - number of players in the team/pool
 - number of players required for each fixture

And provides you with a distribution matrix that you can optionally load into a spreadsheet to create your team selection sheet.

## Installation

### All platforms
Download the release archive (`gamedist-<version>.zip`) from the project's Releases page and extract it to somewhere
convenient, e.g. `unzip gamedist-<version>.zip -d <somewhereconvenient>`. On Windows, right-click the file and choose
Extract All. The archive contains the binaries for each platform and this README, so use the binaries for yours:
 - Linux: `gamedist` (CLI) and `gamedistapp` (desktop app)
 - Windows: `gamedist.exe` (CLI) and `gamedistapp.exe` (desktop app)

### Linux
Run the following commands
 - `echo $PATH`
 - look for a convenient path in which to install the utility. `/usr/local/bin`, `/home/yourname/bin` or `/home/yourname/.local/bin`
are good choices
 - move the `<somewhereconvenient>/gamedist` files to your chosen location. e.g. `mv <somewhereconvenient>/gamedist /usr/local/bin`
   - NB - you may need to prefix the last command with `sudo` depending on your target
 - type `gamedist help` to ensure you have installed correctly
 - to use the desktop app, move `<somewhereconvenient>/gamedistapp` to the same place and run `gamedistapp`. It needs OpenGL, Wayland and X11 libraries. On Debian and Ubuntu, install them with `sudo apt install libgl1 libwayland-client0 libx11-6`. The release build runs on Ubuntu 22.04 and later, and Debian 12 and later.
 - remove the extract directory and release archive to clean up 
   - `rm gamedist-<version>.zip`
   - `rm -rf <somewhereconvenient>`  
   - Do take care that you are targeting correctly! Else use your graphical file manager

### Windows (64 bit)
 - create a new directory at `C:\Program Files\Gamedist`
 - move the `<somewhereconvient>\gamedist.exe` to `C:\Program Files\Gamedist`
 - goto:
   - -----> Control Panel-> System-> Advanced System Settings->Environment Variables
 - in the System Variables Window, locate `path` and select edit
 - Add your new path, `;C:\Program Files\Gamedist`
 - Open a Cmd terminal
 - Type `gamedist.exe help` to ensure you have installed correctly
 - to use the desktop app, move `gamedistapp.exe` to the same folder and double-click it
 - remove the extract directory and release archive to clean up.

## Building from source

You need Go 1.22 or later and `make`. Run `make build` from the project root to build the CLI and the desktop app for
Linux and Windows. The desktop app needs cgo, and the Windows build is cross-compiled on Linux. The packages for each
platform, and what `make build` produces, are in [Building](https://github.com/chippyash/game-distribution/blob/master/docs/building.md).

Contributors should read [Contributing](https://github.com/chippyash/game-distribution/blob/master/CONTRIBUTING.md). Maintainers who publish releases should read
[Releasing](https://github.com/chippyash/game-distribution/blob/master/docs/releasing.md).

## Usage:
 - For Linux the command is `gamedist`. For Windows it is `gamedist.exe`. Both commands are used in a terminal.
 - `gamedist help` displays help information
 - `gamedist 24 7 3 -a 10000` will display a permutation for 24 fixtures, 7 players and 3 players per fixture
 - `gamedist 24 7 3 -a 10000 -c` will do the same and produce a `distribution.csv` file that you can load into your spreadsheet application

### Usage notes:

How a distribution is built, including attempts, the maximum players together and fixture spacing, is described in [How a distribution is built](https://github.com/chippyash/game-distribution/blob/master/docs/algorithm.md).

#### Number of attempts
The program tries to randomly allocate players across the fixtures as evenly as possible.  It will make, by default, 1000
attempts to do this.  If you consistently do not get an answer, you can increase the number of attempts with the -a flag, e.g.
`gamedist 24 7 3 -c -a 10000`. The max number of allowable attempts is 100000.
#### Max Play Together
The program will try as far as possible to ensure that no two players share more than `maxPlayersTogether` fixtures. By default
this is 1, so players won't play together more than once. If that is impossible, the program raises the limit to the
lowest value it could possibly meet: the average number of fixtures each pair of players shares, rounded up. Meeting that
minimum is not guaranteed, though. The program can still fail to find a permutation at the minimum, especially with
many players or with fixture spacing on, so you may need to raise the limit further. For example,
`gamedist 24 7 3 -a 10000` shares 72 pair-appearances across 21 pairs, so the limit is raised to 4 and the output tells you so.
If you consistently do not get an answer, you can increase the limit with the 4th parameter (maxPlayersTogether), e.g.
`gamedist 24 7 3 5 -c`. The limit can't be set below the minimum, so a lower value is simply raised to it. A failure at the minimum
means only that no permutation was found within the attempts, not that none exists.

A combination of the -a flag and the maxPlayersTogether parameter will work under most circumstance. If after trying out 
combinations of these, you still don't get a permutation, then the answer is that you have asked for something impossible
(for example, more fixtures than the player pool can cover without repeats that exceed the limit).

#### Fixture spacing
The -s flag spreads each player's fixtures out so that they are not in back-to-back fixtures. It is off by default.
Spacing is a preference, not a guarantee: the program avoids back-to-back fixtures where it can, but some may remain.
If there are fewer than twice as many players as players per fixture, back-to-back fixtures cannot be fully avoided, and the
output says so.

**Spacing can conflict with the play-together limit.** The automatic raise of `maxPlayersTogether` described above does not
take spacing into account, so with -s the raised limit may not be enough. For example, `gamedist 24 7 3 -s` raises the limit
to 4. At that limit the search is unreliable: in testing it failed in most runs even with 100000 attempts, and succeeded
in others. Setting the limit to 5 succeeded reliably: `gamedist 24 7 3 5 -s`. If you get `too many attempts` with -s, raise `maxPlayersTogether` manually, one step at a time,
until it works. The "Max Play Together" line in the output shows the limit that was used.

#### Transposition
By default the program outputs a permutation where rows are players and columns are fixtures. 
```
⎡1  0  1  0  0  1  0  0  0  1  1  0  0  0  0  1  0  0  0  0  1  1  1  1⎤
⎢0  1  0  1  1  0  1  0  1  0  0  1  0  0  1  1  0  0  0  1  1  0  0  1⎥
⎢0  0  1  0  1  0  1  0  0  0  0  0  1  0  0  0  1  1  1  1  0  1  1  1⎥
⎢0  0  1  0  0  1  0  1  1  0  1  0  0  1  1  0  1  1  0  0  1  0  0  0⎥
⎢1  1  0  1  0  0  1  0  0  1  1  0  1  1  1  1  0  0  0  0  0  0  0  0⎥
⎢0  0  0  1  1  0  0  1  1  1  0  1  1  0  0  0  1  0  1  0  0  1  0  0⎥
⎣1  1  0  0  0  1  0  1  0  0  0  1  0  1  0  0  0  1  1  1  0  0  1  0⎦
```
If your use case requires it you can add the -t flag to transpose the result. Rows are now fixtures and columns are 
players. e.g. `gamedist 24 7 3 -t`
```
⎡0  1  1  1  0  0  0⎤
⎢1  1  0  0  0  1  0⎥
⎢0  0  0  1  1  0  1⎥
⎢1  1  0  0  0  0  1⎥
⎢1  0  0  1  0  1  0⎥
⎢1  0  0  0  1  0  1⎥
⎢0  1  1  0  0  1  0⎥
⎢0  0  1  1  0  0  1⎥
⎢1  0  1  0  0  0  1⎥
⎢0  0  1  0  1  1  0⎥
⎢1  0  1  0  1  0  0⎥
⎢0  1  0  0  1  1  0⎥
⎢0  1  0  0  1  0  1⎥
⎢0  1  0  1  0  0  1⎥
⎢0  0  1  1  0  1  0⎥
⎢0  0  1  1  1  0  0⎥
⎢1  1  0  1  0  0  0⎥
⎢1  0  0  1  1  0  0⎥
⎢1  1  1  0  0  0  0⎥
⎢0  1  0  1  1  0  0⎥
⎢0  0  1  0  0  1  1⎥
⎢0  0  0  0  1  1  1⎥
⎢1  0  0  0  0  1  1⎥
⎣0  0  0  1  1  1  0⎦
```
## Desktop app

The desktop app does the same job as the CLI, in a window. It is `gamedistapp` on Linux and `gamedistapp.exe` on Windows,
and both are in the release archive alongside the CLI. Run `gamedistapp` (Linux) or double-click `gamedistapp.exe` (Windows).

The [desktop app guide](https://github.com/chippyash/game-distribution/blob/master/docs/desktop-app.md) covers the inputs, running and cancelling a distribution, reading and
saving the results, the menu and the hover help.

## Saving the distribution as CSV

The CLI's `-c` flag and the desktop app's Save CSV button both save the distribution as a CSV file, in the same format. 
The CLI writes `distribution.csv` to the current folder. In the desktop app, the Save CSV button is enabled once a 
distribution has been generated, and opens a save dialog with the default file name `distribution.csv`.

Without Transpose (`-t` in the CLI, or the Transpose option in the app), players are the rows and fixtures are the columns. 
With Transpose on, fixtures are the rows and players are the columns.

The file uses these conventions, which you can change in your spreadsheet:
 - `1` means the player is selected for that fixture, and `0` means not selected. Replace these with any symbols you like.
 - The headings are `P1` to `Pn` for players and `F1` to `Fn` for fixtures. Replace them with human-readable labels, such as names and dates.

## TODO

 - [ ] Find someone willing to build this app on the Mac platform, as the author does not have a Mac and never will.

## License
Copyright (c) 2026 Ashley Kitson, UK

This software is licensed under the Functional Source License, Version 1.1 (FSL-1.1).
You may use, copy, modify, and distribute this software for non-competing purposes.

After two years from the release date of a specific version, that version
automatically converts to the Apache License, Version 2.0.

See the LICENSE file in the project root for full license terms