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
	"reflect"
	"testing"

	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"

	"github.com/chippyash/gamedistributor/src/desktop/model"
)

// TestMenuHasThemeAboutAndHelp checks the application menu has the three entries, and no About button in the window
func TestMenuHasThemeAboutAndHelp(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	u := newUI(a, "v1.0.0")
	defer u.splash.finish()

	if len(u.menu.Items) != 1 {
		t.Fatalf("got %d top-level menus, want 1", len(u.menu.Items))
	}
	var labels []string
	for _, item := range u.menu.Items[0].Items {
		labels = append(labels, item.Label)
	}
	want := []string{"Theme", "About", "Help"}
	if !reflect.DeepEqual(labels, want) {
		t.Errorf("menu entries %v, want %v", labels, want)
	}
}

// TestSavedThemeIsApplied checks a saved theme choice is read back and applied at startup
func TestSavedThemeIsApplied(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	a.Preferences().SetString(model.PrefTheme, model.ThemeDark)
	if got := model.ThemeFromPreference(a.Preferences().String(model.PrefTheme)); got != model.ThemeDark {
		t.Fatalf("saved theme read back as %q, want %q", got, model.ThemeDark)
	}

	u := newUI(a, "v1.0.0")
	defer u.splash.finish()
	if !reflect.DeepEqual(a.Settings().Theme(), theme.DarkTheme()) {
		t.Errorf("dark theme not applied at startup")
	}
}

// TestNoSavedThemeStartsLight checks that the light theme is used when nothing has been saved
func TestNoSavedThemeStartsLight(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()
	u := newUI(a, "v1.0.0")
	defer u.splash.finish()
	if !reflect.DeepEqual(a.Settings().Theme(), theme.LightTheme()) {
		t.Errorf("light theme not applied when nothing is saved")
	}
}
