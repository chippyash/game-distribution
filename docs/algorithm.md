# How a distribution is built

The CLI and the desktop app both use the same algorithm, in `src/distributor/distributor.go`.

## Inputs

- Fixtures, players and players per fixture.
- Attempts: how many restarts are allowed before giving up.
- Maximum players together: the most fixtures any pair of players may share.
- Fixture spacing: whether players should avoid playing back-to-back fixtures.

## Shape of the result

The result is a matrix with one row per player and one column per fixture (or the reverse with transpose). Each cell is 1
if the player is in that fixture, and 0 otherwise.

Each player is given `floor(fixtures × players per fixture / players)` fixtures. The leftover places are given to some
players as one extra fixture.

## Choosing players

Players are chosen at random for each fixture. With fixture spacing on, a player who was in the previous fixture is
passed over when another player is available. The choice never loops forever: if no player is available, the attempt
fails and restarts.

## Restarts

If the matrix cannot be balanced, or the final check fails, the whole attempt restarts from scratch. The attempts limit
caps the number of restarts. Running out of attempts is reported as "too many attempts".

## Maximum players together

The cap limits how many fixtures any pair of players may share. The check counts pairs across the whole matrix after the
last fixture.

The cap is raised to a minimum automatically. The minimum is the average number of shared fixtures per pair, rounded up.
Meeting the minimum is not guaranteed. With fixture spacing on, the cap may need to be higher.

Worked example: 24 fixtures, 7 players, 3 per fixture. Each fixture of 3 players puts 3 pairs together, so there are 24 × 3 = 72 pair-appearances
over 21 pairs. The average is about 3.4, so the minimum is 4. With spacing on, 5 is needed: a cap of 4 fails, and a cap of
5 succeeds.

## Fixture spacing

Fixture spacing is a preference, not a guarantee. It is applied whenever it is requested, and it never fails an attempt.

Back-to-back fixtures can be avoided completely only when there are at least twice as many players as players per
fixture. When that is not the case, the output says so.
