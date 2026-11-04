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
