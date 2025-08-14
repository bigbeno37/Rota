package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"os"
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
	mainMux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mainMux.HandleFunc("/ws", wsHandler)
	mainMux.Handle(
		"/",
		CreateStack(getCookieMiddleware)(authenticatedMux),
	)

	port := "8080"
	if portNum, exists := os.LookupEnv("PORT"); exists == true {
		port = portNum
	}

	fmt.Println("WebSocket server listening on port " + port)
	log.Fatal(http.ListenAndServe(
		"0.0.0.0:"+port,
		mainMux,
	))
}
