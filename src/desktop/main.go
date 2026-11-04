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
	_ "embed"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/go-gl/glfw/v3.4/glfw"
)

// version is set at build time with -ldflags "-X main.version=<VERSION>"
var version = "dev"

// windowIcon is the window icon, embedded so the binary does not depend on the file's location
//
//go:embed resources/icon-linux.jpeg
var windowIcon []byte

func main() {
	// On a Wayland session the pointer is offset from the drawn widgets, so use X11 (XWayland) when it is available
	if glfw.PlatformSupported(glfw.PlatformX11) {
		glfw.InitHint(glfw.PlatformHint, int(glfw.PlatformX11))
	}
	a := app.New()
	a.SetIcon(fyne.NewStaticResource("icon-linux.jpeg", windowIcon))
	w := newWindow(a, version)
	w.ShowAndRun()
}
