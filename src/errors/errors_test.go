// Copyright (c) 2026 Ashley Kitson, UK
//
// This software is licensed under the Functional Source License, Version 1.1 (FSL-1.1).
// You may use, copy, modify, and distribute this software for non-competing purposes.
//
// After two years from the release date of a specific version, that version
// automatically converts to the Apache License, Version 2.0.
//
// See the LICENSE file in the project root for full license terms

package errors_test

import (
	"testing"

	errs "github.com/chippyash/gamedistributor/src/errors"
)

func TestTooManyAttemptsMessage(t *testing.T) {
	err := errs.TooManyAttempts(3, 5)
	want := "too many attempts, exiting. Tried 3 of 5"
	if got := err.Error(); got != want {
		t.Errorf("TooManyAttempts(3, 5).Error(): got %q, want %q", got, want)
	}
}

func TestErrTooFewPlayersMessage(t *testing.T) {
	want := "Number of players must be greater than players per fixture"
	if got := errs.ErrTooFewPlayers.Error(); got != want {
		t.Errorf("ErrTooFewPlayers.Error(): got %q, want %q", got, want)
	}
}
