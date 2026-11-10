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
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/chippyash/gamedistributor/src/desktop/model"
)

func modeOf(t *testing.T, path string) os.FileMode {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	return info.Mode().Perm()
}

func TestWriteEntryCreatesRewritesAndKeeps(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, model.ShortcutName)

	action, err := model.WriteEntry(dir, testShortcut("/usr/bin/gamedistapp"))
	if err != nil || action != model.ActionCreate {
		t.Fatalf("first write = %v, %v; want create", action, err)
	}
	if got, _ := os.ReadFile(path); !bytes.Equal(got, testShortcut("/usr/bin/gamedistapp").Render()) {
		t.Error("created file does not match the rendered entry")
	}
	if m := modeOf(t, path); m != 0o755 {
		t.Errorf("mode = %o, want 0755", m)
	}

	action, err = model.WriteEntry(dir, testShortcut("/usr/bin/gamedistapp"))
	if err != nil || action != model.ActionKeep {
		t.Errorf("same write = %v, %v; want keep", action, err)
	}

	action, err = model.WriteEntry(dir, testShortcut("/home/me/gamedistapp"))
	if err != nil || action != model.ActionRewrite {
		t.Fatalf("moved write = %v, %v; want rewrite", action, err)
	}
	if got, _ := os.ReadFile(path); !bytes.Contains(got, []byte("Exec=\"/home/me/gamedistapp\"")) {
		t.Error("rewritten file does not carry the new Exec")
	}
	if m := modeOf(t, path); m != 0o755 {
		t.Errorf("mode after rewrite = %o, want 0755", m)
	}
}

func TestWriteEntryLeavesAForeignFileAlone(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, model.ShortcutName)
	foreign := []byte("[Desktop Entry]\nName=Mine\n")
	if err := os.WriteFile(path, foreign, 0o644); err != nil {
		t.Fatal(err)
	}
	action, err := model.WriteEntry(dir, testShortcut("/usr/bin/gamedistapp"))
	if err != nil || action != model.ActionConflict {
		t.Errorf("write over a foreign file = %v, %v; want conflict", action, err)
	}
	if got, _ := os.ReadFile(path); !bytes.Equal(got, foreign) {
		t.Error("a foreign file was changed")
	}
}

func TestWriteEntryCreatesTheFolderWhenMissing(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "applications")
	if _, err := model.WriteEntry(dir, testShortcut("/usr/bin/gamedistapp")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, model.ShortcutName)); err != nil {
		t.Errorf("entry not written: %v", err)
	}
}

