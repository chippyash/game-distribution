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

import "fyne.io/fyne/v2"

// newTestUI builds the main window with the desktop shortcut turned off, so no test reads or writes the real Desktop
// or application menu
func newTestUI(a fyne.App, version string) *ui {
	return newUIWithLauncher(a, version, launcherEnv{})
}
