// Copyright (c) 2026 Ashley Kitson, UK
//
// This software is licensed under the Functional Source License, Version 1.1 (FSL-1.1).
// You may use, copy, modify, and distribute this software for non-competing purposes.
//
// After two years from the release date of a specific version, that version
// automatically converts to the Apache License, Version 2.0.
//
// See the LICENSE file in the project root for full license terms

package distributor

import (
	"context"
	"fmt"
	errs "github.com/chippyash/gamedistributor/src/errors"
	"gonum.org/v1/gonum/mat"
	"math"
	"math/rand"
	"sync/atomic"
)

type Distributor struct {
	numFixtures       int          //the number of fixtures to be played
	numPlayers        int          //the number of players in the pool
	playersPerFixture int          //the number of players per fixture
	matrix            mat.Mutable  //the matrix to be populated
	playerSlots       int          //the number of playing slots available per distribution
	playerMustPlay    int          //the number of fixtures each player must play
	remainderSlots    int          //the number of slots remaining after the playerMustPlay
	attempts          int          //the number of times to try to distribute the fixtures
	debugCurrFixture  int          //the current fixture being processed
	debugCurrPlayer   int          //the current player being processed
	fixtureSpacing    bool         //if true, the fixtures will be spaced evenly for each player
	maxPlayerTogether int          //the maximum number of times that players to play together in all fixtures
	minPlayTogether   int          //the minimum number of times that the same players must play together
	minSameTeam       int          //the minimum number of times that the same team must play together
	attemptsMade      atomic.Int64 //the number of attempts started, readable while Distribute is running
}

func NewDistributor(numFixtures, numPlayers, playersPerFixture, attempts, maxPlayerTogether int, fixtureSpacing bool) *Distributor {
	ret := &Distributor{
		numFixtures:       numFixtures,
		numPlayers:        numPlayers,
		playersPerFixture: playersPerFixture,
		playerSlots:       numFixtures * playersPerFixture,
		playerMustPlay:    numFixtures * playersPerFixture / numPlayers,
		remainderSlots:    numFixtures * playersPerFixture % numPlayers,
		attempts:          attempts,
		fixtureSpacing:    fixtureSpacing,
		maxPlayerTogether: maxPlayerTogether,
		minPlayTogether:   minimumPlayTogether(numFixtures, numPlayers, playersPerFixture),
		minSameTeam:       int(math.Ceil(float64(numPlayers) / (float64(numFixtures) / float64(playersPerFixture)))),
	}

	// a cap below the average can never be satisfied, so raise it to the minimum
	if ret.maxPlayerTogether < ret.minPlayTogether {
		ret.maxPlayerTogether = ret.minPlayTogether
	}

	return ret
}

// minimumPlayTogether returns the lowest play-together cap that can be satisfied: the average number of times
// a pair of players shares a fixture, rounded up
func minimumPlayTogether(numFixtures, numPlayers, playersPerFixture int) int {
	pairs := numPlayers * (numPlayers - 1) / 2
	if pairs == 0 {
		return 0
	}
	pairAppearances := numFixtures * playersPerFixture * (playersPerFixture - 1) / 2
	return (pairAppearances + pairs - 1) / pairs
}

// MaxPlayersTogether returns the play-together cap in force, which is the requested cap raised to the minimum
// possible for the given settings if it was lower
func (d *Distributor) MaxPlayersTogether() int {
	return d.maxPlayerTogether
}

// FixtureSpacing returns whether fixture spacing was requested, so that the fixtures are spaced evenly for each player
func (d *Distributor) FixtureSpacing() bool {
	return d.fixtureSpacing
}

// CanFullySpace returns true if no player needs to play in consecutive fixtures to fill every fixture. This requires at
// least twice as many players as there are players per fixture. When it is false, spacing is still applied where possible
func (d *Distributor) CanFullySpace() bool {
	return d.numPlayers >= 2*d.playersPerFixture
}

// Matrix returns the matrix that will be populated with the distribution
// This should only be called after Distribute has been called
// or the results will be undefined
func (d *Distributor) Matrix() mat.Mutable {
	return d.matrix
}

// MustPlay returns the number of fixtures each player must play
// in order to be considered a valid distribution
// i.e. the minimum number of fixtures per player
func (d *Distributor) MustPlay() int {
	return d.playerMustPlay
}

// RemainderSlots returns the number of playing slots that are left over after giving every player
// and equal number of playing slots. These are used to give some players an additional playing slot
func (d *Distributor) RemainderSlots() int {
	return d.remainderSlots
}

// Attempts returns the number of attempts actually made to resolve the permutation
// AttemptsMade returns the number of attempts started so far. It is safe to call while Distribute is running
func (d *Distributor) AttemptsMade() int {
	return int(d.attemptsMade.Load())
}

func (d *Distributor) Attempts() int {
	return d.attempts
}

