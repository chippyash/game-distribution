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
	"bytes"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"time"

	_ "embed"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

const (
	splashHold = 2 * time.Second // how long the splash shows before it starts to fade
	splashFade = 3 * time.Second // how long the fade takes
)

// splashImage is the splash screen, embedded so the binary does not depend on the file's location
//
//go:embed resources/splash.jpeg
var splashImage []byte

// splashScreen is an overlay that covers the main window on launch, then fades out
type splashScreen struct {
	layer    *fyne.Container
	img      *canvas.Image
	bg       *canvas.Rectangle
	texts    []*canvas.Text
	anim     *fyne.Animation
	done     func()
	timer    *time.Timer
	finished bool
}

// newSplashScreen builds the overlay, with the version and copyright under the image. done is called once the splash has been removed
func newSplashScreen(version string, done func()) *splashScreen {
	img := canvas.NewImageFromResource(fyne.NewStaticResource("splash.jpeg", splashImage))
	img.FillMode = canvas.ImageFillContain
	bg := canvas.NewRectangle(color.Black)

	// The disabled button sits under the image and takes all taps, so the controls underneath get no input
	blocker := widget.NewButton("", nil)
	blocker.Disable()

	versionText := canvas.NewText(fmt.Sprintf("Version %s", version), color.White)
	copyrightText := canvas.NewText("Copyright Ashley Kitson, UK 2026", color.White)
	texts := []*canvas.Text{versionText, copyrightText}
	for _, txt := range texts {
		txt.Alignment = fyne.TextAlignCenter
	}
	caption := container.NewVBox(versionText, copyrightText)

	s := &splashScreen{
		layer: container.New(newSplashLayout(), blocker, bg, img, caption),
		img:   img,
		bg:    bg,
		texts: texts,
		done:  done,
	}
	s.anim = fyne.NewAnimation(splashFade, s.tick)
	return s
}

// start shows the splash for splashHold, then fades it out
func (s *splashScreen) start() {
	s.timer = time.AfterFunc(splashHold, func() {
		fyne.Do(func() {
			if !s.finished {
				s.anim.Start()
			}
		})
	})
}

// tick sets the fade for progress f, where 0 is fully shown and 1 is fully faded out
func (s *splashScreen) tick(f float32) {
	if f >= 1 {
		s.finish()
		return
	}
	s.img.Translucency = float64(f)
	s.img.Refresh()
	s.bg.FillColor = color.NRGBA{A: uint8(255 * (1 - f))}
	s.bg.Refresh()
	for _, txt := range s.texts {
		txt.Color = color.NRGBA{R: 255, G: 255, B: 255, A: uint8(255 * (1 - f))}
		txt.Refresh()
	}
}

// finish removes the splash from the window and tells the main window it can take input
func (s *splashScreen) finish() {
	if s.finished {
		return
	}
	s.finished = true
	s.timer.Stop()
	s.layer.Hide()
	s.done()
}

// splashMargin is the gap between the bottom of the image and the caption
const splashMargin = 12

// splashLayout sizes the image to fit the window with its own aspect ratio, and places the caption below the image
type splashLayout struct {
	aspect float32
}

// newSplashLayout reads the aspect ratio of the embedded splash image
func newSplashLayout() *splashLayout {
	cfg, _, err := image.DecodeConfig(bytes.NewReader(splashImage))
	if err != nil || cfg.Height == 0 {
		return &splashLayout{aspect: 1}
	}
	return &splashLayout{aspect: float32(cfg.Width) / float32(cfg.Height)}
}

// Layout expects the objects in the order: blocker, background, image, caption
func (l *splashLayout) Layout(objs []fyne.CanvasObject, size fyne.Size) {
	blocker, bg, img, caption := objs[0], objs[1], objs[2], objs[3]
	blocker.Move(fyne.NewPos(0, 0))
	blocker.Resize(size)
	bg.Move(fyne.NewPos(0, 0))
	bg.Resize(size)

	w := size.Width
	h := w / l.aspect
	if h > size.Height {
		h = size.Height
		w = h * l.aspect
	}
	pos := fyne.NewPos((size.Width-w)/2, (size.Height-h)/2)
	img.Move(pos)
	img.Resize(fyne.NewSize(w, h))

	capSize := caption.MinSize()
	caption.Resize(capSize)
	// below the bottom edge of the image, or at the bottom of the window if there is no room below the image
	capY := pos.Y + h + splashMargin
	if capY+capSize.Height > size.Height {
		capY = size.Height - capSize.Height
	}
	caption.Move(fyne.NewPos(pos.X+(w-capSize.Width)/2, capY))
}

// MinSize is zero, so the splash can be as small as the window
func (l *splashLayout) MinSize(objs []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(0, 0)
}
