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
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/chippyash/gamedistributor/src/desktop/model"
	"github.com/chippyash/gamedistributor/src/distributor"
)

// ui holds the widgets and the state of the window. Widgets are only touched from the Fyne UI thread.
type ui struct {
	win     fyne.Window
	app     fyne.App
	menu    *fyne.MainMenu
	version string

	fixtures, players, perFixture, maxTogether, attempts *helpEntry
	fieldErrs                                            map[string]*widget.Label
	form                                                 *fyne.Container
	tip                                                  *helpTip
	inputs                                               []fyne.Disableable
	spacing, transpose                                   *helpCheck
	start                                                *enterButton
	cancel, save                                         *widget.Button
	splash                                               *splashScreen
	splashing                                            bool
	spinner                                              *widget.ProgressBarInfinite
	status, panel                                        *widget.Label
	table                                                *widget.Table

	result               *distributor.Distributor
	rowLabels, colLabels []string
	cells                [][]string
	cancelRun            context.CancelFunc
	running              bool
	cancelling           bool
}

// newWindow builds the main window
func newWindow(a fyne.App, version string) fyne.Window {
	return newUI(a, version).win
}

// newUI builds the main window and returns the state behind it
// inputSpacing is the gap in pixels between the input fields
const inputSpacing = 2

func newUI(a fyne.App, version string) *ui {
	u := &ui{win: a.NewWindow("Game Distributor"), app: a, version: version, fieldErrs: map[string]*widget.Label{}}
	u.win.Resize(fyne.NewSize(1000, 600)) // Fyne grows this to the content's minimum height
	u.tip = newHelpTip(u.inputsActive)
	applyTheme(a, model.ThemeFromPreference(a.Preferences().String(model.PrefTheme)))

	u.fixtures = u.numberEntry("", model.FieldFixtures, "Fixtures")
	u.players = u.numberEntry("", model.FieldPlayers, "Players")
	u.perFixture = u.numberEntry("", model.FieldPlayersPerFixture, "Players per fixture")
	defaults := model.DefaultSettings()
	u.maxTogether = u.numberEntry(strconv.Itoa(defaults.MaxPlayersTogether), model.FieldMaxPlayersTogether, "Max players together")
	u.attempts = u.numberEntry(strconv.Itoa(defaults.Attempts), model.FieldAttempts, "Attempts")

	u.spacing = newHelpCheck("Fixture spacing (-s)", model.InputHelp(model.FieldSpacing), func(bool) { u.inputChanged() }, u.tip)
	u.transpose = newHelpCheck("Transpose (-t)", model.InputHelp(model.FieldTranspose), func(bool) { u.transposeChanged() }, u.tip)
	u.inputs = []fyne.Disableable{u.fixtures, u.players, u.perFixture, u.maxTogether, u.attempts, u.spacing}

	u.start = newEnterButton("Start", u.onStart)
	u.cancel = widget.NewButton("Cancel", u.onCancel)
	u.cancel.Disable()
	u.save = widget.NewButton("Save CSV", u.onSave)
	u.save.Disable()

	u.spinner = widget.NewProgressBarInfinite()
	u.spinner.Hide()
	u.status = widget.NewLabel("")
	u.panel = widget.NewLabel("")
	u.panel.Wrapping = fyne.TextWrapWord

	u.table = widget.NewTableWithHeaders(u.tableSize, u.tableCell, u.tableUpdate)

	inputForm := container.New(layout.NewCustomPaddedVBoxLayout(inputSpacing),
		u.field("Num. of Fixture", u.fixtures, model.FieldFixtures),
		u.field("Num. of Players", u.players, model.FieldPlayers),
		u.field("Players per fixture", u.perFixture, model.FieldPlayersPerFixture),
		u.field("Max players together", u.maxTogether, model.FieldMaxPlayersTogether),
		u.field("Max. Attempts", u.attempts, model.FieldAttempts),
		u.spacing,
		u.transpose,
	)
	u.form = inputForm
	controls := container.NewVBox(
		inputForm,
		container.NewGridWithColumns(3, u.start, u.cancel, u.save),
		u.spinner,
		u.status,
	)
	panel := container.New(layout.NewGridWrapLayout(fyne.NewSize(600, 100)), container.NewVScroll(u.panel))
	u.menu = fyne.NewMainMenu(fyne.NewMenu("Menu",
		fyne.NewMenuItem("Theme", u.onTheme),
		fyne.NewMenuItem("About", u.onAbout),
		fyne.NewMenuItem("Help", u.onHelp),
	))
	u.win.SetMainMenu(u.menu)
	main := container.NewBorder(nil, panel, controls, nil, u.table)
	u.splashing = true
	u.splash = newSplashScreen(version, u.splashDone)
	u.win.SetContent(container.NewStack(main, u.splash.layer, u.tip.layer))

	u.updateButtons()
	u.splash.start()
	return u
}

