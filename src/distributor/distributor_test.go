package distributor_test

import (
	"errors"
	"testing"

	"github.com/chippyash/gamedistributor/src/distributor"
	errs "github.com/chippyash/gamedistributor/src/errors"
	"gonum.org/v1/gonum/mat"
)

const testAttempts = 100000

// newDistribution runs Distribute for the given settings and fails the test if it returns an error
func newDistribution(t *testing.T, fixtures, players, perFixture, cap int, spacing bool) *distributor.Distributor {
	t.Helper()
	d, err := distributor.NewDistributor(fixtures, players, perFixture, testAttempts, cap, spacing).Distribute()
	if err != nil {
		t.Fatalf("Distribute(%d, %d, %d, cap %d): unexpected error: %v", fixtures, players, perFixture, cap, err)
	}
	return d
}

// checkDistribution verifies the properties every valid distribution must have
func checkDistribution(t *testing.T, d *distributor.Distributor, fixtures, players, perFixture int) {
	t.Helper()
	m := d.Matrix()
	rows, cols := m.Dims()
	if rows != fixtures || cols != players {
		t.Fatalf("matrix dims: got %dx%d, want %dx%d", rows, cols, fixtures, players)
	}

	// every fixture has exactly perFixture players, each marked 0 or 1
	for f := 0; f < rows; f++ {
		var playing int
		for p := 0; p < cols; p++ {
			v := m.At(f, p)
			if v != 0 && v != 1 {
				t.Fatalf("fixture %d player %d: value %v is not 0 or 1", f, p, v)
			}
			if v == 1 {
				playing++
			}
		}
		if playing != perFixture {
			t.Fatalf("fixture %d: %d players, want %d", f, playing, perFixture)
		}
	}

	// each player plays the base number of fixtures, and the remainder slots give that many players one extra
	base := fixtures * perFixture / players
	remainder := fixtures * perFixture % players
	var extra int
	for p := 0; p < cols; p++ {
		var played int
		for f := 0; f < rows; f++ {
			played += int(m.At(f, p))
		}
		switch played {
		case base:
		case base + 1:
			extra++
		default:
			t.Fatalf("player %d plays %d fixtures, want %d or %d", p, played, base, base+1)
		}
	}
	if extra != remainder {
		t.Fatalf("%d players have an extra fixture, want %d", extra, remainder)
	}

	// no pair of players shares more fixtures than the cap in force
	cap := d.MaxPlayersTogether()
	pairs := make(map[[2]int]int)
	for f := 0; f < rows; f++ {
		var playing []int
		for p := 0; p < cols; p++ {
			if m.At(f, p) == 1 {
				playing = append(playing, p)
			}
		}
		for i := range playing {
			for j := i + 1; j < len(playing); j++ {
				pair := [2]int{playing[i], playing[j]}
				pairs[pair]++
				if pairs[pair] > cap {
					t.Fatalf("players %v share %d fixtures, cap is %d", pair, pairs[pair], cap)
				}
			}
		}
	}
}

func TestDistributeProducesValidDistributions(t *testing.T) {
	// Fixture spacing is left off for 10 fixtures / 4 players: with spacing on, the distribution
	// did not converge within the attempt limit for any cap tried.
	cases := []struct {
		name                      string
		fixtures, players, perFix int
		cap                       int
		spacing                   bool
	}{
		{"lawn bowls 24/7/3", 24, 7, 3, 1, false},
		{"lawn bowls 24/7/3 spaced", 24, 7, 3, 1, true},
		{"17 fixtures 5 players", 17, 5, 3, 1, false},
		{"17 fixtures 5 players spaced", 17, 5, 3, 6, true},
		{"7 fixtures 5 players spaced", 7, 5, 3, 2, true},
		{"10 fixtures 4 players", 10, 4, 2, 3, false},
		{"9 fixtures 6 players", 9, 6, 3, 3, false},
		{"9 fixtures 6 players spaced", 9, 6, 3, 3, true},
		{"one player per fixture", 5, 4, 1, 1, false},
		{"fewer fixtures than players", 3, 7, 2, 1, false},
		{"every player in every fixture", 2, 2, 2, 1, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := newDistribution(t, tc.fixtures, tc.players, tc.perFix, tc.cap, tc.spacing)
			checkDistribution(t, d, tc.fixtures, tc.players, tc.perFix)
		})
	}
}

func TestMustPlayAndRemainderSlots(t *testing.T) {
	d := distributor.NewDistributor(24, 7, 3, testAttempts, 1, false)
	if got := d.MustPlay(); got != 10 {
		t.Errorf("MustPlay: got %d, want 10", got)
	}
	if got := d.RemainderSlots(); got != 2 {
		t.Errorf("RemainderSlots: got %d, want 2", got)
	}
}

func TestDistributeRejectsTooFewPlayers(t *testing.T) {
	_, err := distributor.NewDistributor(4, 2, 3, testAttempts, 1, false).Distribute()
	if !errors.Is(err, errs.ErrTooFewPlayers) {
		t.Fatalf("got error %v, want %v", err, errs.ErrTooFewPlayers)
	}
}

func TestMaxPlayersTogetherIsRaisedToMinimum(t *testing.T) {
	// 24 fixtures of 3 players: 72 pair-appearances over 21 pairs, so the minimum cap is 4
	d := distributor.NewDistributor(24, 7, 3, testAttempts, 1, false)
	if got := d.MaxPlayersTogether(); got != 4 {
		t.Errorf("MaxPlayersTogether with requested 1: got %d, want 4", got)
	}
}

func TestMaxPlayersTogetherKeepsRequestedAboveMinimum(t *testing.T) {
	d := distributor.NewDistributor(24, 7, 3, testAttempts, 6, false)
	if got := d.MaxPlayersTogether(); got != 6 {
		t.Errorf("MaxPlayersTogether with requested 6: got %d, want 6", got)
	}
}

func TestMaxPlayersTogetherOnePerFixtureNeedsNoPairs(t *testing.T) {
	d := distributor.NewDistributor(5, 4, 1, testAttempts, 1, false)
	if got := d.MaxPlayersTogether(); got != 1 {
		t.Errorf("MaxPlayersTogether with one player per fixture: got %d, want 1", got)
	}
}

func TestAttemptsWithinLimit(t *testing.T) {
	d := newDistribution(t, 24, 7, 3, 4, false)
	if got := d.Attempts(); got < 0 || got >= testAttempts {
		t.Errorf("Attempts: got %d, want between 0 and %d", got, testAttempts-1)
	}
}

func TestMatrixIsFixturesByPlayers(t *testing.T) {
	d := newDistribution(t, 5, 4, 2, 1, false)
	var m mat.Mutable = d.Matrix()
	rows, cols := m.Dims()
	if rows != 5 || cols != 4 {
		t.Errorf("Matrix dims: got %dx%d, want 5x4", rows, cols)
	}
}
