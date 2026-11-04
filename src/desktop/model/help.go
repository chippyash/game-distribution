// Copyright (c) 2026 Ashley Kitson, UK
//
// This software is licensed under the Functional Source License, Version 1.1 (FSL-1.1).
// You may use, copy, modify, and distribute this software for non-competing purposes.
//
// After two years from the release date of a specific version, that version
// automatically converts to the Apache License, Version 2.0.
//
// See the LICENSE file in the project root for full license terms

package model

// Keys for the inputs that have hover help. The field keys for the numbers are the FieldErrors keys
const (
	FieldSpacing   = "spacing"
	FieldTranspose = "transpose"
)

var inputHelp = map[string]string{
	FieldFixtures:           "The number of fixtures to be played. Each fixture is one group of players playing together.",
	FieldPlayers:            "The number of players in the pool to be spread across the fixtures.",
	FieldPlayersPerFixture:  "How many players take part in each fixture. It cannot be more than the number of players.",
	FieldMaxPlayersTogether: "The most fixtures any two players may share. If the other settings make this impossible, it is raised automatically to the lowest value that can work, and the summary says so.",
	FieldAttempts:           "How many random attempts the program makes before it gives up. More attempts make a solution more likely, but take longer. Between 1000 and 100000.",
	FieldSpacing:            "Tries to stop players from playing in back-to-back fixtures. This is a preference, not a guarantee.",
	FieldTranspose:          "Swaps the rows and columns of the matrix and of the saved CSV, so fixtures are the rows and players are the columns.",
}

// InputHelp returns the hover help text for an input, or an empty string if it has none
func InputHelp(field string) string {
	return inputHelp[field]
}
