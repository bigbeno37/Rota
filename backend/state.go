package main

import (
	"log/slog"
	"sync"
)

var (
	lobbiesMu sync.RWMutex
	lobbies   = map[string]*Lobby{}

	playersMu sync.RWMutex
	players   = map[string]*Player{}
)

type GlobalStateManager struct {
	logger *slog.Logger
}

func NewGlobalStateManager(logger *slog.Logger) *GlobalStateManager {
	return &GlobalStateManager{
		logger: logger,
	}
}

func (manager *GlobalStateManager) CreateLobby(lobbyId string, lobby *Lobby) {
	log := manager.logger

	lobbiesMu.Lock()
	defer lobbiesMu.Unlock()

	lobbies[lobbyId] = lobby
	log.Debug("Lobby " + lobbyId + " now active with user " + lobby.Player1)
}

func (manager *GlobalStateManager) CreatePlayer(id string, player *Player) {
	log := manager.logger

	playersMu.Lock()
	defer playersMu.Unlock()

	players[id] = player
	log.Debug("Created player with id " + id)
}

func (manager *GlobalStateManager) RemoveLobby(lobbyId string) {
	log := manager.logger

	lobbiesMu.Lock()
	defer lobbiesMu.Unlock()

	delete(lobbies, lobbyId)
	log.Debug("Lobby " + lobbyId + " removed from lobbies")
}

func (manager *GlobalStateManager) GetLobbyWithId(lobbyId string) *Lobby {
	lobbiesMu.RLock()
	defer lobbiesMu.RUnlock()

	return lobbies[lobbyId]
}

func (manager *GlobalStateManager) GetPlayerWithId(playerId string) *Player {
	playersMu.RLock()
	defer playersMu.RUnlock()

	return players[playerId]
}
