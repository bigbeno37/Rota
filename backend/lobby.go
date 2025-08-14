package main

import (
	"log/slog"
	"sync"
)

type Lobby struct {
	mu      sync.RWMutex
	LobbyId string
	Player1 string
	Player2 *string
	Game    *Game
}

type LobbyManager struct {
	lobby  *Lobby
	logger *slog.Logger
}

func NewLobbyManager(lobby *Lobby, logger *slog.Logger) *LobbyManager {
	return &LobbyManager{
		lobby:  lobby,
		logger: logger,
	}
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

func (manager *LobbyManager) Player1() string {
	manager.RLock()
	defer manager.RUnlock()

	return manager.lobby.Player1
}

func (manager *LobbyManager) SetPlayer1(player1 string) {
	manager.Lock()
	defer manager.Unlock()

	manager.lobby.Player1 = player1
}

func (manager *LobbyManager) Player2() *string {
	manager.RLock()
	defer manager.RUnlock()

	return manager.lobby.Player2
}

func (manager *LobbyManager) SetPlayer2(player2 *string) {
	manager.Lock()
	defer manager.Unlock()

	manager.lobby.Player2 = player2
}

func (manager *LobbyManager) Game() *Game {
	manager.RLock()
	defer manager.RUnlock()

	return manager.lobby.Game
}

func (manager *LobbyManager) SetGame(game *Game) {
	manager.Lock()
	defer manager.Unlock()

	manager.lobby.Game = game
}

func (manager *LobbyManager) Lock() {
	manager.lobby.mu.Lock()
}

func (manager *LobbyManager) RLock() {
	manager.lobby.mu.Lock()
}

func (manager *LobbyManager) Unlock() {
	manager.lobby.mu.Unlock()
}

func (manager *LobbyManager) RUnlock() {
	manager.lobby.mu.Unlock()
}

func (manager *LobbyManager) Broadcast(message *LobbyEventMessage) {
	lobby, log := manager, manager.logger

	state := NewGlobalStateManager(log)
	player1 := state.GetPlayerWithId(lobby.Player1())

	var player2 *Player
	if lobby.Player2() != nil {
		player2 = state.GetPlayerWithId(*lobby.Player2())
		player2.SendMessage(message, log)
	}

	player1.SendMessage(message, log)
}
