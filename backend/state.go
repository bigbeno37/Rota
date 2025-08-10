package main

import "github.com/gorilla/websocket"

type Lobby struct {
	Player1 string
	Player2 string
	Game    *Game
}

func (lobby *Lobby) broadcastLobby() {
	player1, player2 := players[lobby.Player1], players[lobby.Player2]
	game := lobby.Game

	player1.WriteJSON(game)
	player2.WriteJSON(game)
}

var lobbies = map[string]*Lobby{}
var players = map[string]*websocket.Conn{}
