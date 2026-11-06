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

import _ "embed"

// launcherIcon is the icon the desktop shortcut and menu entry use when the package has not installed one.
// It is the same file as packaging/icons/hicolor/256x256/apps/gamedistapp.png.
//
//go:embed resources/gamedistapp-256.png
var launcherIcon []byte
