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
	"os"
	"path/filepath"
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"

	"github.com/chippyash/gamedistributor/src/desktop/model"
)

// fakeTools records the commands the launcher runs. xdg-user-dir reports the Desktop folder, and gio and
// update-desktop-database are present
type fakeTools struct {
	calls [][]string
}

func (f *fakeTools) commands(desktop string) model.Commands {
	return model.Commands{
		Has: func(name string) bool { return name == "gio" || name == "update-desktop-database" },
		Run: func(name string, args ...string) (string, error) {
			if name == "xdg-user-dir" {
				return desktop + "\n", nil
			}
			f.calls = append(f.calls, append([]string{name}, args...))
			return "", nil
		},
	}
}

// newTestLaunch returns a launcher with temporary folders, an enabled desktop, and fake tools
func newTestLaunch(t *testing.T) (launcherEnv, *fakeTools) {
	t.Helper()
	home := t.TempDir()
	desktop := filepath.Join(home, "Desktop")
	if err := os.Mkdir(desktop, 0o755); err != nil {
		t.Fatal(err)
	}
	tools := &fakeTools{}
	return launcherEnv{
		enabled:     true,
		home:        home,
		dataDir:     filepath.Join(home, ".local", "share"),
		exe:         filepath.Join(home, "opt", "gamedistapp"),
		tempDir:     filepath.Join(string(filepath.Separator), "no-such-temp"),
		systemEntry: filepath.Join(home, "usr", "share", "applications", "gamedistapp.desktop"),
		systemIcon:  filepath.Join(home, "usr", "share", "icons", "gamedistapp.png"),
		commands:    tools.commands(desktop),
	}, tools
}

// quietUI builds the window with a launcher, and with the splash already finished, so no timer runs during the test
func quietUI(t *testing.T, env launcherEnv) (*ui, fyne.App) {
	t.Helper()
	a := test.NewApp()
	t.Cleanup(a.Quit)
	u := newUIWithLauncher(a, "v1.3.0", env)
	u.splash.timer.Stop()
	u.splash.finished = true
	u.splashing = false
	return u, a
}

func questionShown(u *ui) bool {
	return u.win.Canvas().Overlays().Top() != nil
}

func desktopPath(env launcherEnv) string {
	return filepath.Join(env.home, "Desktop", model.ShortcutName)
}

func menuPath(env launcherEnv) string {
	return filepath.Join(env.dataDir, model.MenuFolderRel, model.ShortcutName)
}

func readOrEmpty(path string) string {
	b, _ := os.ReadFile(path)
	return string(b)
}

func TestQuestionAppearsOnlyAfterTheSplash(t *testing.T) {
	env, _ := newTestLaunch(t)
	u, _ := quietUI(t, env)
	u.splashing = true
	u.splashDone()
	if !questionShown(u) {
		t.Fatal("the question did not appear after the splash")
	}
}

func TestQuestionIsAskedOnlyInTheRightStates(t *testing.T) {
	cases := []struct {
		name      string
		choice    string
		root      bool
		noDesktop bool
		disabled  bool
		wantAsked bool
	}{
		{"no choice yet", "", false, false, false, true},
		{"created stops the question", model.ChoiceCreated, false, false, false, false},
		{"never stops the question", model.ChoiceNever, false, false, false, false},
		{"root is never asked", "", true, false, false, false},
		{"no desktop folder is not asked", "", false, true, false, false},
		{"disabled on Windows", "", false, false, true, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			env, _ := newTestLaunch(t)
			env.root = tc.root
			env.enabled = !tc.disabled
			if tc.noDesktop {
				if err := os.RemoveAll(filepath.Join(env.home, "Desktop")); err != nil {
					t.Fatal(err)
				}
			}
			u, a := quietUI(t, env)
			if tc.choice != "" {
				a.Preferences().SetString(model.ChoiceKey, tc.choice)
			}
			u.splashing = true
			u.splashDone()
			if got := questionShown(u); got != tc.wantAsked {
				t.Errorf("question shown = %v, want %v", got, tc.wantAsked)
			}
		})
	}
}

