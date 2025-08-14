package main

import (
	"fmt"
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
	fmt.Println("GetLobbyWithPlayerId: Attempting to get read lock on global lobbies...")
	lobbiesMu.RLock()
	fmt.Println("GetLobbyWithPlayerId: Read lock acquired!")
	defer lobbiesMu.RUnlock()
	defer fmt.Println("GetLobbyWithPlayerId: Read lock released")

	fmt.Println("GetLobbyWithPlayerId: Searching lobbies...")
	for _, lobby := range lobbies {
		if lobby.Player1 == playerId {
			fmt.Println("GetLobbyWithPlayerId: Found lobby " + lobby.LobbyId + " with player " + playerId)
			return lobby
		}

		if lobby.Player2 != nil {
			if *lobby.Player2 == playerId {
				fmt.Println("GetLobbyWithPlayerId: Found lobby " + lobby.LobbyId + " with player " + playerId)
				return lobby
			}
		}
	}

	fmt.Println("GetLobbyWithPlayerId: No lobbies found. Returning nil...")
	return nil
}

func RemoveLobby(lobbyId string) {
	lobbiesMu.Lock()
	defer lobbiesMu.Unlock()

	delete(lobbies, lobbyId)
}

func GetPlayer(playerId string) *websocket.Conn {
	fmt.Println("GetPlayer: Attempting to get read lock on global players...")
	playersMu.RLock()
	fmt.Println("GetPlayer: Read lock obtained!")
	defer playersMu.RUnlock()
	defer fmt.Println("GetPlayer: Read lock released")

	return players[playerId]
}
