// Copyright (c) 2026 Ashley Kitson, UK
//
// This software is licensed under the Functional Source License, Version 1.1 (FSL-1.1).
// You may use, copy, modify, and distribute this software for non-competing purposes.
//
// After two years from the release date of a specific version, that version
// automatically converts to the Apache License, Version 2.0.
//
// See the LICENSE file in the project root for full license terms

package distributor_test

import (
	"context"
	"errors"
	"testing"

	"github.com/chippyash/gamedistributor/src/distributor"
)

func TestDistributeReturnsContextErrorWhenCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	d, err := distributor.NewDistributor(24, 7, 3, testAttempts, 4, false).Distribute(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("got error %v, want %v", err, context.Canceled)
	}
	if d != nil {
		t.Errorf("cancelled run returned a distributor, want nil")
	}
}

func TestAttemptsMadeCountsEveryAttemptStarted(t *testing.T) {
	d, err := distributor.NewDistributor(24, 7, 3, testAttempts, 4, false).Distribute(context.Background())
	if err != nil {
		t.Fatalf("Distribute: unexpected error: %v", err)
	}
	// Attempts() counts the failed attempts; the successful attempt is one more
	if got, want := d.AttemptsMade(), d.Attempts()+1; got != want {
		t.Errorf("AttemptsMade: got %d, want %d", got, want)
	}
}

func TestDistributePanicReturnsError(t *testing.T) {
	// zero fixtures makes the matrix allocation panic inside Distribute
	d, err := distributor.NewDistributor(0, 3, 1, testAttempts, 1, false).Distribute(context.Background())
	if err == nil {
		t.Fatalf("got no error, want a recovered panic as an error")
	}
	if d != nil {
		t.Errorf("recovered panic returned a distributor, want nil")
	}
}