func TestRefreshRunsBeforeTheQuestion(t *testing.T) {
	env, _ := newTestLaunch(t)
	old := env.exe
	env.exe = filepath.Join(env.home, "moved", "gamedistapp")
	// an entry the app wrote earlier, from the old location
	if _, err := model.WriteEntry(filepath.Join(env.home, "Desktop"), testLaunchShortcut(old)); err != nil {
		t.Fatal(err)
	}
	u, _ := quietUI(t, env)
	u.splashing = true
	u.splashDone()
	if got := readOrEmpty(desktopPath(env)); !strings.Contains(got, env.exe) {
		t.Errorf("the entry was not refreshed to %s before the question:\n%s", env.exe, got)
	}
}

func testLaunchShortcut(exe string) model.Shortcut {
	return model.Shortcut{Name: "Game Distributor", Comment: "Distribute players across a set of fixtures",
		Categories: "Utility;", Exec: exe, Icon: "/usr/share/icons/hicolor/256x256/apps/gamedistapp.png"}
}

func TestAnswersStoreTheRightChoice(t *testing.T) {
	t.Run("never ask stores never and makes no shortcut", func(t *testing.T) {
		env, _ := newTestLaunch(t)
		u, a := quietUI(t, env)
		u.applyLauncherAnswer("never")
		if got := a.Preferences().String(model.ChoiceKey); got != model.ChoiceNever {
			t.Errorf("choice = %q, want never", got)
		}
		if _, err := os.Stat(desktopPath(env)); err == nil {
			t.Error("a shortcut was made for Never ask")
		}
	})
	t.Run("not now stores nothing", func(t *testing.T) {
		env, _ := newTestLaunch(t)
		u, a := quietUI(t, env)
		u.applyLauncherAnswer("")
		if got := a.Preferences().String(model.ChoiceKey); got != "" {
			t.Errorf("choice = %q, want none", got)
		}
	})
	t.Run("create makes the shortcut and stores created", func(t *testing.T) {
		env, _ := newTestLaunch(t)
		u, a := quietUI(t, env)
		u.applyLauncherAnswer("create")
		if got := a.Preferences().String(model.ChoiceKey); got != model.ChoiceCreated {
			t.Errorf("choice = %q, want created", got)
		}
		info, err := os.Stat(desktopPath(env))
		if err != nil {
			t.Fatalf("shortcut not made: %v", err)
		}
		if info.Mode().Perm() != 0o755 {
			t.Errorf("mode = %o, want 0755", info.Mode().Perm())
		}
	})
}

func TestCreateStoresNothingWhenTheFileIsForeign(t *testing.T) {
	env, _ := newTestLaunch(t)
	foreign := []byte("[Desktop Entry]\nName=Mine\n")
	if err := os.WriteFile(desktopPath(env), foreign, 0o644); err != nil {
		t.Fatal(err)
	}
	u, a := quietUI(t, env)
	r := u.createLauncher(false)
	if r.desktop != model.ActionConflict {
		t.Errorf("desktop action = %v, want conflict", r.desktop)
	}
	if got := a.Preferences().String(model.ChoiceKey); got != "" {
		t.Errorf("choice = %q, want none", got)
	}
	if got := readOrEmpty(desktopPath(env)); got != string(foreign) {
		t.Error("the foreign file was changed")
	}
}

func TestCreateRefusedFromATemporaryFolder(t *testing.T) {
	env, _ := newTestLaunch(t)
	env.tempDir = env.home // the copy lives in a temporary folder
	u, a := quietUI(t, env)
	r := u.createLauncher(false)
	if r.desktopErr != errTemp {
		t.Errorf("desktop error = %v, want the temporary folder error", r.desktopErr)
	}
	if _, err := os.Stat(desktopPath(env)); err == nil {
		t.Error("a shortcut was made from a temporary folder")
	}
	if got := a.Preferences().String(model.ChoiceKey); got != "" {
		t.Errorf("choice = %q, want none", got)
	}
}

func TestCreateWithoutADesktopFolder(t *testing.T) {
	env, _ := newTestLaunch(t)
	if err := os.RemoveAll(filepath.Join(env.home, "Desktop")); err != nil {
		t.Fatal(err)
	}
	u, a := quietUI(t, env)
	r := u.createLauncher(true)
	if r.desktopErr != errNoDesktop {
		t.Errorf("desktop error = %v, want no Desktop folder", r.desktopErr)
	}
	if got := a.Preferences().String(model.ChoiceKey); got != "" {
		t.Errorf("choice = %q, want none", got)
	}
}

