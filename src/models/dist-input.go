// Copyright (c) 2026 Ashley Kitson, UK
//
// This software is licensed under the Functional Source License, Version 1.1 (FSL-1.1).
// You may use, copy, modify, and distribute this software for non-competing purposes.
//
// After two years from the release date of a specific version, that version
// automatically converts to the Apache License, Version 2.0.
//
// See the LICENSE file in the project root for full license terms

package models

type GameDistInput struct {
	NumPlayers        int  `json:"numPlayers"`
	NumFixtures       int  `json:"numFixtures"`
	PlayersPerFixture int  `json:"playersPerFixture"`
	Attempts          int  `json:"attempts,omitempty"`
	FixtureSpacing    bool `json:"fixtureSpacing,omitempty"`
	MaxPlayerTogether int  `json:"maxPlayerTogether,omitempty"`
}

func NewGameDistInput() *GameDistInput {
	return &GameDistInput{
		NumPlayers:        0,
		NumFixtures:       0,
		PlayersPerFixture: 0,
		Attempts:          1000,
		FixtureSpacing:    false,
		MaxPlayerTogether: 0,
	}
}
