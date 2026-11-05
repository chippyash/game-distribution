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
	"context"
	"strings"
	"testing"

	"github.com/chippyash/gamedistributor/src/desktop/model"
	"github.com/chippyash/gamedistributor/src/distributor"
	"gonum.org/v1/gonum/mat"
)

func run(t *testing.T, s model.Settings) *distributor.Distributor {
	t.Helper()
	d, err := distributor.NewDistributor(s.Fixtures, s.Players, s.PlayersPerFixture, s.Attempts, s.MaxPlayersTogether, s.Spacing).Distribute(context.Background())
	if err != nil {
		t.Fatalf("Distribute: %v", err)
	}
	return d
}

func TestSummaryMatchesCLIWording(t *testing.T) {
	s := validSettings()
	s.Attempts = 100000
	d := run(t, s)
	summary := model.Summary(s, d)
	for _, want := range []string{
		"Players: 7, Fixtures: 24, Players per fixture: 3",
		"Players must play: 10 fixtures",
		"2 Players will play an extra fixture",
		"Autospacing is false",
		"Max Play Together = 4 (requested 1 is below the minimum possible for these settings)",
	} {
		if !strings.Contains(summary, want) {
			t.Errorf("summary missing %q:\n%s", want, summary)
		}
	}
}

func TestSummaryNotesLimitedSpacing(t *testing.T) {
	s := model.DefaultSettings()
	s.Fixtures, s.Players, s.PlayersPerFixture = 9, 5, 3
	s.Spacing = true
	d := run(t, s)
	if !strings.Contains(model.Summary(s, d), "Autospacing is true (back-to-back fixtures cannot be fully avoided") {
		t.Errorf("limited spacing note missing:\n%s", model.Summary(s, d))
	}
}

func TestProgressAndCancelledMessages(t *testing.T) {
	if got := model.AttemptProgress(3, 1000); got != "Attempt 3 of 1000" {
		t.Errorf("AttemptProgress: got %q", got)
	}
	if got := model.CancelledMessage(42); got != "Cancelled after 42 attempts" {
		t.Errorf("CancelledMessage: got %q", got)
	}
}

func TestCellsWithoutTransposeHavePlayerRows(t *testing.T) {
	// 2 fixtures, 3 players: fixture 1 has players 1 and 3, fixture 2 has player 2
	m := mat.NewDense(2, 3, []float64{1, 0, 1, 0, 1, 0})
	rows, cols, cells := model.Cells(m, false)
	if len(rows) != 3 || rows[0] != "P1" || rows[2] != "P3" {
		t.Errorf("row labels %v, want players", rows)
	}
	if len(cols) != 2 || cols[0] != "F1" || cols[1] != "F2" {
		t.Errorf("column labels %v, want fixtures", cols)
	}
	want := [][]string{{"X", ""}, {"", "X"}, {"X", ""}}
	for r := range want {
		for c := range want[r] {
			if cells[r][c] != want[r][c] {
				t.Errorf("cell [%d][%d]: got %q, want %q", r, c, cells[r][c], want[r][c])
			}
		}
	}
}

func TestCellsWithTransposeHaveFixtureRows(t *testing.T) {
	m := mat.NewDense(2, 3, []float64{1, 0, 1, 0, 1, 0})
	rows, cols, cells := model.Cells(m, true)
	if len(rows) != 2 || rows[0] != "F1" {
		t.Errorf("row labels %v, want fixtures", rows)
	}
	if len(cols) != 3 || cols[2] != "P3" {
		t.Errorf("column labels %v, want players", cols)
	}
	if cells[0][0] != "X" || cells[0][1] != "" || cells[1][1] != "X" {
		t.Errorf("transposed cells %v", cells)
	}
}

func TestCSVMatchesCLILayout(t *testing.T) {
	m := mat.NewDense(2, 3, []float64{1, 0, 1, 0, 1, 0})
	got := model.CSV(m, false)
	want := [][]string{
		{"", "F1", "F2"},
		{"P1", "1", "0"},
		{"P2", "0", "1"},
		{"P3", "1", "0"},
	}
	compare(t, "CSV (players as rows)", got, want)

	gotT := model.CSV(m, true)
	wantT := [][]string{
		{"", "P1", "P2", "P3"},
		{"F1", "1", "0", "1"},
		{"F2", "0", "1", "0"},
	}
	compare(t, "CSV (fixtures as rows)", gotT, wantT)
}

func compare(t *testing.T, name string, got, want [][]string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s: got %d records, want %d", name, len(got), len(want))
	}
	for r := range want {
		if strings.Join(got[r], ",") != strings.Join(want[r], ",") {
			t.Errorf("%s row %d: got %v, want %v", name, r, got[r], want[r])
		}
	}
}
