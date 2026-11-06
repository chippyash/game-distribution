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
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/chippyash/gamedistributor/src/desktop/model"
)

var (
	errNoDesktop = errors.New("there is no Desktop folder")
	errTemp      = errors.New("the app runs from a temporary folder, so an installed copy is needed")
	errTempFresh = errors.New("the app runs from a temporary folder, so the existing shortcut was not changed")
)

// launcherEnv is what the desktop shortcut needs from the system. Tests replace it with temporary folders.
type launcherEnv struct {
	enabled     bool // false on Windows, where the launcher is not offered
	root        bool
	gnome       bool
	home        string
	dataDir     string // ~/.local/share
	exe         string
	tempDir     string
	systemEntry string
	systemIcon  string
	commands    model.Commands
}

// defaultLauncherEnv reads the real system
func defaultLauncherEnv() launcherEnv {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Printf("desktop shortcut: no home folder: %v", err)
	}
	exe, err := os.Executable()
	if err != nil {
		log.Printf("desktop shortcut: no executable path: %v", err)
	}
	return launcherEnv{
		enabled:     runtime.GOOS != "windows",
		root:        os.Geteuid() == 0,
		gnome:       model.IsGNOME(os.Getenv("XDG_CURRENT_DESKTOP")),
		home:        home,
		dataDir:     filepath.Join(home, ".local", "share"),
		exe:         exe,
		tempDir:     os.TempDir(),
		systemEntry: model.SystemEntry,
		systemIcon:  model.SystemIcon,
		commands:    model.Commands{Has: hasCommand, Run: runCommand},
	}
}

// hasCommand reports whether a tool is on the path
func hasCommand(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// runCommand runs a tool and returns its standard output. A failure includes the tool's error output
func runCommand(name string, args ...string) (string, error) {
	out, err := exec.Command(name, args...).Output()
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return string(out), fmt.Errorf("%w: %s", err, strings.TrimSpace(string(exitErr.Stderr)))
	}
	return string(out), err
}

// desktopFolder returns the user's Desktop folder, if there is one
func (e launcherEnv) desktopFolder() (string, bool) {
	return model.DesktopFolder(e.home, e.commands.Run)
}

// menuFolder is the user's application menu folder
func (e launcherEnv) menuFolder() string {
	return filepath.Join(e.dataDir, model.MenuFolderRel)
}

// iconPath is the icon the entries use: the package's icon if it is installed, otherwise the one the app writes
func (e launcherEnv) iconPath() string {
	if _, err := os.Stat(e.systemIcon); err == nil {
		return e.systemIcon
	}
	return filepath.Join(e.dataDir, model.LocalIconRel)
}

// shortcut is what the app writes into the desktop entries
func (e launcherEnv) shortcut() model.Shortcut {
	return model.Shortcut{
		Name:       "Game Distributor",
		Comment:    "Distribute players across a set of fixtures",
		Categories: "Utility;",
		Exec:       e.exe,
		Icon:       e.iconPath(),
	}
}

// planEntry says what writing want into dir would do, without writing anything
func planEntry(dir string, want model.Shortcut) (model.Action, error) {
	content, err := os.ReadFile(filepath.Join(dir, model.ShortcutName))
	if errors.Is(err, os.ErrNotExist) {
		return model.Plan(false, nil, want), nil
	}
	if err != nil {
		return model.ActionConflict, err
	}
	return model.Plan(true, content, want), nil
}

// ownedEntry reports whether dir holds an entry the app wrote
func ownedEntry(dir string) bool {
	content, err := os.ReadFile(filepath.Join(dir, model.ShortcutName))
	return err == nil && model.IsManaged(content)
}

// launchReport records what happened for each entry
type launchReport struct {
	desktop, menu       model.Action
	desktopErr, menuErr error
	trustErr            error
	menuOffered         bool
}

// startLauncher runs after the splash. It refreshes the app's existing entries, then asks the first-run question if it is due
func (u *ui) startLauncher() {
	e := u.launch
	if !e.enabled || e.root {
		return
	}
	u.refreshLauncher()
	_, hasDesktop := e.desktopFolder()
	if model.ShouldAsk(u.app.Preferences().String(model.ChoiceKey), e.root, hasDesktop) {
		u.askLauncher()
	}
}