func TestMenuEntryAddedOnlyWhenNoSystemEntry(t *testing.T) {
	t.Run("no system entry, the menu entry is added", func(t *testing.T) {
		env, _ := newTestLaunch(t)
		u, _ := quietUI(t, env)
		r := u.createLauncher(true)
		if !r.menuOffered || r.menuErr != nil {
			t.Fatalf("menu offered=%v err=%v", r.menuOffered, r.menuErr)
		}
		if _, err := os.Stat(menuPath(env)); err != nil {
			t.Errorf("menu entry not written: %v", err)
		}
	})
	t.Run("system entry exists, nothing is offered or written", func(t *testing.T) {
		env, _ := newTestLaunch(t)
		if err := os.MkdirAll(filepath.Dir(env.systemEntry), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(env.systemEntry, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		u, _ := quietUI(t, env)
		r := u.createLauncher(true)
		if r.menuOffered {
			t.Error("menu entry offered although a system entry exists")
		}
		if _, err := os.Stat(menuPath(env)); err == nil {
			t.Error("menu entry written although a system entry exists")
		}
	})
}

func TestCreateCallsGioAndTheDatabaseTool(t *testing.T) {
	env, tools := newTestLaunch(t)
	u, _ := quietUI(t, env)
	u.createLauncher(true)
	var gio, db bool
	for _, c := range tools.calls {
		if len(c) >= 2 && c[0] == "gio" && c[1] == "set" && c[2] == desktopPath(env) {
			gio = true
		}
		if len(c) >= 1 && c[0] == "update-desktop-database" {
			db = true
		}
	}
	if !gio {
		t.Errorf("gio did not mark the shortcut as trusted: %v", tools.calls)
	}
	if !db {
		t.Errorf("the desktop database was not updated: %v", tools.calls)
	}
}

func TestUpToDateAfterCreating(t *testing.T) {
	env, _ := newTestLaunch(t)
	u, _ := quietUI(t, env)
	if u.upToDate(filepath.Join(env.home, "Desktop")) {
		t.Error("up to date before anything was created")
	}
	u.createLauncher(true)
	if !u.upToDate(filepath.Join(env.home, "Desktop")) {
		t.Error("not up to date straight after creating both entries")
	}
}

func TestRefreshRewritesOwnEntryAndLeavesForeignOne(t *testing.T) {
	env, _ := newTestLaunch(t)
	env.exe = filepath.Join(env.home, "moved", "gamedistapp")
	if _, err := model.WriteEntry(filepath.Join(env.home, "Desktop"), testLaunchShortcut(filepath.Join(env.home, "old"))); err != nil {
		t.Fatal(err)
	}
	u, _ := quietUI(t, env)
	u.refreshLauncher()
	if got := readOrEmpty(desktopPath(env)); !strings.Contains(got, env.exe) {
		t.Errorf("own entry not rewritten:\n%s", got)
	}
	if u.launcherNote != "" {
		t.Errorf("unexpected refresh problem: %s", u.launcherNote)
	}
}

func TestRefreshRemovesOwnMenuEntryWhenSystemEntryExists(t *testing.T) {
	env, _ := newTestLaunch(t)
	if err := os.MkdirAll(filepath.Dir(menuPath(env)), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(menuPath(env), testLaunchShortcut(env.exe).Render(), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(env.systemEntry), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(env.systemEntry, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	u, _ := quietUI(t, env)
	u.refreshLauncher()
	if _, err := os.Stat(menuPath(env)); err == nil {
		t.Error("own menu entry kept although a system entry exists")
	}
}

func TestRefreshDoesNotPointAtATemporaryCopy(t *testing.T) {
	env, _ := newTestLaunch(t)
	if _, err := model.WriteEntry(filepath.Join(env.home, "Desktop"), testLaunchShortcut("/usr/bin/gamedistapp")); err != nil {
		t.Fatal(err)
	}
	env.tempDir = env.home // the running copy is in a temporary folder
	u, _ := quietUI(t, env)
	u.refreshLauncher()
	if got := readOrEmpty(desktopPath(env)); strings.Contains(got, env.exe) {
		t.Errorf("the shortcut was pointed at the temporary copy:\n%s", got)
	}
	if !strings.Contains(u.launcherNote, "temporary folder") {
		t.Errorf("launcherNote = %q, want the temporary folder note", u.launcherNote)
	}
}

func TestReportSaysWhatHappened(t *testing.T) {
	env, _ := newTestLaunch(t)
	u, _ := quietUI(t, env)
	created := u.launchReportLines(launchReport{desktop: model.ActionCreate})
	if !strings.Contains(strings.Join(created, "\n"), "was created") {
		t.Errorf("created report = %v", created)
	}
	if strings.Contains(strings.Join(created, "\n"), "Desktop icons") {
		t.Error("the hidden-icons sentence shown off GNOME")
	}
	u.launch.gnome = true
	gnome := u.launchReportLines(launchReport{desktop: model.ActionCreate})
	if !strings.Contains(strings.Join(gnome, "\n"), "Desktop icons") {
		t.Errorf("the hidden-icons sentence missing on GNOME: %v", gnome)
	}
	conflict := u.launchReportLines(launchReport{desktop: model.ActionConflict})
	if !strings.Contains(strings.Join(conflict, "\n"), "left alone") {
		t.Errorf("conflict report = %v", conflict)
	}
	trust := u.launchReportLines(launchReport{desktop: model.ActionCreate, trustErr: os.ErrPermission})
	if !strings.Contains(strings.Join(trust, "\n"), "Allow executing file as program") {
		t.Errorf("trust report lacks the manual step: %v", trust)
	}
}

func TestMenuItemIsOnlyOfferedWhenEnabled(t *testing.T) {
	const label = "Create desktop shortcut"
	labels := func(u *ui) []string {
		var out []string
		for _, item := range u.menu.Items[0].Items {
			out = append(out, item.Label)
		}
		return out
	}
	contains := func(list []string, s string) bool {
		for _, v := range list {
			if v == s {
				return true
			}
		}
		return false
	}
	env, _ := newTestLaunch(t)
	u, _ := quietUI(t, env)
	if !contains(labels(u), label) {
		t.Errorf("menu lacks %q on Linux: %v", label, labels(u))
	}
	env.enabled = false
	u, _ = quietUI(t, env)
	if contains(labels(u), label) {
		t.Errorf("menu offers %q where it is disabled: %v", label, labels(u))
	}
}

func TestQuestionButtonsStoreTheRightAnswer(t *testing.T) {
	t.Run("never ask stores never", func(t *testing.T) {
		env, _ := newTestLaunch(t)
		u, a := quietUI(t, env)
		d, buttons := u.launchQuestion()
		d.Show()
		buttons["never"].OnTapped()
		if got := a.Preferences().String(model.ChoiceKey); got != model.ChoiceNever {
			t.Errorf("choice = %q, want never", got)
		}
	})
	t.Run("not now stores nothing", func(t *testing.T) {
		env, _ := newTestLaunch(t)
		u, a := quietUI(t, env)
		d, buttons := u.launchQuestion()
		d.Show()
		buttons["not-now"].OnTapped()
		if got := a.Preferences().String(model.ChoiceKey); got != "" {
			t.Errorf("choice = %q, want none", got)
		}
	})
	t.Run("create makes the shortcut and stores created", func(t *testing.T) {
		env, _ := newTestLaunch(t)
		u, a := quietUI(t, env)
		d, buttons := u.launchQuestion()
		d.Show()
		buttons["create"].OnTapped()
		if got := a.Preferences().String(model.ChoiceKey); got != model.ChoiceCreated {
			t.Errorf("choice = %q, want created", got)
		}
		if _, err := os.Stat(desktopPath(env)); err != nil {
			t.Errorf("shortcut not made: %v", err)
		}
	})
	t.Run("closing without a button is Not now", func(t *testing.T) {
		env, _ := newTestLaunch(t)
		u, a := quietUI(t, env)
		d, _ := u.launchQuestion()
		d.Show()
		d.Hide()
		if got := a.Preferences().String(model.ChoiceKey); got != "" {
			t.Errorf("choice = %q, want none", got)
		}
		if _, err := os.Stat(desktopPath(env)); err == nil {
			t.Error("a shortcut was made when the dialog was closed")
		}
	})
}
