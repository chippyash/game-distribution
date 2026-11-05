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
	"time"

	"fyne.io/fyne/v2/test"
)

func TestSplashTimings(t *testing.T) {
	if splashHold != 2*time.Second {
		t.Errorf("splash holds for %v, want 2s", splashHold)
	}
	if splashFade != 3*time.Second {
		t.Errorf("splash fades over %v, want 3s", splashFade)
	}
}

// TestSplashBlocksInputUntilFaded checks the controls are locked while the splash shows, and released once it has faded
func TestSplashBlocksInputUntilFaded(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	u := newUI(a, "test")
	defer u.splash.finish()

	u.fixtures.SetText("24")
	u.players.SetText("7")
	u.perFixture.SetText("3")
	u.updateButtons()

	if !u.splashing {
		t.Fatalf("splash not showing at launch")
	}
	if !u.start.Disabled() {
		t.Errorf("start enabled while the splash is showing")
	}
	for _, in := range u.inputs {
		if !in.Disabled() {
			t.Errorf("an input is enabled while the splash is showing")
		}
	}

	u.splash.tick(0.5)
	if u.splash.img.Translucency != 0.5 {
		t.Errorf("translucency at half fade: got %v, want 0.5", u.splash.img.Translucency)
	}
	if !u.splashing {
		t.Errorf("splash removed before the fade finished")
	}

	u.splash.tick(1)
	if u.splashing {
		t.Errorf("splash still showing after the fade")
	}
	if u.start.Disabled() {
		t.Errorf("start still disabled after the splash")
	}
	for _, in := range u.inputs {
		if in.Disabled() {
			t.Errorf("an input is still disabled after the splash")
		}
	}
	if u.splash.layer.Visible() {
		t.Errorf("splash layer still visible after the fade")
	}
}

// TestSplashShowsVersionAndCopyright checks the caption under the splash image
func TestSplashShowsVersionAndCopyright(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	u := newUI(a, "v1.0.0")
	defer u.splash.finish()

	if got := u.splash.texts[0].Text; got != "Version v1.0.0" {
		t.Errorf("version caption %q, want %q", got, "Version v1.0.0")
	}
	if got := u.splash.texts[1].Text; got != "Copyright Ashley Kitson, UK 2026" {
		t.Errorf("copyright caption %q, want %q", got, "Copyright Ashley Kitson, UK 2026")
	}
}
