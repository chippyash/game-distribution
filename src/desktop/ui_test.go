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
	"errors"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/widget"

	"github.com/chippyash/gamedistributor/src/desktop/model"
	"github.com/chippyash/gamedistributor/src/distributor"
)

// TestGridAt100FixturesBy50Players checks the matrix grid at the target size, in both orientations
func TestGridAt100FixturesBy50Players(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	u := newTestUI(a, "test")
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

// TestEnterButtonRunsOnEnterOnlyWhenEnabled checks that Enter taps the Start button, and that a disabled button ignores it
func TestEnterButtonRunsOnEnterOnlyWhenEnabled(t *testing.T) {
	taps := 0
	b := newEnterButton("Start", func() { taps++ })
	enter := &fyne.KeyEvent{Name: fyne.KeyReturn}

	b.TypedKey(enter)
	if taps != 1 {
		t.Fatalf("Enter on an enabled button: %d taps, want 1", taps)
	}
	b.Disable()
	b.TypedKey(enter)
	if taps != 1 {
		t.Errorf("Enter on a disabled button: %d taps, want 1", taps)
	}
}

// maxContentHeight is the tallest the window content may be, in logical pixels. It leaves room for the title bar and
// menu on a 1080 pixel screen at 150% scaling
const maxContentHeight = 540

// TestWindowFitsSmallScreen checks the window content is no taller than a 150% scaled 1080 pixel screen allows
func TestWindowFitsSmallScreen(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	u := newTestUI(a, "test")
	defer u.splash.finish()

	if h := u.win.Content().MinSize().Height; h > maxContentHeight {
		t.Errorf("window content is %v tall, want at most %v", h, maxContentHeight)
	}
}

// TestStartRunsFromEnterInWindow checks that Start takes the focus in the window and runs on Enter
func TestStartRunsFromEnterInWindow(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	u := newTestUI(a, "test")
	defer u.splash.finish()
	u.splash.finish()

	u.fixtures.SetText("24")
	u.players.SetText("7")
	u.perFixture.SetText("3")
	u.attempts.SetText("5000")
	u.updateButtons()

	c := u.win.Canvas()
	c.Focus(u.start)
	if c.Focused() != u.start {
		t.Fatalf("Start did not take the focus")
	}
	c.Focused().TypedKey(&fyne.KeyEvent{Name: fyne.KeyReturn})
	if !u.running {
		t.Fatalf("Enter on the focused Start did not start a run")
	}

	u.cancelRun()
	for i := 0; u.running && i < 500; i++ {
		time.Sleep(10 * time.Millisecond)
	}
}

// TestFailedRunShowsSummaryTab checks that a run which fails, or is cancelled, moves to the Summary tab, where its message is
func TestFailedRunShowsSummaryTab(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	u := newTestUI(a, "test")
	defer u.splash.finish()
	u.splash.finish()

	u.finish(model.Settings{}, nil, nil, errors.New("too many attempts, exiting. Tried 1000 of 1000"))
	if u.tabs.Selected() != u.summary {
		t.Errorf("after an error the %q tab is selected, want Summary", u.tabs.Selected().Text)
	}
	if got := u.panel.Text; got != "too many attempts, exiting. Tried 1000 of 1000" {
		t.Errorf("summary text %q", got)
	}

	u.tabs.Select(u.summary)
	d := distributor.NewDistributor(24, 7, 3, 1000, 1, false)
	u.finish(model.Settings{}, d, nil, context.Canceled)
	if u.tabs.Selected() != u.summary {
		t.Errorf("after a cancel the %q tab is selected, want Summary", u.tabs.Selected().Text)
	}
}

// TestSuccessfulRunShowsMatrixTab checks that a run which produces a matrix moves to the Matrix tab
func TestSuccessfulRunShowsMatrixTab(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	u := newTestUI(a, "test")
	defer u.splash.finish()
	u.splash.finish()

	s := model.Settings{Fixtures: 24, Players: 7, PlayersPerFixture: 3, Attempts: 5000, MaxPlayersTogether: 1}
	d := distributor.NewDistributor(s.Fixtures, s.Players, s.PlayersPerFixture, s.Attempts, s.MaxPlayersTogether, s.Spacing)
	res, err := d.Distribute(context.Background())
	if err != nil {
		t.Fatalf("Distribute: %v", err)
	}

	u.tabs.Select(u.summary)
	u.finish(s, d, res, nil)
	if u.tabs.Selected() != u.matrix {
		t.Errorf("after a successful run the %q tab is selected, want Matrix", u.tabs.Selected().Text)
	}
}
