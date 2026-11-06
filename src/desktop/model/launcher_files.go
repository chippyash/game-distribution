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

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

const (
	// SystemEntry is the desktop entry the package installs for the application menu
	SystemEntry = "/usr/share/applications/gamedistapp.desktop"
	// SystemIcon is the icon the package installs
	SystemIcon = "/usr/share/icons/hicolor/256x256/apps/gamedistapp.png"
	// LocalIconRel is the icon path under the user's data folder, used when the package has not installed the icon
	LocalIconRel = "icons/hicolor/256x256/apps/gamedistapp.png"
	// MenuFolderRel is the user's application menu folder, under the data folder
	MenuFolderRel = "applications"
)

// Commands are the external tools the launcher may use. Has reports whether a tool is on the path, and Run executes it
// and returns its output. Both are replaced in tests.
type Commands struct {
	Has func(name string) bool
	Run func(name string, args ...string) (string, error)
}

// WriteEntry writes the shortcut into dir according to Plan, and returns the action it took. A conflicting file is left
// alone, and the action says so.
func WriteEntry(dir string, want Shortcut) (Action, error) {
	path := filepath.Join(dir, ShortcutName)
	existing, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return ActionConflict, err
	}
	action := Plan(err == nil, existing, want)
	if action != ActionCreate && action != ActionRewrite {
		return action, nil
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return action, err
	}
	if err := os.WriteFile(path, want.Render(), 0o755); err != nil {
		return action, err
	}
	// WriteFile keeps the mode of a file that already exists, so set it explicitly
	return action, os.Chmod(path, 0o755)
}

// EnsureIcon returns the icon path the shortcut should use. The package's icon is used if the package installed it. Otherwise
// the embedded icon is written to localIcon, when that file is missing or its bytes differ.
func EnsureIcon(systemIcon, localIcon string, embedded []byte) (string, error) {
	if _, err := os.Stat(systemIcon); err == nil {
		return systemIcon, nil
	}
	current, err := os.ReadFile(localIcon)
	if err == nil && bytes.Equal(current, embedded) {
		return localIcon, nil
	}
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(localIcon), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(localIcon, embedded, 0o644); err != nil {
		return "", err
	}
	return localIcon, os.Chmod(localIcon, 0o644)
}

// RemoveOwnMenuEntry deletes the app's own user menu entry in userDir, but only when the system entry at systemEntry exists,
// and only when the file carries the managed marker. It reports whether it removed a file.
func RemoveOwnMenuEntry(userDir, systemEntry string) (bool, error) {
	if _, err := os.Stat(systemEntry); err != nil {
		return false, nil
	}
	path := filepath.Join(userDir, ShortcutName)
	content, err := os.ReadFile(path)
	if err != nil {
		return false, nil
	}
	if !IsManaged(content) {
		return false, nil
	}
	return true, os.Remove(path)
}

// MarkTrusted marks a desktop file as trusted with gio, which GNOME uses to allow a double-click to launch it. It does
// nothing when gio is not installed.
func MarkTrusted(path string, c Commands) error {
	if !c.Has("gio") {
		return nil
	}
	if _, err := c.Run("gio", "set", path, "metadata::trusted", "true"); err != nil {
		return fmt.Errorf("marking %s as trusted: %w", path, err)
	}
	return nil
}

// RefreshDatabase updates the desktop database in dir, so that a new or changed entry shows in the menu. It does nothing
// when update-desktop-database is not installed.
func RefreshDatabase(dir string, c Commands) error {
	if !c.Has("update-desktop-database") {
		return nil
	}
	if _, err := c.Run("update-desktop-database", dir); err != nil {
		return fmt.Errorf("updating the desktop database: %w", err)
	}
	return nil
}
