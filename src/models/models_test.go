package models_test

import (
	"encoding/json"
	"testing"

	"github.com/chippyash/gamedistributor/src/models"
)

func TestNewGameDistInputDefaults(t *testing.T) {
	in := models.NewGameDistInput()
	if in.NumPlayers != 0 || in.NumFixtures != 0 || in.PlayersPerFixture != 0 {
		t.Errorf("counts: got %+v, want zero values", in)
	}
	if in.Attempts != 1000 {
		t.Errorf("Attempts: got %d, want 1000", in.Attempts)
	}
	if in.FixtureSpacing {
		t.Errorf("FixtureSpacing: got true, want false")
	}
	if in.MaxPlayerTogether != 0 {
		t.Errorf("MaxPlayerTogether: got %d, want 0", in.MaxPlayerTogether)
	}
}

func TestGameDistInputJSONKeys(t *testing.T) {
	var in models.GameDistInput
	body := []byte(`{"numPlayers":7,"numFixtures":24,"playersPerFixture":3,"attempts":500,"fixtureSpacing":true,"maxPlayerTogether":4}`)
	if err := json.Unmarshal(body, &in); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if in.NumPlayers != 7 || in.NumFixtures != 24 || in.PlayersPerFixture != 3 ||
		in.Attempts != 500 || !in.FixtureSpacing || in.MaxPlayerTogether != 4 {
		t.Errorf("unmarshalled %+v does not match input", in)
	}
}

func TestDistOutputJSONKeys(t *testing.T) {
	out := models.DistOutput{
		NumFixtures:        24,
		NumPlayers:         7,
		PlayersPerFixture:  3,
		MaxPlayersTogether: 4,
		FixtureSpacing:     true,
		Attempts:           12,
		MaxAttempts:        1000,
		PlayersMustPlay:    10,
		RemainderSlots:     2,
		Distribution:       [][]int{{1, 0}, {0, 1}},
	}
	raw, err := json.Marshal(out)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var got map[string]interface{}
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	want := map[string]interface{}{
		"numFixtures":        24.0,
		"numPlayers":         7.0,
		"playersPerFixture":  3.0,
		"maxPlayersTogether": 4.0,
		"fixtureSpacing":     true,
		"attempts":           12.0,
		"maxAttempts":        1000.0,
		"playersMustPlay":    10.0,
		"remainderSlots":     2.0,
		"distribution":       []interface{}{[]interface{}{1.0, 0.0}, []interface{}{0.0, 1.0}},
	}
	if len(got) != len(want) {
		t.Errorf("got %d keys, want %d: %v", len(got), len(want), got)
	}
	for k, v := range want {
		if g, ok := got[k]; !ok {
			t.Errorf("missing key %q", k)
		} else if !jsonEqual(g, v) {
			t.Errorf("key %q: got %v, want %v", k, g, v)
		}
	}
}

func jsonEqual(a, b interface{}) bool {
	x, _ := json.Marshal(a)
	y, _ := json.Marshal(b)
	return string(x) == string(y)
}
