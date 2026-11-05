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

func TestEveryInputHasHelpText(t *testing.T) {
	for _, field := range []string{
		model.FieldFixtures, model.FieldPlayers, model.FieldPlayersPerFixture,
		model.FieldMaxPlayersTogether, model.FieldAttempts, model.FieldSpacing, model.FieldTranspose,
	} {
		if model.InputHelp(field) == "" {
			t.Errorf("no help text for %s", field)
		}
	}
	if model.InputHelp("unknown") != "" {
		t.Errorf("unknown field has help text")
	}
}
