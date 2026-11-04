package distributor_test

import (
	"testing"

	"github.com/chippyash/gamedistributor/src/distributor"
	"gonum.org/v1/gonum/mat"
)

// consecutiveAppearances counts how often a player plays in two fixtures in a row
func consecutiveAppearances(m mat.Mutable) int {
	rows, cols := m.Dims()
	var n int
	for p := 0; p < cols; p++ {
		for f := 0; f < rows-1; f++ {
			if m.At(f, p) == 1 && m.At(f+1, p) == 1 {
				n++
			}
		}
	}
	return n
}

// meanConsecutive runs the distributor several times and returns the average consecutive count
func meanConsecutive(t *testing.T, runs, fixtures, players, perFixture, cap int, spacing bool) float64 {
	t.Helper()
	var total int
	for i := 0; i < runs; i++ {
		d := newDistribution(t, fixtures, players, perFixture, cap, spacing)
		total += consecutiveAppearances(d.Matrix())
	}
	return float64(total) / float64(runs)
}

func TestFixtureSpacingReducesBackToBackFixtures(t *testing.T) {
	// cap 5 is the lowest cap that completes with spacing on for 24 fixtures, 7 players, 3 per fixture
	const runs = 5
	off := meanConsecutive(t, runs, 24, 7, 3, 5, false)
	on := meanConsecutive(t, runs, 24, 7, 3, 5, true)
	if on*4 > off {
		t.Errorf("spacing on: %.1f back-to-back appearances on average, spacing off: %.1f; want spacing on to be well below", on, off)
	}
}

func TestFixtureSpacingDoesNotFailAttempts(t *testing.T) {
	// spacing is a preference: a feasible configuration must still complete with spacing on
	for i := 0; i < 5; i++ {
		d := newDistribution(t, 17, 5, 3, 6, true)
		checkDistribution(t, d, 17, 5, 3)
	}
}

func TestDistributionWithoutSpacingIsValid(t *testing.T) {
	// with spacing off the matrix is still valid; this guards the default path after the spacing refactor
	d := newDistribution(t, 24, 7, 3, 4, false)
	checkDistribution(t, d, 24, 7, 3)
}

func TestFixtureSpacingReportsEffectiveValue(t *testing.T) {
	cases := []struct {
		name                          string
		fixtures, players, perFixture int
		requested, want               bool
	}{
		{"requested", 24, 7, 3, true, true},
		{"not requested", 24, 7, 3, false, false},
		{"requested with fewer players than twice the players per fixture", 9, 5, 3, true, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := distributor.NewDistributor(tc.fixtures, tc.players, tc.perFixture, testAttempts, 1, tc.requested)
			if got := d.FixtureSpacing(); got != tc.want {
				t.Errorf("FixtureSpacing(): got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestCanFullySpace(t *testing.T) {
	cases := []struct {
		name                          string
		fixtures, players, perFixture int
		want                          bool
	}{
		{"exactly twice the players per fixture", 10, 6, 3, true},
		{"fewer players than twice the players per fixture", 9, 5, 3, false},
		{"one player per fixture", 5, 2, 1, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := distributor.NewDistributor(tc.fixtures, tc.players, tc.perFixture, testAttempts, 1, true)
			if got := d.CanFullySpace(); got != tc.want {
				t.Errorf("CanFullySpace(): got %v, want %v", got, tc.want)
			}
		})
	}
}
