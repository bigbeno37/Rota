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
	http.HandleFunc("POST /api/create-lobby", createLobbyHandler)
	http.HandleFunc("POST /api/join-lobby", joinLobbyHandler)
	http.HandleFunc("/api/make-move", makeMoveHandler)
	http.HandleFunc("/ws", wsHandler)

	fmt.Println("WebSocket server listening on port 8080!")
	log.Fatal(http.ListenAndServe(
		":8080",
		CreateStack(
			getCookieMiddleware,
			//MyMiddleware2,
		)(http.DefaultServeMux),
	))
}
