package main

import (
	"github.com/gorilla/websocket"
	"log/slog"
	"sync"
)

type Player struct {
	mu           sync.RWMutex
	id           string
	conn         *websocket.Conn
	currentLobby *string
}

func NewPlayer(id string, conn *websocket.Conn) *Player {
	return &Player{
		id:   id,
		conn: conn,
	}
}

func (p *Player) SendMessage(m interface{}, log *slog.Logger) {
	p.mu.Lock()
	defer p.mu.Unlock()

	log.Info("Broadcasting message to player " + p.id + "...")
	p.conn.WriteJSON(m)
}

func (p *Player) SetLobby(lobbyId *string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.currentLobby = lobbyId
}
