# Game Distributor
(c) Ashley Kitson 2024, UK

This utility will attempt to work out a random selection of players for a set of fixtures.  Although originally inspired
by the need to distribute an even number of fixtures amongst a pool of players in a team across a set of fixtures in Lawn Bowls,
there is no reason it cannot be used in other games: e.g.

 - junior football/rugby where you want to give the children/youths a fair crack at the game

Even where you may want to tweak the actual team selection, this utility will give you a good starting point.

It takes 3 main parameters:
 - number of fixtures
 - number of players in the team/pool
 - number of players required for each fixture

And provides you with a distribution matrix that you can optionally load into a spreadsheet to create your team selection sheet.

## Installation

### All platforms
Unzip the supplied gamedist.zip file to somewhere convenient

### Linux
Run the following commands
 - `echo $PATH`
 - look for a convenient path in which to install the utility. `/usr/local/bin`, `/home/yourname/bin` or `/home/yourname/.local/bin`
are good choices
 - move the `<somewhereconvenient>/gamedist` file to your chosen location. e.g. `mv <somewhereconvenient>/gamedist /usr/local/bin`
   - NB - you may need to prefix the last command with `sudo` depending on your target
 - type `gamedist help` to ensure you have installed correctly
 - remove the extract directory and zip file to clean up 
   - `rm gamedist.zip`
   - `rm -rf <somewhereconvenient>`  
   - Do take care that you are targeting correctly! Else use your graphical file manager
   - You may want to keep the readme file somewhere for later reference.

### Windows (64 bit)
 - create a new directory at `C:/Program Files/Gamedist`
 - move the `<somewhereconvient>/gamedist.exe` to `C:/Program Files/Gamedist`
 - goto:
   - -----> Control Panel-> System-> Advanced System Settings->Environment Variables
 - in the System Variables Window, locate `path` and select edit
 - Add your new path, `;C:/Program Files/Gamedist`
 - Open a Cmd terminal
 - Type `gamedist.exe help` to ensure you have installed correctly
 - remove the extract directory and zip file to clean up. You may want to keep the readme file somewhere for later reference.

## Usage:
 - For Linux the command is `gamedist`. For Windows it is `gamedist.exe`. Both commands are used in a terminal.
 - `gamedist help` displays help information
 - `gamedist 24 7 3 -a 10000` will display a permutation for 24 fixtures, 7 players and 3 players per fixture
 - `gamedist 24 7 3 -a 10000 -c` will do the same and produce a `distribution.csv` file that you can load into your spreadsheet application

### Usage notes:
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