func TestEnsureIcon(t *testing.T) {
	embedded := []byte("embedded icon")
	t.Run("the package installed the icon", func(t *testing.T) {
		system := filepath.Join(t.TempDir(), "system.png")
		if err := os.WriteFile(system, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		local := filepath.Join(t.TempDir(), "local.png")
		got, err := model.EnsureIcon(system, local, embedded)
		if err != nil || got != system {
			t.Errorf("EnsureIcon() = %q, %v; want the system icon", got, err)
		}
		if _, err := os.Stat(local); err == nil {
			t.Error("a local icon was written although the package installed one")
		}
	})
	t.Run("no icon yet, so it is written", func(t *testing.T) {
		local := filepath.Join(t.TempDir(), "icons", "gamedistapp.png")
		got, err := model.EnsureIcon(filepath.Join(t.TempDir(), "missing.png"), local, embedded)
		if err != nil || got != local {
			t.Fatalf("EnsureIcon() = %q, %v; want the local icon", got, err)
		}
		if b, _ := os.ReadFile(local); !bytes.Equal(b, embedded) {
			t.Error("the embedded icon was not written")
		}
		if m := modeOf(t, local); m != 0o644 {
			t.Errorf("mode = %o, want 0644", m)
		}
	})
	t.Run("the local icon differs, so it is rewritten", func(t *testing.T) {
		local := filepath.Join(t.TempDir(), "gamedistapp.png")
		if err := os.WriteFile(local, []byte("old icon"), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := model.EnsureIcon(filepath.Join(t.TempDir(), "missing.png"), local, embedded); err != nil {
			t.Fatal(err)
		}
		if b, _ := os.ReadFile(local); !bytes.Equal(b, embedded) {
			t.Error("the icon was not rewritten")
		}
	})
}

func TestRemoveOwnMenuEntry(t *testing.T) {
	system := filepath.Join(t.TempDir(), "gamedistapp.desktop")
	missingSystem := filepath.Join(t.TempDir(), "absent.desktop")

	t.Run("system entry exists, own entry removed", func(t *testing.T) {
		userDir := t.TempDir()
		if err := os.WriteFile(system, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(userDir, model.ShortcutName), testShortcut("/x").Render(), 0o755); err != nil {
			t.Fatal(err)
		}
		removed, err := model.RemoveOwnMenuEntry(userDir, system)
		if err != nil || !removed {
			t.Fatalf("RemoveOwnMenuEntry() = %v, %v; want removed", removed, err)
		}
		if _, err := os.Stat(filepath.Join(userDir, model.ShortcutName)); err == nil {
			t.Error("own entry still exists")
		}
	})
	t.Run("no system entry, own entry kept", func(t *testing.T) {
		userDir := t.TempDir()
		if err := os.WriteFile(filepath.Join(userDir, model.ShortcutName), testShortcut("/x").Render(), 0o755); err != nil {
			t.Fatal(err)
		}
		removed, err := model.RemoveOwnMenuEntry(userDir, missingSystem)
		if err != nil || removed {
			t.Errorf("RemoveOwnMenuEntry() = %v, %v; want kept", removed, err)
		}
	})
	t.Run("system entry exists, foreign entry kept", func(t *testing.T) {
		userDir := t.TempDir()
		foreign := filepath.Join(userDir, model.ShortcutName)
		if err := os.WriteFile(foreign, []byte("[Desktop Entry]\nName=Mine\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		removed, err := model.RemoveOwnMenuEntry(userDir, system)
		if err != nil || removed {
			t.Errorf("RemoveOwnMenuEntry() = %v, %v; want kept", removed, err)
		}
		if _, err := os.Stat(foreign); err != nil {
			t.Error("a foreign entry was deleted")
		}
	})
}

// recorder is a Commands whose tools are present or absent as set, and which records each run
type recorder struct {
	present map[string]bool
	err     error
	calls   [][]string
}

func (r *recorder) commands() model.Commands {
	return model.Commands{
		Has: func(name string) bool { return r.present[name] },
		Run: func(name string, args ...string) (string, error) {
			r.calls = append(r.calls, append([]string{name}, args...))
			return "", r.err
		},
	}
}

func TestMarkTrusted(t *testing.T) {
	t.Run("gio absent does nothing", func(t *testing.T) {
		r := &recorder{}
		if err := model.MarkTrusted("/d/x.desktop", r.commands()); err != nil || len(r.calls) != 0 {
			t.Errorf("MarkTrusted() = %v with calls %v; want nil and no calls", err, r.calls)
		}
	})
	t.Run("gio present marks the file", func(t *testing.T) {
		r := &recorder{present: map[string]bool{"gio": true}}
		if err := model.MarkTrusted("/d/x.desktop", r.commands()); err != nil {
			t.Fatal(err)
		}
		want := []string{"gio", "set", "/d/x.desktop", "metadata::trusted", "true"}
		if len(r.calls) != 1 || !equalStrings(r.calls[0], want) {
			t.Errorf("calls = %v, want %v", r.calls, want)
		}
	})
	t.Run("a failing gio is reported", func(t *testing.T) {
		r := &recorder{present: map[string]bool{"gio": true}, err: errors.New("not supported")}
		if err := model.MarkTrusted("/d/x.desktop", r.commands()); err == nil {
			t.Error("MarkTrusted() = nil, want the error")
		}
	})
}

func TestRefreshDatabase(t *testing.T) {
	t.Run("absent does nothing", func(t *testing.T) {
		r := &recorder{}
		if err := model.RefreshDatabase("/d", r.commands()); err != nil || len(r.calls) != 0 {
			t.Errorf("RefreshDatabase() = %v with calls %v; want nil and no calls", err, r.calls)
		}
	})
	t.Run("present runs on the folder", func(t *testing.T) {
		r := &recorder{present: map[string]bool{"update-desktop-database": true}}
		if err := model.RefreshDatabase("/d", r.commands()); err != nil {
			t.Fatal(err)
		}
		want := []string{"update-desktop-database", "/d"}
		if len(r.calls) != 1 || !equalStrings(r.calls[0], want) {
			t.Errorf("calls = %v, want %v", r.calls, want)
		}
	})
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