// refreshLauncher rewrites the app's own entries when they are out of date, and removes the user menu entry when a system
// entry exists. Failures go to the log and to the menu item, and never to a dialog
func (u *ui) refreshLauncher() {
	e := u.launch
	var errs []error
	// A copy started from a temporary folder must not point the shortcut at itself, so it does not rewrite the entries
	temp := model.UnderTempDir(e.exe, e.tempDir)
	if temp {
		errs = append(errs, errTempFresh)
	}
	if folder, ok := e.desktopFolder(); ok && !temp && ownedEntry(folder) {
		if _, err := model.EnsureIcon(e.systemIcon, filepath.Join(e.dataDir, model.LocalIconRel), launcherIcon); err != nil {
			errs = append(errs, err)
		}
		if action, err := model.WriteEntry(folder, e.shortcut()); err != nil {
			errs = append(errs, err)
		} else if action == model.ActionRewrite {
			errs = append(errs, model.MarkTrusted(filepath.Join(folder, model.ShortcutName), e.commands))
		}
	}
	menuDir := e.menuFolder()
	removed, err := model.RemoveOwnMenuEntry(menuDir, e.systemEntry)
	if err != nil {
		errs = append(errs, err)
	}
	if removed {
		errs = append(errs, model.RefreshDatabase(menuDir, e.commands))
	} else if !temp && ownedEntry(menuDir) {
		if action, err := model.WriteEntry(menuDir, e.shortcut()); err != nil {
			errs = append(errs, err)
		} else if action == model.ActionRewrite {
			errs = append(errs, model.RefreshDatabase(menuDir, e.commands))
		}
	}
	u.launcherNote = joinErrors(errs)
	if u.launcherNote != "" {
		log.Printf("desktop shortcut: %s", u.launcherNote)
	}
}

// joinErrors makes one line from the errors that are not nil
func joinErrors(errs []error) string {
	var parts []string
	for _, err := range errs {
		if err != nil {
			parts = append(parts, err.Error())
		}
	}
	return strings.Join(parts, "; ")
}

// askLauncher shows the first-run question. Dismissing it counts as "Not now"
func (u *ui) askLauncher() {
	d, _ := u.launchQuestion()
	d.Show()
}

// launchQuestion builds the first-run question, with its buttons keyed "create", "not-now" and "never". Each button sets
// the answer before it hides the dialog, and the close handler acts on the answer. Closing the dialog without a button
// leaves the answer empty, which is Not now
func (u *ui) launchQuestion() (*dialog.CustomDialog, map[string]*widget.Button) {
	answer := ""
	content := widget.NewLabel("Create a desktop shortcut for Game Distributor?")
	content.Wrapping = fyne.TextWrapWord
	d := dialog.NewCustomWithoutButtons("Desktop shortcut", content, u.win)
	buttons := map[string]*widget.Button{
		"create":  {Text: "Create", Importance: widget.HighImportance},
		"not-now": {Text: "Not now"},
		"never":   {Text: "Never ask"},
	}
	buttons["create"].OnTapped = func() {
		answer = "create"
		d.Hide()
	}
	buttons["not-now"].OnTapped = d.Hide
	buttons["never"].OnTapped = func() {
		answer = "never"
		d.Hide()
	}
	d.SetButtons([]fyne.CanvasObject{buttons["create"], buttons["not-now"], buttons["never"]})
	d.SetOnClosed(func() { u.applyLauncherAnswer(answer) })
	return d, buttons
}

// applyLauncherAnswer acts on the first-run answer: "create", "never", or "" for Not now, which stores nothing
func (u *ui) applyLauncherAnswer(answer string) {
	switch answer {
	case "create":
		u.showLaunchReport(u.createLauncher(false))
	case "never":
		u.app.Preferences().SetString(model.ChoiceKey, model.ChoiceNever)
	}
}

// createLauncher writes the Desktop shortcut and, when includeMenu is set and no system entry exists, the user menu entry.
// It stores "created" when the Desktop shortcut exists and is the app's
func (u *ui) createLauncher(includeMenu bool) launchReport {
	e := u.launch
	var r launchReport
	folder, ok := e.desktopFolder()
	if !ok {
		r.desktopErr = errNoDesktop
	} else if model.UnderTempDir(e.exe, e.tempDir) {
		r.desktopErr = errTemp
	} else {
		r = u.writeEntries(folder, includeMenu)
	}
	if r.desktopErr == nil && r.desktop != model.ActionConflict {
		u.app.Preferences().SetString(model.ChoiceKey, model.ChoiceCreated)
	}
	return r
}