// enterButton is a button that also runs on Enter. Fyne buttons only respond to Space, so a focused Start button
// ignored Enter. The wrapper is the owner, so that the button's own Disable and Enable still apply
type enterButton struct {
	*widget.Button
}

func newEnterButton(label string, tapped func()) *enterButton {
	b := &enterButton{Button: &widget.Button{Text: label, OnTapped: tapped}}
	b.ExtendBaseWidget(b)
	return b
}

// TypedKey runs the button on Enter, and leaves other keys to the Fyne button
func (b *enterButton) TypedKey(ev *fyne.KeyEvent) {
	if ev.Name == fyne.KeyReturn || ev.Name == fyne.KeyEnter {
		b.Tapped(nil)
		return
	}
	b.Button.TypedKey(ev)
}

// numberEntry creates an entry for a whole number. Changes are checked by inputChanged
func (u *ui) numberEntry(text, field, placeholder string) *helpEntry {
	e := newHelpEntry(model.InputHelp(field), u.tip)
	e.SetPlaceHolder(placeholder)
	e.SetText(text)
	e.OnChanged = func(string) { u.inputChanged() }
	return e
}

// field lays out a labelled input with a place for its validation message
func (u *ui) field(label string, e *helpEntry, field string) fyne.CanvasObject {
	errLabel := widget.NewLabel("")
	errLabel.Importance = widget.DangerImportance
	errLabel.Hide()
	u.fieldErrs[field] = errLabel
	return container.New(layout.NewCustomPaddedVBoxLayout(0), newHelpLabel(label, model.InputHelp(field), u.tip), e, errLabel)
}

// settings reads the inputs. The returned errors include whole-number checks and the model's limits
func (u *ui) settings() (model.Settings, model.FieldErrors) {
	s := model.Settings{Spacing: u.spacing.Checked, Transpose: u.transpose.Checked}
	errs := model.FieldErrors{}
	read := func(e *helpEntry, field string, dst *int) {
		v, err := strconv.Atoi(strings.TrimSpace(e.Text))
		if err != nil {
			errs[field] = "must be a whole number"
			return
		}
		*dst = v
	}
	read(u.fixtures, model.FieldFixtures, &s.Fixtures)
	read(u.players, model.FieldPlayers, &s.Players)
	read(u.perFixture, model.FieldPlayersPerFixture, &s.PlayersPerFixture)
	read(u.maxTogether, model.FieldMaxPlayersTogether, &s.MaxPlayersTogether)
	read(u.attempts, model.FieldAttempts, &s.Attempts)
	for field, msg := range s.Validate() {
		if _, ok := errs[field]; !ok {
			errs[field] = msg
		}
	}
	return s, errs
}

// inputChanged runs when a value other than transpose changes. The matrix described no longer applies, so it is cleared
func (u *ui) inputChanged() {
	_, errs := u.settings()
	for field, label := range u.fieldErrs {
		label.SetText(errs[field])
		if errs[field] == "" {
			label.Hide()
		} else {
			label.Show()
		}
	}
	if u.win.Content() != nil {
		u.win.Content().Refresh()
	}
	if !u.running {
		u.result = nil
		u.panel.SetText("")
		u.redraw()
	}
	u.updateButtons()
}

