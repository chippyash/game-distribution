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

type DistOutput struct {
	NumFixtures        int     `json:"numFixtures"`
	NumPlayers         int     `json:"numPlayers"`
	PlayersPerFixture  int     `json:"playersPerFixture"`
	MaxPlayersTogether int     `json:"maxPlayersTogether"`
	FixtureSpacing     bool    `json:"fixtureSpacing"`
	Attempts           int     `json:"attempts"`
	MaxAttempts        int     `json:"maxAttempts"`
	PlayersMustPlay    int     `json:"playersMustPlay"`
	RemainderSlots     int     `json:"remainderSlots"`
	Distribution       [][]int `json:"distribution"`
}
