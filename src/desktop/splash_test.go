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

	"fyne.io/fyne/v2"
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

// TestSplashShowsVersionAndCopyright checks the one-line caption under the splash image
func TestSplashShowsVersionAndCopyright(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	u := newUI(a, "v1.0.0")
	defer u.splash.finish()

	want := "Version v1.0.0    Copyright Ashley Kitson, UK 2026"
	if got := u.splash.caption.Text; got != want {
		t.Errorf("caption %q, want %q", got, want)
	}
}

// TestSplashCaptionIsBelowImage checks that the image and the caption never overlap, at heights where the image is limited
// by the height of the window
func TestSplashCaptionIsBelowImage(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	s := newSplashScreen("v1.0.0", func() {})

	for _, size := range []fyne.Size{fyne.NewSize(1000, 300), fyne.NewSize(1000, 500), fyne.NewSize(1000, 700)} {
		s.layer.Resize(size)
		imgBottom := s.img.Position().Y + s.img.Size().Height
		if imgBottom > s.caption.Position().Y {
			t.Errorf("at %v the image ends at %v, below the caption top at %v", size, imgBottom, s.caption.Position().Y)
		}
		if bottom := s.caption.Position().Y + s.caption.Size().Height; bottom > size.Height {
			t.Errorf("at %v the caption ends at %v, below the window", size, bottom)
		}
	}
}
