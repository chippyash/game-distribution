// Copyright (c) 2026 Ashley Kitson, UK
//
// This software is licensed under the Functional Source License, Version 1.1 (FSL-1.1).
// You may use, copy, modify, and distribute this software for non-competing purposes.
//
// After two years from the release date of a specific version, that version
// automatically converts to the Apache License, Version 2.0.
//
// See the LICENSE file in the project root for full license terms

package errors

import "github.com/pkg/errors"

var tplTooManyAttempts = "too many attempts, exiting. Tried %d of %d"

func TooManyAttempts(attempts, maxAttempts int) error {
	return errors.Errorf(tplTooManyAttempts, attempts, maxAttempts)
}

var ErrTooFewPlayers = errors.New("Number of players must be greater than players per fixture")