// transposeChanged redraws the current matrix in the other orientation, and the CSV follows the screen
func (u *ui) transposeChanged() {
	u.redraw()
	u.updateButtons()
}

// updateButtons sets which buttons and inputs can be used for the current state
func (u *ui) updateButtons() {
	_, errs := u.settings()
	if u.splashing {
		for _, in := range u.inputs {
			in.Disable()
		}
		u.start.Disable()
		u.save.Disable()
		return
	}
	if u.running || len(errs) > 0 {
		u.start.Disable()
	} else {
		u.start.Enable()
	}
	if u.result != nil && !u.running {
		u.save.Enable()
	} else {
		u.save.Disable()
	}
	for _, in := range u.inputs {
		if u.running {
			in.Disable()
		} else {
			in.Enable()
		}
	}
}

// splashDone runs when the splash has faded out, and the controls can be used
func (u *ui) splashDone() {
	u.splashing = false
	u.updateButtons()
}

// setRunning shows or hides the progress indicator and cancel button, and enables or disables the inputs
func (u *ui) setRunning(on bool) {
	u.running = on
	if on {
		u.spinner.Show()
		u.cancel.Enable()
	} else {
		u.spinner.Hide()
		u.cancel.Disable()
		u.cancelling = false
		u.cancelRun = nil
	}
	u.updateButtons()
}

func (u *ui) onStart() {
	s, errs := u.settings()
	if len(errs) > 0 {
		return
	}
	u.result = nil
	u.redraw()
	u.panel.SetText("")
	u.status.SetText("")
	u.setRunning(true)

	ctx, cancel := context.WithCancel(context.Background())
	u.cancelRun = cancel
	d := distributor.NewDistributor(s.Fixtures, s.Players, s.PlayersPerFixture, s.Attempts, s.MaxPlayersTogether, s.Spacing)

	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				made := d.AttemptsMade()
				fyne.Do(func() {
					if u.running && !u.cancelling {
						u.status.SetText(model.AttemptProgress(made, s.Attempts))
					}
				})
			}
		}
	}()
	go func() {
		res, err := d.Distribute(ctx)
		close(done)
		fyne.Do(func() { u.finish(s, d, res, err) })
	}()
}

// finish shows the outcome of a run on the UI thread
func (u *ui) finish(s model.Settings, d *distributor.Distributor, res *distributor.Distributor, err error) {
	u.setRunning(false)
	u.status.SetText("")
	switch {
	case errors.Is(err, context.Canceled):
		u.result = nil
		u.panel.SetText(model.CancelledMessage(d.AttemptsMade()))
	case err != nil:
		u.result = nil
		u.panel.SetText(err.Error())
	default:
		u.result = res
		u.panel.SetText(model.Summary(s, res))
	}
	u.redraw()
	u.updateButtons()
}

func (u *ui) onCancel() {
	if u.cancelRun == nil {
		return
	}
	u.cancelling = true
	u.cancel.Disable()
	u.status.SetText("Cancelling…")
	u.cancelRun()
}

// redraw rebuilds the displayed matrix from the current result, or clears it
func (u *ui) redraw() {
	if u.result == nil {
		u.rowLabels, u.colLabels, u.cells = nil, nil, nil
	} else {
		u.rowLabels, u.colLabels, u.cells = model.Cells(u.result.Matrix(), u.transpose.Checked)
	}
	u.table.Refresh()
}

func (u *ui) tableSize() (int, int) {
	return len(u.rowLabels) + 1, len(u.colLabels) + 1
}

func (u *ui) tableCell() fyne.CanvasObject {
	l := widget.NewLabel("")
	l.Alignment = fyne.TextAlignCenter
	return l
}

