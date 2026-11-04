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
	"image"
	"testing"
)

// TestEmbeddedImagesDecode checks the embedded window icon and splash image are present and decode as images
func TestEmbeddedImagesDecode(t *testing.T) {
	for name, data := range map[string][]byte{"window icon": windowIcon, "splash": splashImage} {
		if len(data) == 0 {
			t.Errorf("%s is empty", name)
			continue
		}
		if _, _, err := image.Decode(bytes.NewReader(data)); err != nil {
			t.Errorf("%s does not decode: %v", name, err)
		}
	}
}
