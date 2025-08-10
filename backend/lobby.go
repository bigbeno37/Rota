package main

import (
	"github.com/gorilla/websocket"
	"sync"
)

type Lobby struct {
	mu      sync.RWMutex
	LobbyId string
	Player1 string
	Player2 *string
	Game    *Game
}

func (lobby *Lobby) SetPlayer2(player2Id string) {
	lobby.mu.Lock()
	defer lobby.mu.Unlock()

	lobby.Player2 = &player2Id
}

func (lobby *Lobby) SetGame(game *Game) {
	lobby.mu.Lock()
	defer lobby.mu.Unlock()

	lobby.Game = game
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

func (lobby *Lobby) Broadcast(message *LobbyEventMessage) {
	lobby.mu.RLock()
	defer lobby.mu.RUnlock()

	player1 := players[lobby.Player1]

	var player2 *websocket.Conn
	if lobby.Player2 != nil {
		player2 = players[*lobby.Player2]
		player2.WriteJSON(message)
	}

	player1.WriteJSON(message)

}

func (lobby *Lobby) BroadcastGameUpdate() {
	message := LobbyEventMessage{
		Event: GameUpdate,
		Game:  lobby.Game,
	}

	lobby.Broadcast(&message)
}
