package models

type GameDistInput struct {
	NumPlayers        int  `json:"numPlayers"`
	NumFixtures       int  `json:"numFixtures"`
	PlayersPerFixture int  `json:"playersPerFixture"`
	Attempts          int  `json:"attempts,omitempty"`
	FixtureSpacing    bool `json:"fixtureSpacing,omitempty"`
	MaxPlayerTogether int  `json:"maxPlayerTogether,omitempty"`
}

func NewGameDistInput() *GameDistInput {
	return &GameDistInput{
		NumPlayers:        0,
		NumFixtures:       0,
		PlayersPerFixture: 0,
		Attempts:          1000,
		FixtureSpacing:    false,
		MaxPlayerTogether: 0,
	}
}
