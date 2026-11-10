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
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/chippyash/gamedistributor/src/desktop/model"
)

func testShortcut(exec string) model.Shortcut {
	return model.Shortcut{
		Name:       "Game Distributor",
		Comment:    "Distribute players across a set of fixtures",
		Categories: "Utility;",
		Exec:       exec,
		Icon:       "/usr/share/icons/hicolor/256x256/apps/gamedistapp.png",
	}
}

func fakeRun(out string, err error) func(string, ...string) (string, error) {
	return func(string, ...string) (string, error) { return out, err }
}

func TestRenderWritesTheEntryKeysInOrder(t *testing.T) {
	want := "[Desktop Entry]\n" +
		"Type=Application\n" +
		"Name=Game Distributor\n" +
		"Comment=Distribute players across a set of fixtures\n" +
		"Exec=\"/usr/bin/gamedistapp\"\n" +
		"Icon=/usr/share/icons/hicolor/256x256/apps/gamedistapp.png\n" +
		"Terminal=false\n" +
		"Categories=Utility;\n" +
		"X-Gamedist-Managed=true\n"
	if got := string(testShortcut("/usr/bin/gamedistapp").Render()); got != want {
		t.Errorf("Render() =\n%s\nwant\n%s", got, want)
	}
}

func TestRenderQuotesExecForSpecialCharacters(t *testing.T) {
	cases := []struct {
		name string
		path string
		want string // the Exec line as it is written in the file
	}{
		{"plain", "/usr/bin/gamedistapp", `Exec="/usr/bin/gamedistapp"`},
		{"space", "/opt/my apps/gamedistapp", `Exec="/opt/my apps/gamedistapp"`},
		{"double quote", `/a"b`, `Exec="/a\\"b"`},
		{"dollar", "/a$b", `Exec="/a\\$b"`},
		{"backslash", `/a\b`, `Exec="/a\\\\b"`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := model.ParseEntry(testShortcut(tc.path).Render())["Exec"]
			if "Exec="+got != tc.want {
				t.Errorf("Exec line = %q, want %q", "Exec="+got, tc.want)
			}
		})
	}
}

func TestRenderEscapesNewlinesInValues(t *testing.T) {
	s := testShortcut("/usr/bin/gamedistapp")
	s.Comment = "line one\nline two"
	if got := model.ParseEntry(s.Render())["Comment"]; got != `line one\nline two` {
		t.Errorf("Comment = %q, want the newline escaped", got)
	}
}

func TestParseEntryReadsOnlyTheDesktopEntryGroup(t *testing.T) {
	content := []byte("# a comment\n[Other]\nName=Wrong\n[Desktop Entry]\nName=Right\n\n[Desktop Action x]\nExec=no\n")
	keys := model.ParseEntry(content)
	if keys["Name"] != "Right" {
		t.Errorf("Name = %q, want Right", keys["Name"])
	}
	if _, ok := keys["Exec"]; ok {
		t.Error("Exec from another group was read")
	}
}

func TestIsManaged(t *testing.T) {
	if !model.IsManaged(testShortcut("/usr/bin/gamedistapp").Render()) {
		t.Error("a file the app wrote is not managed")
	}
	if model.IsManaged([]byte("[Desktop Entry]\nName=Mine\n")) {
		t.Error("a file without the marker is managed")
	}
}

