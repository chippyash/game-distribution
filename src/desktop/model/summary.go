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

import (
	"fmt"
	"strings"

	"github.com/chippyash/gamedistributor/src/distributor"
)

// Summary returns the run summary in the same wording as the CLI, one item per line
func Summary(s Settings, dist *distributor.Distributor) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Players: %d, Fixtures: %d, Players per fixture: %d\n", s.Players, s.Fixtures, s.PlayersPerFixture)
	fmt.Fprintf(&b, "Players must play: %d fixtures\n", dist.MustPlay())
	fmt.Fprintf(&b, "%d Players will play an extra fixture\n", dist.RemainderSlots())

	fmt.Fprintf(&b, "Autospacing is %t", dist.FixtureSpacing())
	if dist.FixtureSpacing() && !dist.CanFullySpace() {
		b.WriteString(" (back-to-back fixtures cannot be fully avoided: fewer than twice as many players as players per fixture)")
	}
	b.WriteString("\n")

	fmt.Fprintf(&b, "Max Play Together = %d", dist.MaxPlayersTogether())
	if dist.MaxPlayersTogether() != s.MaxPlayersTogether {
		fmt.Fprintf(&b, " (requested %d is below the minimum possible for these settings)", s.MaxPlayersTogether)
	}
	b.WriteString("\n")

	fmt.Fprintf(&b, "Attempt %d of %d", dist.Attempts()+1, s.Attempts)
	return b.String()
}

// AttemptProgress returns the progress text shown while a run is in progress
func AttemptProgress(made, maxAttempts int) string {
	return fmt.Sprintf("Attempt %d of %d", made, maxAttempts)
}

// CancelledMessage returns the message shown when a run is cancelled
func CancelledMessage(made int) string {
	return fmt.Sprintf("Cancelled after %d attempts", made)
}