// Distribute will attempt to distribute the fixtures into the matrix
// It will return nil if it fails to find a valid distribution after the
// number of attempts specified in the constructor
// Otherwise it will return the distributor instance
// The run stops with the context's error if ctx is cancelled. A panic is returned as an error.
// This function is not thread safe, but AttemptsMade may be read while it runs
func (d *Distributor) Distribute(ctx context.Context) (dist *Distributor, err error) {
	defer func() {
		if r := recover(); r != nil {
			dist, err = nil, d.recoveredError(r)
		}
	}()

	var attempted int

	// simple validation
	if d.numPlayers < d.playersPerFixture {
		return nil, errs.ErrTooFewPlayers
	}

	//old skool goto jump
START:
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	d.attemptsMade.Add(1)
	if attempted >= d.attempts {
		return nil, errs.TooManyAttempts(attempted, d.attempts)
	}
	d.matrix = mat.NewDense(d.numFixtures, d.numPlayers, nil)
	// How many fixtures have each player been allocated to
	playerFixtures := make(map[int]int, d.numPlayers)
	//is player currently blocked from playing?
	// map[player]bool
	playerBlock := make(map[int]bool, d.numPlayers)
	for i := 0; i < d.numPlayers; i++ {
		playerFixtures[i] = 0
		playerBlock[i] = false
	}

	// for each fixture try to compute the players
	for currFixture := 0; currFixture < d.numFixtures; currFixture++ {
		d.debugCurrFixture = currFixture
		playerIndex := 0
		if d.fixtureSpacing {
			//block players who were selected in previous fixture
			if currFixture > 0 {
				for k := 0; k < d.numPlayers; k++ {
					playerBlock[k] = false
					p := d.matrix.At(currFixture-1, k)
					atLimit := playerFixtures[k] == d.playerMustPlay
					if p > 0 && !atLimit {
						playerBlock[k] = true
					}
				}
				//check to see that player blocking won't stop matrix from being fulfilled
				var playerBlockCount, playable int
				for k := 0; k < d.numPlayers; k++ {
					if playerBlock[k] && playerFixtures[k] < d.playerMustPlay {
						playerBlockCount++
					}
					if playerFixtures[k] < d.playerMustPlay {
						playable++
					}
				}
			}
		}
		for j := 0; j < d.playersPerFixture; j++ {
			var filled int
			filled = d.getFilled(playerFixtures)
			if filled == d.numPlayers && d.remainderSlots > 0 {
				d.playerMustPlay++
				filled = d.getFilled(playerFixtures)
			}
			currentPlayer := d.choosePlayer(currFixture, playerFixtures, playerBlock)
			d.debugCurrPlayer = currentPlayer
			if currentPlayer == -1 {
				//we can't allocate a player to this fixture
				//which means the matrix is out of balance, so try again
				attempted++
				//reset the mustPlay figure
				d.playerMustPlay = d.numFixtures * d.playersPerFixture / d.numPlayers
				goto START
			}
			d.matrix.Set(currFixture, currentPlayer, float64(1))
			playerIndex++
			if playerIndex >= d.playersPerFixture {
				break
			}
		}
		if currFixture == d.numFixtures-1 {
			//reset the mustPlay figure
			d.playerMustPlay = d.numFixtures * d.playersPerFixture / d.numPlayers
			//Check the maxPlayerTogether
			if d.checkPlayTogether() || d.checkMinPlays(playerFixtures) {
				attempted++
				goto START
			}
		}
	}

	// set attempted to the figure for the amount of attempts actually made
	d.attempts = attempted
	//reset the mustPlay figure
	d.playerMustPlay = d.numFixtures * d.playersPerFixture / d.numPlayers

	return d, nil
}

// checkPlayTogether checks to see if any pair of players has played together in more than
// maxPlayerTogether fixtures
func (d *Distributor) checkPlayTogether() bool {
	pairs := make(map[[2]int]int)
	for f := 0; f < d.numFixtures; f++ {
		fixture := mat.Row(nil, f, d.matrix)
		var players []int
		for i := 0; i < d.numPlayers; i++ {
			if fixture[i] == 1 {
				players = append(players, i)
			}
		}
		for i := 0; i < len(players); i++ {
			for j := i + 1; j < len(players); j++ {
				pair := [2]int{players[i], players[j]}
				pairs[pair]++
				if pairs[pair] > d.maxPlayerTogether {
					return true
				}
			}
		}
	}
	return false
}

// getFilled returns the number of players who have filled all their 'must play' slots
func (d *Distributor) getFilled(playerFixtures map[int]int) int {
	var filled int
	for _, v := range playerFixtures {
		if v == d.playerMustPlay {
			filled++
		}
	}
	return filled
}

// choosePlayer picks a random player for a fixture from those who still have must-play slots and are not already in
// the fixture. Players blocked by fixture spacing are avoided when possible, but are used rather than failing the
// attempt. It returns -1 if no player is available
func (d *Distributor) choosePlayer(currFixture int, playerFixtures map[int]int, playerBlock map[int]bool) int {
	var preferred, eligible []int
	for k := 0; k < d.numPlayers; k++ {
		if playerFixtures[k] >= d.playerMustPlay || d.isInFixture(k, currFixture) {
			continue
		}
		eligible = append(eligible, k)
		if !playerBlock[k] {
			preferred = append(preferred, k)
		}
	}
	pool := preferred
	if len(pool) == 0 {
		pool = eligible
	}
	if len(pool) == 0 {
		return -1
	}
	player := pool[rand.Intn(len(pool))]
	playerFixtures[player]++
	return player
}

// isInFixture returns true if a player is already in a fixture
func (d *Distributor) isInFixture(player, fixture int) bool {
	fix := mat.Row(nil, fixture, d.matrix)
	if fix[player] == 1 {
		return true
	}
	return false
}

// recoveredError converts a panic recovered by Distribute into an error that says where the run had reached
func (d *Distributor) recoveredError(r interface{}) error {
	return fmt.Errorf("unexpected failure at fixture %d, player %d: %v", d.debugCurrFixture, d.debugCurrPlayer, r)
}

// checkMinPlays checks that all players are playing the Must Play figure at least
func (d *Distributor) checkMinPlays(fixtures map[int]int) bool {
	for _, plays := range fixtures {
		if plays < d.playerMustPlay {
			return true
		}
	}
	return false
}
