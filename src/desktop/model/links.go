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

import "fmt"

const (
	helpURLFormat    = "https://github.com/chippyash/game-distribution/blob/%s/README.md"
	releaseURLFormat = "https://github.com/chippyash/game-distribution/releases/tag/%s"
)

// HelpURL returns the link to the README for a version, for example v1.0.0
func HelpURL(version string) string {
	return fmt.Sprintf(helpURLFormat, version)
}

// ReleaseURL returns the link to the release documentation for a version, for example v1.0.0
func ReleaseURL(version string) string {
	return fmt.Sprintf(releaseURLFormat, version)
}
