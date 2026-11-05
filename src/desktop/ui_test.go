// Copyright (c) 2026 Ashley Kitson, UK
//
// This software is licensed under the Functional Source License, Version 1.1 (FSL-1.1).
// You may use, copy, modify, and distribute this software for non-competing purposes.
//
// After two years from the release date of a specific version, that version
// automatically converts to the Apache License, Version 2.0.
//
// See the LICENSE file in the project root for full license terms

//go:build desktop

package main

import (
	"context"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"

	"github.com/chippyash/gamedistributor/src/distributor"
)

// TestGridAt100FixturesBy50Players checks the matrix grid at the target size, in both orientations
func TestGridAt100FixturesBy50Players(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	u := newUI(a, "test")
	defer u.splash.finish()

	u.fixtures.SetText("100")
	u.players.SetText("50")
	u.perFixture.SetText("2")
	d, err := distributor.NewDistributor(100, 50, 2, 100000, 1, false).Distribute(context.Background())
	if err != nil {
		t.Fatalf("Distribute: %v", err)
	}
	u.result = d

	// Players are rows by default: 50 players and a header row, 100 fixtures and a header column
	u.redraw()
	assertGrid(t, u, 51, 101)
	assertCellsMatchModel(t, u)
	assertEachFixtureHasTwoPlayers(t, u)

	u.transpose.Checked = true
	u.redraw()
	assertGrid(t, u, 101, 51)
	assertCellsMatchModel(t, u)

	// The grid must lay out at a window size smaller than the matrix, so it can scroll
	u.table.Resize(fyne.NewSize(800, 600))
	u.table.Refresh()
	if size := u.table.Size(); size.Width != 800 || size.Height != 600 {
		t.Errorf("table size %v, want 800x600", size)
	}
}

// assertGrid checks the number of rows and columns, including the header row and column
func assertGrid(t *testing.T, u *ui, rows, cols int) {
	t.Helper()
	gotRows, gotCols := u.tableSize()
	if gotRows != rows || gotCols != cols {
		t.Fatalf("grid size %dx%d, want %dx%d", gotRows, gotCols, rows, cols)
	}
}

// assertCellsMatchModel reads every displayed cell through the table callbacks and compares it with the model's cells
func assertCellsMatchModel(t *testing.T, u *ui) {
	t.Helper()
	rows, cols := u.tableSize()
	for r := 1; r < rows; r++ {
		for c := 1; c < cols; c++ {
			label := widget.NewLabel("")
			u.tableUpdate(widget.TableCellID{Row: r, Col: c}, label)
			if want := u.cells[r-1][c-1]; label.Text != want {
				t.Fatalf("cell (%d,%d) shows %q, want %q", r, c, label.Text, want)
			}
		}
	}
	// header labels: row labels down the first column and column labels along the first row
	label := widget.NewLabel("")
	u.tableUpdate(widget.TableCellID{Row: 1, Col: 0}, label)
	if label.Text != u.rowLabels[0] {
		t.Errorf("first row label %q, want %q", label.Text, u.rowLabels[0])
	}
	u.tableUpdate(widget.TableCellID{Row: 0, Col: cols - 1}, label)
	if label.Text != u.colLabels[cols-2] {
		t.Errorf("last column label %q, want %q", label.Text, u.colLabels[cols-2])
	}
}

// assertEachFixtureHasTwoPlayers checks the model: with players as rows, each fixture column has two X cells
func assertEachFixtureHasTwoPlayers(t *testing.T, u *ui) {
	t.Helper()
	for c := 0; c < len(u.colLabels); c++ {
		var n int
		for r := 0; r < len(u.rowLabels); r++ {
			if u.cells[r][c] == "X" {
				n++
			}
		}
		if n != 2 {
			t.Fatalf("fixture %s has %d players, want 2", u.colLabels[c], n)
		}
	}
}