func TestPlanDecidesWhatToDoWithAnEntry(t *testing.T) {
	want := testShortcut("/usr/bin/gamedistapp")
	current := want.Render()
	moved := testShortcut("/home/me/gamedistapp").Render()
	renamed := testShortcut("/usr/bin/gamedistapp")
	renamed.Name = "Game Distributor (old)"
	foreign := []byte("[Desktop Entry]\nType=Application\nName=Mine\n")

	cases := []struct {
		name     string
		exists   bool
		existing []byte
		want     model.Action
	}{
		{"no file", false, nil, model.ActionCreate},
		{"same content", true, current, model.ActionKeep},
		{"exec changed, e.g. after a move", true, moved, model.ActionRewrite},
		{"a text key changed", true, renamed.Render(), model.ActionRewrite},
		{"file without the marker", true, foreign, model.ActionConflict},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := model.Plan(tc.exists, tc.existing, want); got != tc.want {
				t.Errorf("Plan() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestDesktopFolder(t *testing.T) {
	home := t.TempDir()
	desktop := filepath.Join(home, "Desktop")
	bureau := filepath.Join(home, "Bureau")
	for _, d := range []string{desktop, bureau} {
		if err := os.Mkdir(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	errMissing := errors.New("xdg-user-dir: not found")

	cases := []struct {
		name   string
		home   string
		out    string
		err    error
		want   string
		wantOK bool
	}{
		{"xdg-user-dir names an existing folder", home, bureau + "\n", nil, bureau, true},
		{"xdg-user-dir returns the home folder", home, home + "\n", nil, "", false},
		{"xdg-user-dir names a missing folder", home, filepath.Join(home, "Missing") + "\n", nil, "", false},
		{"xdg-user-dir cannot be run, ~/Desktop exists", home, "", errMissing, desktop, true},
		{"xdg-user-dir returns nothing, ~/Desktop exists", home, "", nil, desktop, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := model.DesktopFolder(tc.home, fakeRun(tc.out, tc.err))
			if got != tc.want || ok != tc.wantOK {
				t.Errorf("DesktopFolder() = %q, %v; want %q, %v", got, ok, tc.want, tc.wantOK)
			}
		})
	}
}

func TestDesktopFolderWithoutDesktopFolder(t *testing.T) {
	home := t.TempDir()
	if got, ok := model.DesktopFolder(home, fakeRun("", errors.New("missing"))); ok {
		t.Errorf("DesktopFolder() = %q, want no folder when ~/Desktop does not exist", got)
	}
}

func TestShouldAsk(t *testing.T) {
	cases := []struct {
		choice     string
		root       bool
		hasDesktop bool
		want       bool
	}{
		{"", false, true, true},
		{"", false, false, false},
		{"", true, true, false},
		{model.ChoiceCreated, false, true, false},
		{model.ChoiceNever, false, true, false},
		{"unknown", false, true, true},
	}
	for _, tc := range cases {
		if got := model.ShouldAsk(tc.choice, tc.root, tc.hasDesktop); got != tc.want {
			t.Errorf("ShouldAsk(%q, root=%v, desktop=%v) = %v, want %v", tc.choice, tc.root, tc.hasDesktop, got, tc.want)
		}
	}
}

func TestUnderTempDir(t *testing.T) {
	cases := []struct {
		path, tempDir string
		want          bool
	}{
		{"/tmp/go-build/gamedistapp", "/tmp", true},
		{"/tmp", "/tmp", true},
		{"/tmp2/gamedistapp", "/tmp", false},
		{"/usr/bin/gamedistapp", "/tmp", false},
		{"/tmp/../usr/bin/gamedistapp", "/tmp", false},
	}
	for _, tc := range cases {
		if got := model.UnderTempDir(tc.path, tc.tempDir); got != tc.want {
			t.Errorf("UnderTempDir(%q, %q) = %v, want %v", tc.path, tc.tempDir, got, tc.want)
		}
	}
}

func TestIsGNOME(t *testing.T) {
	cases := []struct {
		value string
		want  bool
	}{
		{"GNOME", true},
		{"ubuntu:GNOME", true},
		{"GNOME-Classic:GNOME", true},
		{"gnome", true},
		{"KDE", false},
		{"XFCE", false},
		{"", false},
	}
	for _, tc := range cases {
		if got := model.IsGNOME(tc.value); got != tc.want {
			t.Errorf("IsGNOME(%q) = %v, want %v", tc.value, got, tc.want)
		}
	}
}
