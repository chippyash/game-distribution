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
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	// ShortcutName is the file name of the Desktop shortcut and the user menu entry
	ShortcutName = "gamedistapp.desktop"
	// ManagedKey marks a desktop entry as one the app wrote, so the app may change it
	ManagedKey = "X-Gamedist-Managed"
	// ChoiceKey is the Fyne preference that stores the answer to the desktop shortcut question
	ChoiceKey = "desktopShortcut"
	// ChoiceCreated means the user chose "Create"
	ChoiceCreated = "created"
	// ChoiceNever means the user chose "Never ask"
	ChoiceNever = "never"
)

// Shortcut holds the values the app writes into a desktop entry
type Shortcut struct {
	Name       string
	Comment    string
	Categories string
	Exec       string // the full path of the running binary, unquoted
	Icon       string // the absolute path of the icon file
}

// Action is what the app should do with a desktop entry it wants to write
type Action int

const (
	// ActionCreate means there is no file, so the app writes one
	ActionCreate Action = iota
	// ActionRewrite means the app wrote the file, and a key it writes differs, so the app rewrites it
	ActionRewrite
	// ActionKeep means the app wrote the file, and every key it writes is current
	ActionKeep
	// ActionConflict means a file of that name exists and the app did not write it, so the app leaves it alone
	ActionConflict
)

// DesktopFolder finds the user's Desktop folder. run executes a command and returns its output, and is used for
// xdg-user-dir. A result equal to home, or a folder that does not exist, is treated as no Desktop folder. ~/Desktop is
// used only when xdg-user-dir cannot be run.
func DesktopFolder(home string, run func(name string, args ...string) (string, error)) (string, bool) {
	out, err := run("xdg-user-dir", "DESKTOP")
	dir := strings.TrimSpace(out)
	if err != nil || dir == "" {
		return usableFolder(filepath.Join(home, "Desktop"), home)
	}
	return usableFolder(dir, home)
}

// usableFolder reports whether dir is an existing folder that is not the home folder
func usableFolder(dir, home string) (string, bool) {
	if filepath.Clean(dir) == filepath.Clean(home) {
		return "", false
	}
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		return "", false
	}
	return dir, true
}

// Render returns the desktop entry file content for the shortcut, with the managed marker
func (s Shortcut) Render() []byte {
	var b bytes.Buffer
	b.WriteString("[Desktop Entry]\n")
	b.WriteString("Type=Application\n")
	fmt.Fprintf(&b, "Name=%s\n", escapeValue(s.Name))
	fmt.Fprintf(&b, "Comment=%s\n", escapeValue(s.Comment))
	fmt.Fprintf(&b, "Exec=%s\n", escapeValue(quoteExec(s.Exec)))
	fmt.Fprintf(&b, "Icon=%s\n", escapeValue(s.Icon))
	b.WriteString("Terminal=false\n")
	fmt.Fprintf(&b, "Categories=%s\n", escapeValue(s.Categories))
	fmt.Fprintf(&b, "%s=true\n", ManagedKey)
	return b.Bytes()
}

// quoteExec quotes a path for the Exec key, as the desktop entry specification requires. Inside the quotes, a double
// quote, a backtick, a dollar sign or a backslash is escaped with a backslash.
func quoteExec(path string) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range path {
		if strings.ContainsRune("\"`$\\", r) {
			b.WriteByte('\\')
		}
		b.WriteRune(r)
	}
	b.WriteByte('"')
	return b.String()
}

// escapeValue escapes a value for a desktop entry string, which uses backslash escapes
func escapeValue(v string) string {
	r := strings.NewReplacer("\\", "\\\\", "\n", "\\n", "\t", "\\t")
	return r.Replace(v)
}

// ParseEntry returns the keys of the [Desktop Entry] group, with their values as they are written in the file
func ParseEntry(content []byte) map[string]string {
	keys := map[string]string{}
	inGroup := false
	scanner := bufio.NewScanner(bytes.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		switch {
		case line == "[Desktop Entry]":
			inGroup = true
		case strings.HasPrefix(line, "["):
			inGroup = false
		case inGroup && line != "" && !strings.HasPrefix(line, "#"):
			if k, v, ok := strings.Cut(line, "="); ok {
				keys[strings.TrimSpace(k)] = strings.TrimSpace(v)
			}
		}
	}
	return keys
}

// IsManaged reports whether the desktop entry was written by the app
func IsManaged(content []byte) bool {
	return ParseEntry(content)[ManagedKey] == "true"
}

// Plan decides what to do with a desktop entry. exists says whether a file is there, and existing is its content.
// The app rewrites a file only when it wrote it and a key the app writes differs.
func Plan(exists bool, existing []byte, want Shortcut) Action {
	if !exists {
		return ActionCreate
	}
	if !IsManaged(existing) {
		return ActionConflict
	}
	have := ParseEntry(existing)
	for k, v := range ParseEntry(want.Render()) {
		if have[k] != v {
			return ActionRewrite
		}
	}
	return ActionKeep
}

// ShouldAsk reports whether the first-run question is due. choice is the stored value of ChoiceKey, which is empty until
// the user answers. The question is never asked as root, and never when there is no Desktop folder.
func ShouldAsk(choice string, root, hasDesktop bool) bool {
	if choice == ChoiceCreated || choice == ChoiceNever {
		return false
	}
	return !root && hasDesktop
}

// UnderTempDir reports whether path is tempDir or inside it. A folder that only starts with the same name does not count.
func UnderTempDir(path, tempDir string) bool {
	p := filepath.Clean(path)
	t := filepath.Clean(tempDir)
	return p == t || strings.HasPrefix(p, t+string(filepath.Separator))
}

// IsGNOME reports whether the XDG_CURRENT_DESKTOP value names a GNOME desktop. The value is a list of names separated by
// colons, such as "ubuntu:GNOME", and a name such as "GNOME-Classic" also counts as GNOME.
func IsGNOME(value string) bool {
	for _, name := range strings.Split(value, ":") {
		name = strings.ToUpper(strings.TrimSpace(name))
		if name == "GNOME" || strings.HasPrefix(name, "GNOME-") {
			return true
		}
	}
	return false
}