func (u *ui) tableUpdate(id widget.TableCellID, obj fyne.CanvasObject) {
	l := obj.(*widget.Label)
	switch {
	case id.Row == 0 && id.Col == 0:
		l.SetText("")
	case id.Row == 0:
		l.SetText(u.colLabels[id.Col-1])
	case id.Col == 0:
		l.SetText(u.rowLabels[id.Row-1])
	default:
		l.SetText(u.cells[id.Row-1][id.Col-1])
	}
}

func (u *ui) onSave() {
	if u.result == nil {
		return
	}
	records := model.CSV(u.result.Matrix(), u.transpose.Checked)
	d := dialog.NewFileSave(func(wc fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(err, u.win)
			return
		}
		if wc == nil {
			return
		}
		defer wc.Close()
		if err := csv.NewWriter(wc).WriteAll(records); err != nil {
			dialog.ShowError(err, u.win)
		}
	}, u.win)
	d.SetFileName("distribution.csv")
	d.Show()
}

func (u *ui) onAbout() {
	if u.splashing {
		return
	}
	text := fmt.Sprintf("Game Distributor %s\n\nCopyright (c) 2026 Ashley Kitson, UK\n\n"+
		"This software is licensed under the Functional Source License, Version 1.1 (FSL-1.1).\n"+
		"You may use, copy, modify, and distribute this software for non-competing purposes.\n\n"+
		"After two years from the release date of a specific version, that version automatically converts to the "+
		"Apache License, Version 2.0.", u.version)
	body := widget.NewLabel(text)
	body.Wrapping = fyne.TextWrapWord
	link, err := url.Parse(model.ReleaseURL(u.version))
	if err != nil {
		dialog.ShowError(err, u.win)
		return
	}
	dialog.NewCustom("About", "Close", container.NewVBox(body, widget.NewHyperlink(link.String(), link)), u.win).Show()
}

// themeChoices are the theme options shown in the Theme dialog, with the label the user sees for each
var themeChoices = []struct{ code, label string }{
	{model.ThemeLight, "Light"},
	{model.ThemeDark, "Dark"},
	{model.ThemeSystem, "System"},
}

// applyTheme sets the Fyne theme for the app. System follows the operating system's light or dark setting
func applyTheme(a fyne.App, code string) {
	switch code {
	case model.ThemeDark:
		a.Settings().SetTheme(newVariantTheme(theme.VariantDark))
	case model.ThemeSystem:
		a.Settings().SetTheme(theme.DefaultTheme())
	default:
		a.Settings().SetTheme(newVariantTheme(theme.VariantLight))
	}
}

// onTheme opens the Theme dialog. OK applies and saves the chosen theme. Cancel changes nothing
func (u *ui) onTheme() {
	if u.splashing {
		return
	}
	current := model.ThemeFromPreference(u.app.Preferences().String(model.PrefTheme))
	labels := make([]string, len(themeChoices))
	selected := ""
	for i, c := range themeChoices {
		labels[i] = c.label
		if c.code == current {
			selected = c.label
		}
	}
	group := widget.NewRadioGroup(labels, nil)
	group.SetSelected(selected)
	dialog.NewCustomConfirm("Theme", "OK", "Cancel", group, func(ok bool) {
		if !ok {
			return
		}
		for _, c := range themeChoices {
			if c.label == group.Selected {
				u.app.Preferences().SetString(model.PrefTheme, c.code)
				applyTheme(u.app, c.code)
			}
		}
	}, u.win).Show()
}

// onHelp opens the desktop app guide for this version in the default browser. A failure is shown in a dialog
func (u *ui) onHelp() {
	if u.splashing {
		return
	}
	link, err := url.Parse(model.HelpURL(u.version))
	if err != nil {
		dialog.ShowError(err, u.win)
		return
	}
	if err := u.app.OpenURL(link); err != nil {
		dialog.ShowError(err, u.win)
	}
}

// inputsActive reports whether the inputs can respond. They do not while the splash is showing
func (u *ui) inputsActive() bool {
	return !u.splashing
}
