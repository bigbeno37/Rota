package main

import (
	"github.com/gorilla/websocket"
	"sync"
)

var (
	lobbiesMu sync.RWMutex
	lobbies   = map[string]*Lobby{}

	playersMu sync.RWMutex
	players   = map[string]*websocket.Conn{}
)

func CreateLobby(lobbyId string, lobby *Lobby) {
	lobbiesMu.Lock()
	defer lobbiesMu.Unlock()

	lobbies[lobbyId] = lobby
}

func GetLobbies() map[string]*Lobby {
	lobbiesMu.RLock()
	defer lobbiesMu.RUnlock()

	return lobbies
}

func GetLobbyWithPlayerId(playerId string) *Lobby {
	lobbiesMu.RLock()
	defer lobbiesMu.RUnlock()

	for _, lobby := range lobbies {
		if lobby.Player1 == playerId || (lobby.Player2 != nil && *lobby.Player2 == playerId) {
			return lobby
		}
	}
	return nil
}

func RemoveLobby(lobbyId string) {
	lobbiesMu.Lock()
	defer lobbiesMu.Unlock()

	lobbies[lobbyId] = nil
}

func GetPlayer(playerId string) *websocket.Conn {
	playersMu.RLock()
	defer playersMu.RUnlock()

	return players[playerId]
}
