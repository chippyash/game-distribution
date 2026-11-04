// Copyright (c) 2026 Ashley Kitson, UK
//
// This software is licensed under the Functional Source License, Version 1.1 (FSL-1.1).
// You may use, copy, modify, and distribute this software for non-competing purposes.
//
// After two years from the release date of a specific version, that version
// automatically converts to the Apache License, Version 2.0.
//
// See the LICENSE file in the project root for full license terms

package model_test

import (
	"testing"

	"github.com/chippyash/gamedistributor/src/desktop/model"
)

func validSettings() model.Settings {
	s := model.DefaultSettings()
	s.Fixtures, s.Players, s.PlayersPerFixture = 24, 7, 3
	return s
}

func TestValidSettingsHaveNoErrors(t *testing.T) {
	if errs := validSettings().Validate(); len(errs) != 0 {
		t.Errorf("got errors %v, want none", errs)
	}
}

func TestDefaultSettingsMatchCLI(t *testing.T) {
	s := model.DefaultSettings()
	if s.Attempts != 1000 || s.MaxPlayersTogether != 1 || s.Spacing || s.Transpose {
		t.Errorf("defaults %+v do not match the CLI", s)
	}
}

func TestValidateRejectsOutOfRangeValues(t *testing.T) {
	cases := []struct {
		name   string
		change func(*model.Settings)
		field  string
	}{
		{"zero fixtures", func(s *model.Settings) { s.Fixtures = 0 }, model.FieldFixtures},
		{"too many fixtures", func(s *model.Settings) { s.Fixtures = model.MaxFixtures + 1 }, model.FieldFixtures},
		{"zero players", func(s *model.Settings) { s.Players = 0 }, model.FieldPlayers},
		{"too many players", func(s *model.Settings) { s.Players = model.MaxPlayers + 1 }, model.FieldPlayers},
		{"zero players per fixture", func(s *model.Settings) { s.PlayersPerFixture = 0 }, model.FieldPlayersPerFixture},
		{"players per fixture above players", func(s *model.Settings) { s.PlayersPerFixture = 8 }, model.FieldPlayersPerFixture},
		{"zero max players together", func(s *model.Settings) { s.MaxPlayersTogether = 0 }, model.FieldMaxPlayersTogether},
		{"attempts below minimum", func(s *model.Settings) { s.Attempts = model.MinAttempts - 1 }, model.FieldAttempts},
		{"attempts above maximum", func(s *model.Settings) { s.Attempts = model.MaxAttempts + 1 }, model.FieldAttempts},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := validSettings()
			tc.change(&s)
			errs := s.Validate()
			if _, ok := errs[tc.field]; !ok {
				t.Errorf("no error for %s; got %v", tc.field, errs)
			}
		})
	}
}

func TestValidateAcceptsBoundaryValues(t *testing.T) {
	s := validSettings()
	s.Fixtures, s.Players, s.PlayersPerFixture = model.MaxFixtures, model.MaxPlayers, model.MaxPlayers
	s.Attempts = model.MaxAttempts
	if errs := s.Validate(); len(errs) != 0 {
		t.Errorf("boundary values rejected: %v", errs)
	}
	s.Attempts = model.MinAttempts
	s.Fixtures, s.Players, s.PlayersPerFixture = 1, 1, 1
	if errs := s.Validate(); len(errs) != 0 {
		t.Errorf("minimum values rejected: %v", errs)
	}
}
