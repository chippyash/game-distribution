package models

type DistOutput struct {
	NumFixtures        int     `json:"numFixtures"`
	NumPlayers         int     `json:"numPlayers"`
	PlayersPerFixture  int     `json:"playersPerFixture"`
	MaxPlayersTogether int     `json:"maxPlayersTogether"`
	FixtureSpacing     bool    `json:"fixtureSpacing"`
	Attempts           int     `json:"attempts"`
	MaxAttempts        int     `json:"maxAttempts"`
	PlayersMustPlay    int     `json:"playersMustPlay"`
	RemainderSlots     int     `json:"remainderSlots"`
	Distribution       [][]int `json:"distribution"`
}
