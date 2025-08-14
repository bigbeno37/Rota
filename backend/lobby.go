package main

import (
	"fmt"
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
	fmt.Println("SetPlayer2: Locking lobby...")
	lobby.mu.Lock()
	fmt.Println("SetPlayer2: Lobby locked!")
	defer lobby.mu.Unlock()
	defer fmt.Println("SetPlayer2: Lobby lock released")

	lobby.Player2 = &player2Id
	fmt.Println("SetPlayer2: Set Player2 to " + player2Id)
}

func (lobby *Lobby) SetGame(game *Game) {
	fmt.Println("SetGame: Locking lobby...")
	lobby.mu.Lock()
	fmt.Println("SetGame: Lobby locked!")
	defer lobby.mu.Unlock()
	defer fmt.Println("SetGame: Lobby unlocked")

	lobby.Game = game
	fmt.Println("SetGame: Lobby game has been set")
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
	fmt.Println("Broadcast: Getting read lock on lobby...")
	lobby.mu.RLock()
	fmt.Println("Broadcast: Read lock acquired!")
	defer lobby.mu.RUnlock()
	defer fmt.Println("Broadcast: Read lock released")

	player1 := players[lobby.Player1]

	var player2 *websocket.Conn
	if lobby.Player2 != nil {
		player2 = players[*lobby.Player2]
		fmt.Println("Broadcast: Broadcasting to player 2")
		player2.WriteJSON(message)
	}

	fmt.Println("Broadcast: Broadcasting to player 1")
	player1.WriteJSON(message)
}

func (lobby *Lobby) BroadcastGameUpdate() {
	fmt.Println("BroadcastGameUpdate: Creating new LobbyEventMessage...")
	message := LobbyEventMessage{
		Event: GameUpdate,
		Game:  lobby.Game,
	}

	fmt.Println("BroadcastGameUpdate: Broadcasting message... ")
	lobby.Broadcast(&message)
}
