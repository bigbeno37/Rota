package main

import (
	"backend/player"
	"backend/position"
)

type GameState string

const (
	Setup    GameState = "SETUP"
	Playing            = "PLAYING"
	GameOver           = "GAME_OVER"
)

type Game struct {
	State GameState
	Turn  player.Player
	Board []position.Position
}

func NewGame() *Game {
	return &Game{
		State: Setup,
		Turn:  player.Player1,
		Board: []position.Position{
			position.Empty,
			position.Empty,
			position.Empty,
			position.Empty,
			position.Empty,
			position.Empty,
			position.Empty,
			position.Empty,
			position.Empty,
		},
	}
}
