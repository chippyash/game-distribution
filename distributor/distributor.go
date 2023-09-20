package distributor

import (
	"fmt"
	"github.com/kelindar/bitmap"
	"gonum.org/v1/gonum/mat"
	"math/rand"
	"runtime/debug"
)

type Distributor struct {
	numFixtures       int         //the number of fixtures to be played
	numPlayers        int         //the number of players in the pool
	playersPerFixture int         //the number of players per fixture
	matrix            mat.Mutable //the matrix to be populated
	playerSlots       int         //the number of playing slots available per distribution
	playerMustPlay    int         //the number of fixtures each player must play
	remainderSlots    int         //the number of slots remaining after the playerMustPlay
	attempts          int         //the number of times to try to distribute the fixtures
	debugCurrFixture  int         //the current fixture being processed
	debugCurrPlayer   int         //the current player being processed
	fixtureSpacing    bool        //if true, the fixtures will be spaced evenly for each player
	maxPlayerTogether int         //the maximum number of times that players to play together in all fixtures
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
		fixtureSpacing:    fixtureSpacing && numFixtures/numPlayers > 1,
		maxPlayerTogether: maxPlayerTogether,
	}

	return ret
}

func (d *Distributor) Distribute() mat.Matrix {
	defer d.DoRecovery()
	var allreadyPrinted bool
	var attempted int
START:
	if attempted >= d.attempts {
		fmt.Printf("Too many attempts, exiting. Tried %d of %d\n", attempted, d.attempts)
		return nil
	}
	d.matrix = mat.NewDense(d.numFixtures, d.numPlayers, nil)
	// How many fixtures have each player been allocated to
	// map[player]numFixtures
	playerFixtures := make(map[int]int, d.numPlayers)
	//is player currently blocked from playing?
	// map[player]bool
	playerBlock := make(map[int]bool, d.numPlayers)
	for i := 0; i < d.numPlayers; i++ {
		playerFixtures[i] = 0
		playerBlock[i] = false
	}

	// Print out the game information
	if !allreadyPrinted {
		d.printHeader()
		//validation
		if d.numPlayers < d.playersPerFixture {
			fmt.Println("Number of players must be greater than players per fixture")
			return nil
		}
		allreadyPrinted = true
	}

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
				if playable-playerBlockCount < d.playersPerFixture {
					//cancel the block
					d.fixtureSpacing = false
					playerBlock = make(map[int]bool, d.numPlayers)
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
			var currentPlayer int
			if d.cantFill(filled) {
				currentPlayer = d.pickPlayer(currFixture, playerFixtures, playerBlock)
			} else {
				currentPlayer = d.genRandPlayer(currFixture, playerFixtures, playerBlock)
			}
			d.debugCurrPlayer = currentPlayer
			if currentPlayer == -1 {
				//we can't allocate a player to this fixture
				//which means the matrix is out of balance, so try again
				attempted++
				goto START
			}
			d.matrix.Set(currFixture, currentPlayer, float64(1))
			playerIndex++
			if playerIndex >= d.playersPerFixture {
				break
			}
		}
		if currFixture == d.numFixtures-1 {
			//Check the maxPlayerTogether
			if d.maxPlayerTogether > 0 {
				//playedTogether map[fixture]bimapOfPlayers
				playedTogether := make(map[int][]bitmap.Bitmap, d.numPlayers)
				for f := 0; f < d.numFixtures; f++ {
					fixture := mat.Row(nil, f, d.matrix)
					var together bitmap.Bitmap
					for i := 0; i < d.numPlayers; i++ {
						if fixture[i] == 1 {
							together.Set(uint32(i))
						}
					}
					playedTogether[f] = append(playedTogether[f], together)
				}
				//check to see if we have fixtures with the same players
				samePlayers := make(map[int]uint16)
				for f := 0; f < d.numFixtures; f++ {
					samePlayers[uint64(playedTogether[f][0])]++
				}
				fmt.Printf("played = %v\n", playedTogether)
			}

			d.Print(fmt.Sprintf("After attempt: %d", attempted+1))
		}
	}

	t := mat.Transpose{Matrix: d.matrix}
	return t
}

func (d *Distributor) printHeader() {
	var plural string
	fmt.Printf("Players: %d, Fixtures: %d, Players per fixture: %d\n", d.numPlayers, d.numFixtures, d.playersPerFixture)
	if d.playerSlots/d.numPlayers > 1 {
		plural = "s"
	}
	fmt.Printf("Players must play: %d fixture%s\n", d.playerSlots/d.numPlayers, plural)
	plural = ""
	if d.remainderSlots > 1 {
		plural = "s"
	}
	fmt.Printf("%d Player%s can play an extra fixture\n", d.remainderSlots, plural)
	if d.fixtureSpacing {
		fmt.Println("Autospacing is active")
	}
	noPlayPlayers := d.numPlayers - d.playerSlots
	if noPlayPlayers > 0 {
		plural = ""
		if noPlayPlayers > 1 {
			plural = "s"
		}
		fmt.Printf("Warning: %d player%s will not play\n", noPlayPlayers, plural)
	}
}

// cantFill returns true if iteration cannot fill the matrix with players
func (d *Distributor) cantFill(filled int) bool {
	return d.numPlayers-filled < d.playersPerFixture
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

func (d *Distributor) pickPlayer(currFixture int, playerFixtures map[int]int, playerBlock map[int]bool) int {
	for k, played := range playerFixtures {
		if played < d.playerMustPlay && !d.isInFixture(k, currFixture) && !playerBlock[k] {
			playerFixtures[k]++
			return k
		}
	}
	return -1
}

// genRandPlayer generates a random player for a given number of players per fixture taking into account the number of
// fixtures each player has been allocated to
func (d *Distributor) genRandPlayer(currFixture int, playerFixtures map[int]int, playerBlock map[int]bool) int {
	for {
		player := rand.Intn(d.numPlayers)
		if playerFixtures[player] < d.playerMustPlay && !d.isInFixture(player, currFixture) && !playerBlock[player] {
			playerFixtures[player]++
			return player
		}
	}
}

func (d *Distributor) isInFixture(player, fixture int) bool {
	fix := mat.Row(nil, fixture, d.matrix)
	if fix[player] == 1 {
		return true
	}
	return false
}

func (d *Distributor) Print(footer string) {
	t := mat.Transpose{Matrix: d.matrix}
	fa := mat.Formatted(t, mat.Prefix("         "), mat.Squeeze())
	println("Columns are fixtures, Rows are Players")
	fmt.Printf("Result = %v\n", fa)
	println(footer, "\n")
}

func (d *Distributor) DoRecovery() {
	if r := recover(); r != nil {
		if err, ok := r.(error); ok {
			//recover returned some error, so log it
			fmt.Printf("Unexpected error: %s\n", err.Error())
		} else {
			//recover returned something else
			fmt.Printf("Unexpected exception: %v\n", r)
		}
		fmt.Printf("Debug: Current Fixture: %d, Current Player: %d\n", d.debugCurrFixture, d.debugCurrPlayer)
		fmt.Printf("Stack Trace: %s", string(debug.Stack()))
	}
}
