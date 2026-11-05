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

import "fmt"

// Limits on the inputs. Attempts match the CLI. The fixture and player limits are the 100 by 50 grid target,
// with headroom.
const (
	MinAttempts = 1000
	MaxAttempts = 100000
	MaxFixtures = 200
	MaxPlayers  = 100
)

// Field names used as keys in FieldErrors
const (
	FieldFixtures           = "fixtures"
	FieldPlayers            = "players"
	FieldPlayersPerFixture  = "playersPerFixture"
	FieldMaxPlayersTogether = "maxPlayersTogether"
	FieldAttempts           = "attempts"
)

// Settings are the user's inputs for one distribution
type Settings struct {
	Fixtures           int
	Players            int
	PlayersPerFixture  int
	MaxPlayersTogether int
	Attempts           int
	Spacing            bool
	Transpose          bool
}

// DefaultSettings returns the CLI defaults for the options. The counts are left at zero, so they must be entered.
func DefaultSettings() Settings {
	return Settings{MaxPlayersTogether: 1, Attempts: MinAttempts}
}

// FieldErrors maps a field name to a message explaining why its value is not allowed
type FieldErrors map[string]string

// Validate checks the settings against the limits. The returned map is empty when the settings are valid
func (s Settings) Validate() FieldErrors {
	errs := FieldErrors{}
	if s.Fixtures < 1 || s.Fixtures > MaxFixtures {
		errs[FieldFixtures] = fmt.Sprintf("must be between 1 and %d", MaxFixtures)
	}
	if s.Players < 1 || s.Players > MaxPlayers {
		errs[FieldPlayers] = fmt.Sprintf("must be between 1 and %d", MaxPlayers)
	}
	if s.PlayersPerFixture < 1 {
		errs[FieldPlayersPerFixture] = "must be at least 1"
	} else if s.Players >= 1 && s.PlayersPerFixture > s.Players {
		errs[FieldPlayersPerFixture] = "must not exceed players"
	}
	if s.MaxPlayersTogether < 1 {
		errs[FieldMaxPlayersTogether] = "must be at least 1"
	}
	if s.Attempts < MinAttempts || s.Attempts > MaxAttempts {
		errs[FieldAttempts] = fmt.Sprintf("must be between %d and %d", MinAttempts, MaxAttempts)
	}
	return errs
}
