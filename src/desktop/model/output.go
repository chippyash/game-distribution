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

import (
	"fmt"

	"gonum.org/v1/gonum/mat"
)

// Cells lays out the matrix for display. Each cell is "X" when the player is in the fixture, and blank otherwise.
// Without transpose, rows are players and columns are fixtures, as in the CLI. Labels follow the same orientation as
// the CSV.
func Cells(m mat.Matrix, transpose bool) (rowLabels, colLabels []string, cells [][]string) {
	grid, rowPrefix, colPrefix := orient(m, transpose)
	rows, cols := grid.Dims()
	rowLabels = labels(rowPrefix, rows)
	colLabels = labels(colPrefix, cols)
	cells = make([][]string, rows)
	for r := 0; r < rows; r++ {
		cells[r] = make([]string, cols)
		for c := 0; c < cols; c++ {
			if grid.At(r, c) == 1 {
				cells[r][c] = "X"
			}
		}
	}
	return rowLabels, colLabels, cells
}

// CSV returns the distribution as CSV records, in the same layout as the CLI's -c flag
func CSV(m mat.Matrix, transpose bool) [][]string {
	grid, rowPrefix, colPrefix := orient(m, transpose)
	rows, cols := grid.Dims()
	records := make([][]string, rows+1)
	records[0] = make([]string, cols+1)
	for c := 0; c < cols; c++ {
		records[0][c+1] = fmt.Sprintf("%s%d", colPrefix, c+1)
	}
	for r := 0; r < rows; r++ {
		records[r+1] = make([]string, cols+1)
		records[r+1][0] = fmt.Sprintf("%s%d", rowPrefix, r+1)
		for c := 0; c < cols; c++ {
			records[r+1][c+1] = fmt.Sprintf("%d", int(grid.At(r, c)))
		}
	}
	return records
}

// orient returns the matrix with rows and columns in the requested orientation, along with the label prefixes for them
func orient(m mat.Matrix, transpose bool) (mat.Matrix, string, string) {
	if transpose {
		return m, "F", "P"
	}
	return mat.Transpose{Matrix: m}, "P", "F"
}

func labels(prefix string, n int) []string {
	out := make([]string, n)
	for i := range out {
		out[i] = fmt.Sprintf("%s%d", prefix, i+1)
	}
	return out
}
