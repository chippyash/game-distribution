// Copyright (c) 2026 Ashley Kitson, UK
//
// This software is licensed under the Functional Source License, Version 1.1 (FSL-1.1).
// You may use, copy, modify, and distribute this software for non-competing purposes.
//
// After two years from the release date of a specific version, that version
// automatically converts to the Apache License, Version 2.0.
//
// See the LICENSE file in the project root for full license terms

package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// gamedistBin is the CLI built once for all tests in this file
var gamedistBin string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "gamedist-cli-test")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	gamedistBin = filepath.Join(dir, "gamedist")
	if out, err := exec.Command("go", "build", "-o", gamedistBin, ".").CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "build failed: %v\n%s", err, out)
		os.RemoveAll(dir)
		os.Exit(1)
	}
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

// runCLI runs the built CLI in a fresh directory and returns its output and the distribution.csv it wrote, if any
func runCLI(t *testing.T, args ...string) (string, [][]string) {
	t.Helper()
	dir := t.TempDir()
	cmd := exec.Command(gamedistBin, append(args, "-c")...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("gamedist %v: %v\n%s", args, err, out)
	}
	f, err := os.Open(filepath.Join(dir, "distribution.csv"))
	if err != nil {
		t.Fatalf("gamedist %v wrote no csv: %v\n%s", args, err, out)
	}
	defer f.Close()
	records, err := csv.NewReader(f).ReadAll()
	if err != nil {
		t.Fatalf("reading csv: %v", err)
	}
	return string(out), records
}

// backToBack counts player appearances in consecutive fixtures from a distribution.csv
// Rows are players (after the header row) and columns are fixtures (after the first column)
func backToBack(records [][]string) int {
	var n int
	for _, row := range records[1:] {
		for f := 1; f+1 < len(row); f++ {
			if row[f] == "1" && row[f+1] == "1" {
				n++
			}
		}
	}
	return n
}

func TestSpacingFlagAbsentReportsAutospacingFalse(t *testing.T) {
	out, _ := runCLI(t, "24", "7", "3", "5", "-a", "100000")
	if !strings.Contains(out, "Autospacing is false") {
		t.Errorf("output does not report Autospacing is false:\n%s", out)
	}
}

func TestSpacingFlagReportsAutospacingTrue(t *testing.T) {
	out, _ := runCLI(t, "24", "7", "3", "5", "-s", "-a", "100000")
	if !strings.Contains(out, "Autospacing is true") {
		t.Errorf("output does not report Autospacing is true:\n%s", out)
	}
}

func TestSpacingFlagSpreadsFixtures(t *testing.T) {
	// same settings with and without -s, averaged to avoid relying on one random run
	const runs = 3
	var off, on int
	for i := 0; i < runs; i++ {
		_, offRecords := runCLI(t, "24", "7", "3", "5", "-a", "100000")
		_, onRecords := runCLI(t, "24", "7", "3", "5", "-s", "-a", "100000")
		off += backToBack(offRecords)
		on += backToBack(onRecords)
	}
	if on*4 > off {
		t.Errorf("with -s there were %d back-to-back appearances over %d runs, without -s there were %d; want -s well below", on, runs, off)
	}
}

func TestSpacingLimitedReportsAutospacingTrueWithNote(t *testing.T) {
	// 5 players for 3 per fixture is fewer than twice the players per fixture, so spacing cannot be complete
	out, _ := runCLI(t, "9", "5", "3", "-s", "-a", "100000")
	want := "Autospacing is true (back-to-back fixtures cannot be fully avoided: fewer than twice as many players as players per fixture)"
	if !strings.Contains(out, want) {
		t.Errorf("output does not report limited spacing:\n%s", out)
	}
}

func TestSpacingNotRequestedHasNoLimitNote(t *testing.T) {
	out, _ := runCLI(t, "9", "5", "3", "-a", "100000")
	if !strings.Contains(out, "Autospacing is false\n") {
		t.Errorf("output should report Autospacing is false with no note:\n%s", out)
	}
}

func TestSpacingFeasibleHasNoNote(t *testing.T) {
	out, _ := runCLI(t, "24", "7", "3", "5", "-s", "-a", "100000")
	if strings.Contains(out, "not feasible") {
		t.Errorf("feasible spacing should have no note:\n%s", out)
	}
}
