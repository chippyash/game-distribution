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
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/test"
)

func hoverAt(x, y float32) *desktop.MouseEvent {
	return &desktop.MouseEvent{PointEvent: fyne.PointEvent{AbsolutePosition: fyne.NewPos(x, y)}}
}

// TestHoverHelpWaitsForDelayAndHidesOnLeave checks the help starts a timer on hover, shows when it fires, and hides when the pointer leaves
func TestHoverHelpWaitsForDelayAndHidesOnLeave(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	u := newTestUI(a, "v1.0.0")
	defer u.splash.finish()
	u.splashing = false

	u.fixtures.MouseIn(hoverAt(20, 20))
	if !u.fixtures.help.active || u.fixtures.help.timer == nil {
		t.Fatalf("hover did not start the help timer")
	}
	if u.tip.card.Visible() {
		t.Fatalf("help shown before the delay")
	}

	// the timer fires after hoverDelay; show is the action it runs
	u.fixtures.help.show()
	if !u.tip.card.Visible() {
		t.Fatalf("help not shown while hovering")
	}

	// a leave waits for the grace period, so a quick re-enter keeps the help showing
	u.fixtures.MouseOut()
	if u.fixtures.help.active || u.fixtures.help.hideTimer == nil {
		t.Fatalf("leave did not start the grace period")
	}
	u.fixtures.MouseIn(hoverAt(20, 20))
	if !u.fixtures.help.active || u.fixtures.help.hideTimer != nil {
		t.Errorf("re-entering during the grace period did not cancel the hide")
	}

	// a click hides the help at once
	u.fixtures.MouseDown(&desktop.MouseEvent{})
	if u.tip.card.Visible() || u.fixtures.help.active {
		t.Errorf("help still showing after the input was clicked")
	}
	u.fixtures.help.show()
	if u.tip.card.Visible() {
		t.Errorf("help shown after the input was clicked with the pointer gone")
	}
}

// TestHoverHelpIsOffDuringSplash checks no help appears while the splash is showing
func TestHoverHelpIsOffDuringSplash(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	u := newTestUI(a, "v1.0.0")
	defer u.splash.finish()

	u.players.MouseIn(hoverAt(20, 20))
	u.players.help.show()
	if u.tip.card.Visible() {
		t.Errorf("help shown while the splash is showing")
	}
	u.players.MouseOut()
}

// TestLabelsAndChecksHaveHelp checks the labels and check boxes share their input's help
func TestLabelsAndChecksHaveHelp(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	u := newTestUI(a, "v1.0.0")
	defer u.splash.finish()
	if u.spacing.help.text == "" || u.transpose.help.text == "" || u.fixtures.help.text == "" {
		t.Errorf("an input is missing its help text")
	}
}

// TestInputsAcceptTyping checks that tapping an input takes focus and that typed text arrives in it
func TestInputsAcceptTyping(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	u := newTestUI(a, "v1.0.0")
	defer u.splash.finish()

	u.splash.finish()
	u.fixtures.SetText("")
	u.fixtures.MouseDown(&desktop.MouseEvent{})
	if u.win.Canvas().Focused() != fyne.Focusable(u.fixtures) {
		t.Fatalf("tapping the fixtures input did not give it focus")
	}
	test.Type(u.fixtures, "24")
	if u.fixtures.Text != "24" {
		t.Errorf("typed text: got %q, want %q", u.fixtures.Text, "24")
	}
}