// writeEntries does the writing for createLauncher
func (u *ui) writeEntries(folder string, includeMenu bool) launchReport {
	e := u.launch
	var r launchReport
	if _, err := model.EnsureIcon(e.systemIcon, filepath.Join(e.dataDir, model.LocalIconRel), launcherIcon); err != nil {
		r.desktopErr = err
		return r
	}
	want := e.shortcut()
	r.desktop, r.desktopErr = model.WriteEntry(folder, want)
	if r.desktopErr == nil && (r.desktop == model.ActionCreate || r.desktop == model.ActionRewrite) {
		r.trustErr = model.MarkTrusted(filepath.Join(folder, model.ShortcutName), e.commands)
	}
	if includeMenu {
		if _, err := os.Stat(e.systemEntry); err != nil {
			r.menuOffered = true
			r.menu, r.menuErr = model.WriteEntry(e.menuFolder(), want)
			if r.menuErr == nil && (r.menu == model.ActionCreate || r.menu == model.ActionRewrite) {
				r.menuErr = model.RefreshDatabase(e.menuFolder(), e.commands)
			}
		}
	}
	return r
}

// showLaunchReport tells the user what happened to each entry, in one dialog
func (u *ui) showLaunchReport(r launchReport) {
	dialog.ShowInformation("Desktop shortcut", strings.Join(u.launchReportLines(r), "\n"), u.win)
}

// launchReportLines is the text of the report for each entry
func (u *ui) launchReportLines(r launchReport) []string {
	var lines []string
	switch {
	case errors.Is(r.desktopErr, errNoDesktop), errors.Is(r.desktopErr, errTemp):
		lines = append(lines, "The desktop shortcut was not created: "+r.desktopErr.Error()+".")
	case r.desktopErr != nil:
		lines = append(lines, "The desktop shortcut could not be created: "+r.desktopErr.Error()+".")
	case r.desktop == model.ActionConflict:
		lines = append(lines, "A file named "+model.ShortcutName+" is already on the Desktop, and it was not made by this app. It was left alone.")
	case r.desktop == model.ActionKeep:
		lines = append(lines, "The desktop shortcut is already up to date.")
	default:
		lines = append(lines, "The desktop shortcut was created.")
		if u.launch.gnome {
			lines = append(lines, "The shortcut may not be shown if Desktop icons are hidden, for example without the Desktop icons extension.")
		}
	}
	if r.trustErr != nil {
		lines = append(lines, "It could not be marked as trusted ("+r.trustErr.Error()+"). If it does not open on a double-click in GNOME, "+
			"right-click it, choose Properties, open the Permissions tab, and tick \"Allow executing file as program\".")
	}
	if r.menuOffered {
		switch {
		case r.menuErr != nil:
			lines = append(lines, "The application menu entry could not be added: "+r.menuErr.Error()+".")
		case r.menu == model.ActionConflict:
			lines = append(lines, "A file named "+model.ShortcutName+" is already in the application menu folder, and it was not made by this app. It was left alone.")
		case r.menu == model.ActionKeep:
			lines = append(lines, "The application menu entry is already up to date.")
		default:
			lines = append(lines, "The application menu entry was added.")
		}
	}
	return lines
}

// onCreateShortcut is the "Create desktop shortcut" menu item. It works at any time, whatever the stored choice
func (u *ui) onCreateShortcut() {
	if u.splashing || !u.launch.enabled {
		return
	}
	e := u.launch
	if e.root {
		dialog.ShowInformation("Desktop shortcut", "Run Game Distributor as your own user to create the shortcut.", u.win)
		return
	}
	folder, ok := e.desktopFolder()
	if !ok {
		dialog.ShowInformation("Desktop shortcut", "There is no Desktop folder, so no shortcut can be created here.", u.win)
		return
	}
	if u.upToDate(folder) {
		dialog.ShowInformation("Desktop shortcut", "The desktop shortcut and the menu entry are already up to date.", u.win)
		return
	}
	var content fyne.CanvasObject = widget.NewLabel("Create a desktop shortcut for Game Distributor?")
	check := widget.NewCheck("Also add to the application menu", nil)
	if _, err := os.Stat(e.systemEntry); err != nil {
		content = container.NewVBox(content, check)
	} else {
		check.Hide()
	}
	if u.launcherNote != "" {
		problem := widget.NewLabel("Last refresh problem: " + u.launcherNote)
		problem.Wrapping = fyne.TextWrapWord
		content = container.NewVBox(content, problem)
	}
	dialog.NewCustomConfirm("Create desktop shortcut", "Create", "Cancel", content, func(ok bool) {
		if ok {
			u.showLaunchReport(u.createLauncher(check.Checked))
		}
	}, u.win).Show()
}

// upToDate reports whether the Desktop shortcut is current, and the menu entry is current or provided by the system
func (u *ui) upToDate(folder string) bool {
	e := u.launch
	desktop, err := planEntry(folder, e.shortcut())
	if err != nil || desktop != model.ActionKeep {
		return false
	}
	if _, err := os.Stat(e.systemEntry); err == nil {
		return true
	}
	menu, err := planEntry(e.menuFolder(), e.shortcut())
	return err == nil && menu == model.ActionKeep
}
