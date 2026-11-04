package main

import (
	_ "embed"
	"encoding/csv"
	"fmt"
	"github.com/chippyash/gamedistributor/src/distributor"
	"github.com/thatisuday/commando"
	"gonum.org/v1/gonum/mat"
	"html/template"
	"os"
	"strconv"
)

const (
	MaxAllowedAttempts = 100000
	MinAllowedAttempts = 1000
)

//go:embed VERSION
var version string

type cmdFunc func(args map[string]commando.ArgValue, flags map[string]commando.FlagValue)

func init() {
	commando.
		SetExecutableName("gamedist").
		SetDescription("Utility to provide player permutation across fixtures").
		SetVersion(version)
	commando.Register("").
		SetShortDescription("Generate a permutation of players across fixtures").
		AddArgument("numFixtures", "Number of fixtures: int", "").
		AddArgument("numPlayers", "Number of players: int", "").
		AddArgument("playersPerFixture", "Players per fixture: int", "").
		AddArgument("maxPlayersTogether", "Max number of fixtures any two players may share: int. Raised to the minimum possible (average number of shared fixtures per pair, rounded up) if lower", "1").
		AddFlag("attempts,a", fmt.Sprintf("Number of attempts to be made (%d - %d): int", MinAllowedAttempts, MaxAllowedAttempts), commando.Int, MinAllowedAttempts).
		AddFlag("fixtureSpacing,s", "If true, the fixtures will be spaced evenly for each player", commando.Bool, true).
		AddFlag("transpose,t", "If true, the resultant distribution matrix will be cols==players, rows==players", commando.Bool, false).
		AddFlag("csv,c", "If true, save a csv file of the distribution to distribution.csv", commando.Bool, false).
		SetAction(cmd())
}

func cmd() cmdFunc {
	return func(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
		numFixtures, err := strconv.Atoi(args["numFixtures"].Value)
		if err != nil {
			println("numFixtures must be an integer")
			return
		}
		numPlayers, err := strconv.Atoi(args["numPlayers"].Value)
		if err != nil {
			println("numPlayers must be an integer")
			return
		}
		playersPerFixture, err := strconv.Atoi(args["playersPerFixture"].Value)
		if err != nil {
			println("playersPerFixture must be an integer")
			return
		}
		maxPlayersTogether, err := strconv.Atoi(args["maxPlayersTogether"].Value)
		if err != nil {
			println("maxPlayersTogether must be an integer")
			return
		}
		fixtureSpacing, err := flags["fixtureSpacing"].GetBool()
		if err != nil {
			println("fixtureSpacing must be a boolean")
			return
		}
		fixtureSpacing = !fixtureSpacing //invert so default is true

		attempts, err := flags["attempts"].GetInt()
		if err != nil {
			println("attempts must be an integer")
			return
		}
		if attempts < MinAllowedAttempts || attempts > MaxAllowedAttempts {
			fmt.Printf("attempts must be between %d and %d\n", MinAllowedAttempts, MaxAllowedAttempts)
			return
		}

		transpose, err := flags["transpose"].GetBool()
		if err != nil {
			println("transpose must be a boolean")
			return
		}

		dist, err := distributor.NewDistributor(numFixtures, numPlayers, playersPerFixture, attempts, maxPlayersTogether, fixtureSpacing).Distribute()
		if err != nil {
			println(err.Error())
			return
		}

		var layoutExplanation string
		var matrix mat.Matrix
		if transpose {
			layoutExplanation = "Rows == fixtures, Columns == players"
			matrix = dist.Matrix()
		} else {
			layoutExplanation = "Rows == players, Columns == fixtures"
			matrix = mat.Transpose{Matrix: dist.Matrix()}
		}
		matrixString := fmt.Sprintf("%v", mat.Formatted(matrix, mat.Prefix(""), mat.Squeeze()))

		type result struct {
			Players                  int
			Fixtures                 int
			PlayersPerFixture        int
			MustPlay                 int
			Remainder                int
			MaxPlayersTogether       int
			RequestedPlayersTogether int
			Autospacing              bool
			ActAttempts              int
			MaxAttempts              int
			Layout                   string
			Matrix                   string
		}
		r := result{
			Players:                  numPlayers,
			Fixtures:                 numFixtures,
			PlayersPerFixture:        playersPerFixture,
			MustPlay:                 dist.MustPlay(),
			Remainder:                dist.RemainderSlots(),
			MaxPlayersTogether:       dist.MaxPlayersTogether(),
			RequestedPlayersTogether: maxPlayersTogether,
			Autospacing:              fixtureSpacing,
			ActAttempts:              dist.Attempts() + 1,
			MaxAttempts:              attempts,
			Layout:                   layoutExplanation,
			Matrix:                   matrixString,
		}
		tpl := `
Players: {{ .Players }}, Fixtures: {{ .Fixtures }}, Players per fixture: {{ .PlayersPerFixture }}
Players must play: {{ .MustPlay }} fixtures
{{ .Remainder }} Players will play an extra fixture
Autospacing is {{ .Autospacing }}
Max Play Together = {{ .MaxPlayersTogether }}{{ if ne .MaxPlayersTogether .RequestedPlayersTogether }} (requested {{ .RequestedPlayersTogether }} is below the minimum possible for these settings){{ end }}
{{ .Layout }}
Result = 
{{ .Matrix }}

Attempt {{ .ActAttempts }} of {{ .MaxAttempts }}
`
		t, err := template.New("result").Parse(tpl)
		if err != nil {
			println(err.Error())
			return
		}
		if err := t.Execute(os.Stdout, r); err != nil {
			println(err.Error())
			return
		}

		if f, _ := flags["csv"].GetBool(); f {
			saveCsv(dist, transpose)
		}
	}
}

func saveCsv(dist *distributor.Distributor, transpose bool) {
	var matrix mat.Matrix
	var colPrefix, rowPrefix string
	if transpose {
		colPrefix = "P%d"
		rowPrefix = "F%d"
		matrix = dist.Matrix()
	} else {
		colPrefix = "F%d"
		rowPrefix = "P%d"
		matrix = mat.Transpose{Matrix: dist.Matrix()}
	}
	numRows, numCols := matrix.Dims()
	var records = make([][]string, numRows+1)
	records[0] = make([]string, numCols+1)
	records[0][0] = ""
	// Fill column headers
	for i := 0; i < numCols; i++ {
		records[0][i+1] = fmt.Sprintf(colPrefix, i+1)
	}
	// Fill rows
	for i := 0; i < numRows; i++ {
		records[i+1] = make([]string, numCols+1)
		records[i+1][0] = fmt.Sprintf(rowPrefix, i+1)
		for j := 0; j < numCols; j++ {
			records[i+1][j+1] = fmt.Sprintf("%d", int(matrix.At(i, j)))
		}
	}
	f, err := os.Create("distribution.csv")
	if err != nil {
		println(err.Error())
		return
	}
	defer f.Close()

	w := csv.NewWriter(f)
	if err := w.WriteAll(records); err != nil {
		println(err.Error())
		return
	}

	println("CSV saved to distribution.csv")
}

func main() {
	commando.Parse(nil)
}
