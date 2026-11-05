// Copyright (c) 2026 Ashley Kitson, UK
//
// This software is licensed under the Functional Source License, Version 1.1 (FSL-1.1).
// You may use, copy, modify, and distribute this software for non-competing purposes.
//
// After two years from the release date of a specific version, that version
// automatically converts to the Apache License, Version 2.0.
//
// See the LICENSE file in the project root for full license terms

package model

// Theme choices, as stored in the preferences. The user sees the labels Light, Dark and System
const (
	ThemeLight  = "light"
	ThemeDark   = "dark"
	ThemeSystem = "system"

	// PrefTheme is the preferences key for the theme choice
	PrefTheme = "theme"
)

// ThemeFromPreference returns the theme to use for a stored preference. Anything unknown, including an empty value,
// is the light theme
func ThemeFromPreference(v string) string {
	switch v {
	case ThemeDark, ThemeSystem:
		return v
	default:
		return ThemeLight
	}
}
