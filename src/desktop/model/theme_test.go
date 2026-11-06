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
	"testing"

	"github.com/chippyash/gamedistributor/src/desktop/model"
)

func TestThemeFromPreference(t *testing.T) {
	cases := map[string]string{
		"":       model.ThemeLight,
		"light":  model.ThemeLight,
		"dark":   model.ThemeDark,
		"system": model.ThemeSystem,
		"neon":   model.ThemeLight,
		"Dark":   model.ThemeLight,
	}
	for in, want := range cases {
		if got := model.ThemeFromPreference(in); got != want {
			t.Errorf("ThemeFromPreference(%q): got %q, want %q", in, got, want)
		}
	}
}

func TestHelpURLPinsVersion(t *testing.T) {
	want := "https://github.com/chippyash/game-distribution/blob/v1.0.0/docs/desktop-app.md"
	if got := model.HelpURL("v1.0.0"); got != want {
		t.Errorf("HelpURL: got %q, want %q", got, want)
	}
}

func TestReleaseURLPinsVersion(t *testing.T) {
	want := "https://github.com/chippyash/game-distribution/releases/tag/v1.0.0"
	if got := model.ReleaseURL("v1.0.0"); got != want {
		t.Errorf("ReleaseURL: got %q, want %q", got, want)
	}
}
