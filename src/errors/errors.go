package errors

import "github.com/pkg/errors"

var tplTooManyAttempts = "too many attempts, exiting. Tried %d of %d"

func TooManyAttempts(attempts, maxAttempts int) error {
	return errors.Errorf(tplTooManyAttempts, attempts, maxAttempts)
}

var ErrTooFewPlayers = errors.New("Number of players must be greater than players per fixture")
