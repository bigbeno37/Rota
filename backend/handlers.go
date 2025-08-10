package main

import (
	"backend/player"
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"strconv"
)

func createLobbyHandler(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value("id").(string)

	lobbyId := uuid.NewString()
	lobbies[lobbyId] = &Lobby{
		Player1: id,
	}

	w.Write([]byte(lobbyId))
}

func joinLobbyHandler(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value("id").(string)

	if !r.URL.Query().Has("lobbyId") {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Invalid Lobby ID!"))
		return
	}

	lobby := lobbies[r.URL.Query().Get("lobbyId")]

	if lobby == nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Invalid lobby ID!"))
		return
	}

	lobby.Player2 = id
	lobby.Game = NewGame()

	w.WriteHeader(http.StatusOK)
	lobby.broadcastLobby()
}

func makeMoveHandler(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value("id").(string)

	var lobby *Lobby
	for _, activeLobby := range lobbies {
		if activeLobby.Player1 == id || activeLobby.Player2 == id {
			lobby = activeLobby
		}
	}

	if lobby == nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Current player is not in a lobby!"))
		return
	}

	var playerMakingRequest player.Player
	if lobby.Player1 == id {
		playerMakingRequest = player.Player1
	} else {
		playerMakingRequest = player.Player2
	}

	game := lobby.Game

	var from *int = nil
	rawFrom := r.URL.Query().Get("from")
	if len(rawFrom) > 0 {
		if initialPosition, err := strconv.Atoi(rawFrom); err == nil {
			from = &initialPosition
		} else {
			http.Error(w, "Invalid 'from' parameter, must be an integer", http.StatusBadRequest)
			return
		}
	}

	var to int
	rawTo := r.URL.Query().Get("to")
	if len(rawTo) > 0 {
		if targetPosition, err := strconv.Atoi(rawTo); err == nil {
			to = targetPosition
		} else {
			http.Error(w, "Invalid 'to' parameter, must be an integer", http.StatusBadRequest)
			return
		}
	} else {
		http.Error(w, "Missing required 'to' parameter, must be an integer", http.StatusBadRequest)
		return
	}

	newGame, err := game.EvaluateMove(playerMakingRequest, PlayerMove{from: from, to: to})
	if err != nil {
		http.Error(w, "Invalid move. Cause: "+err.Error(), http.StatusBadRequest)
	}

	lobby.Game = &newGame

	w.WriteHeader(http.StatusOK)
	lobby.broadcastLobby()
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	var id string
	idCookie, err := r.Cookie("id")
	if err != nil {
		id = uuid.NewString()
		http.SetCookie(w, &http.Cookie{
			Name:     "id",
			Value:    id,
			HttpOnly: true,
		})
	} else {
		id = idCookie.Value
	}

	conn, err := upgrader.Upgrade(w, r, w.Header())
	if err != nil {
		fmt.Println("Failed to upgrade connection:", err)
		return
	}
	//defer conn.Close()

	players[id] = conn
	fmt.Printf("%s connected\n", id)

	//for {
	//	messageType, message, err := conn.ReadMessage()
	//	if err != nil {
	//		fmt.Println("Read error:", err)
	//		break
	//	}
	//
	//	fmt.Printf("Received: %s\n", message)
	//
	//	if err := conn.WriteMessage(messageType, message); err != nil {
	//		fmt.Println("Write error:", err)
	//		break
	//	}
	//}
}
