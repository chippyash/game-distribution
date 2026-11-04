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
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// hoverDelay is how long the pointer must rest on an input before its help appears
const hoverDelay = 1 * time.Second

const (
	tipGap        = 8   // gap between an input and its help
	tipWidth      = 300 // width of the help in pixels
	tipLineHeight = 22  // approximate height of one line of help text
	tipCharsLine  = 40  // approximate characters per line at tipWidth
)

// helpTip is the one help card for the window. It sits in the window content rather than in an overlay, because an
// overlay takes the keyboard away from the inputs
type helpTip struct {
	layer   *fyne.Container
	card    *widget.Card
	label   *widget.Label
	enabled func() bool
}

func newHelpTip(enabled func() bool) *helpTip {
	label := widget.NewLabel("")
	label.Wrapping = fyne.TextWrapWord
	card := widget.NewCard("", "", container.NewPadded(label))
	card.Hide()
	return &helpTip{
		layer:   container.NewWithoutLayout(card),
		card:    card,
		label:   label,
		enabled: enabled,
	}
}

// show places the help beside the owner, inside the window content, and shows it
func (t *helpTip) show(owner fyne.CanvasObject, text string) {
	drv := fyne.CurrentApp().Driver()
	at := drv.AbsolutePositionForObject(owner)
	base := drv.AbsolutePositionForObject(t.layer)
	size := tipSize(text)

	x := at.X - base.X + owner.Size().Width + tipGap
	y := at.Y - base.Y
	if max := t.layer.Size().Width - size.Width; x > max {
		x = max
	}
	if max := t.layer.Size().Height - size.Height; y > max {
		y = max
	}
	t.label.SetText(text)
	t.card.Resize(size)
	t.card.Move(fyne.NewPos(x, y))
	t.card.Show()
}

func (t *helpTip) hide() {
	t.card.Hide()
}

// tipSize returns a size that fits the help text on wrapped lines
func tipSize(text string) fyne.Size {
	lines := len(text)/tipCharsLine + 1
	return fyne.NewSize(tipWidth, float32(lines*tipLineHeight+24))
}

// hoverHelp shows an input's help after the pointer has rested on it for hoverDelay, and hides it when the pointer leaves
type hoverHelp struct {
	owner  fyne.CanvasObject
	text   string
	tip    *helpTip
	active bool
	timer  *time.Timer
}

func (h *hoverHelp) enter() {
	h.active = true
	h.stop()
	h.timer = time.AfterFunc(hoverDelay, func() {
		fyne.Do(h.show)
	})
}

func (h *hoverHelp) leave() {
	h.active = false
	h.stop()
	h.tip.hide()
}

func (h *hoverHelp) stop() {
	if h.timer != nil {
		h.timer.Stop()
		h.timer = nil
	}
}

// show displays the help if the pointer is still on the input and the window accepts input
func (h *hoverHelp) show() {
	if !h.active || !h.tip.enabled() {
		return
	}
	h.tip.show(h.owner, h.text)
}

// helpLabel is a label that shows the input's help text on hover
type helpLabel struct {
	*widget.Label
	help hoverHelp
}

// newHelpLabel builds the label with its wrapper as owner. The wrapper must own the widget, or focus and refresh
// go to the inner widget, which is not in the window's tree
func newHelpLabel(text, help string, tip *helpTip) *helpLabel {
	l := &helpLabel{Label: &widget.Label{}, help: hoverHelp{text: help, tip: tip}}
	l.ExtendBaseWidget(l)
	l.SetText(text)
	l.help.owner = l
	return l
}

func (l *helpLabel) MouseIn(*desktop.MouseEvent)    { l.help.enter() }
func (l *helpLabel) MouseMoved(*desktop.MouseEvent) {}
func (l *helpLabel) MouseOut()                      { l.help.leave() }

// helpEntry is an entry that shows its help text on hover
type helpEntry struct {
	*widget.Entry
	help hoverHelp
}

// newHelpEntry builds the entry with its wrapper as owner, so that tapping it can take focus and typing is shown
func newHelpEntry(help string, tip *helpTip) *helpEntry {
	e := &helpEntry{Entry: &widget.Entry{Wrapping: fyne.TextWrap(fyne.TextTruncateClip)}, help: hoverHelp{text: help, tip: tip}}
	e.ExtendBaseWidget(e)
	e.help.owner = e
	return e
}

func (e *helpEntry) MouseIn(*desktop.MouseEvent)    { e.help.enter() }
func (e *helpEntry) MouseMoved(*desktop.MouseEvent) {}
func (e *helpEntry) MouseOut()                      { e.help.leave() }

// MouseDown hides the help before the entry takes focus
func (e *helpEntry) MouseDown(ev *desktop.MouseEvent) {
	e.help.leave()
	e.Entry.MouseDown(ev)
}

// helpCheck is a check box that shows its help text on hover. Fyne's hit test returns the deepest hoverable widget
// under the pointer, so the check box itself must receive the pointer. A widget's owner is only set once, so the
// check box is built without NewCheck (which makes the check its own owner), and the wrapper becomes the owner
type helpCheck struct {
	*widget.Check
	help hoverHelp
}

func newHelpCheck(label, help string, changed func(bool), tip *helpTip) *helpCheck {
	c := &helpCheck{Check: &widget.Check{Text: label, OnChanged: changed}, help: hoverHelp{text: help, tip: tip}}
	c.ExtendBaseWidget(c)
	c.help.owner = c
	return c
}

func (c *helpCheck) MouseIn(e *desktop.MouseEvent) {
	c.Check.MouseIn(e)
	c.help.enter()
}

func (c *helpCheck) MouseMoved(e *desktop.MouseEvent) {
	c.Check.MouseMoved(e)
}

func (c *helpCheck) MouseOut() {
	c.Check.MouseOut()
	c.help.leave()
}

// Tapped hides the help before the check box changes
func (c *helpCheck) Tapped(e *fyne.PointEvent) {
	c.help.leave()
	c.Check.Tapped(e)
}
