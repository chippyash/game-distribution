package main

import (
	"github.com/chippyash/gamedistributor/distributor"
)

func main() {
	numPlayers := 8
	numFixtures := 20
	playersPerFixture := 4
	attempts := 100
	fixtureSpacing := false
	maxPlayerTogether := 2

	_ = distributor.NewDistributor(numFixtures, numPlayers, playersPerFixture, attempts, maxPlayerTogether, fixtureSpacing).Distribute()

}
