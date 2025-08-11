package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	authenticatedMux := http.NewServeMux()
	authenticatedMux.HandleFunc("POST /api/create-lobby", createLobbyHandler)
	authenticatedMux.HandleFunc("POST /api/join-lobby", joinLobbyHandler)
	authenticatedMux.HandleFunc("POST /api/leave-lobby", leaveLobbyHandler)
	authenticatedMux.HandleFunc("POST /api/make-move", makeMoveHandler)

	mainMux := http.NewServeMux()
	mainMux.HandleFunc("/ws", wsHandler)
	mainMux.Handle(
		"/",
		CreateStack(getCookieMiddleware)(authenticatedMux),
	)

	fmt.Println("WebSocket server listening on port 8080!")
	log.Fatal(http.ListenAndServe(
		":8080",
		mainMux,
	))
}
