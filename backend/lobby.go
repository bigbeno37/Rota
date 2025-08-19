package main

type Lobby struct {
	LobbyId string
	Player1 string
	Player2 *string
	Game    *Game
}

type LobbyEvent string

const (
	GameUpdate   LobbyEvent = "GAME_UPDATE"
	OpponentLeft            = "OPPONENT_LEFT"
)

type LobbyEventMessage struct {
	Event LobbyEvent
	Game  *Game
}
