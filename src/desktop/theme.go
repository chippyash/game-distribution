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
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// variantTheme is the default theme with its colours fixed to one variant, so the app can be light or dark
// regardless of the operating system setting. System uses the default theme, which follows the setting
type variantTheme struct {
	base    fyne.Theme
	variant fyne.ThemeVariant
}

func newVariantTheme(variant fyne.ThemeVariant) fyne.Theme {
	return variantTheme{base: theme.DefaultTheme(), variant: variant}
}

func (t variantTheme) Color(name fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	return t.base.Color(name, t.variant)
}

func (t variantTheme) Font(style fyne.TextStyle) fyne.Resource { return t.base.Font(style) }

func (t variantTheme) Icon(name fyne.ThemeIconName) fyne.Resource { return t.base.Icon(name) }

func (t variantTheme) Size(name fyne.ThemeSizeName) float32 { return t.base.Size(name) }
