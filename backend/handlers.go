package main

import (
	"backend/turn"
	"github.com/google/uuid"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"log/slog"
	"net/http"
	"strconv"
)

func createLobbyHandler(w http.ResponseWriter, r *http.Request) {
	logger := r.Context().Value("logger").(*slog.Logger)
	id := r.Context().Value("id").(string)
	state := NewGlobalStateManager(logger)

	lobbyId, _ := gonanoid.Generate("abcdefghijklmnopqrstuvwxyz0123456789", 8)
	state.CreateLobby(lobbyId, &Lobby{
		LobbyId: lobbyId,
		Player1: id,
	})

	player := state.GetPlayerWithId(id)
	player.SetLobby(&lobbyId)

	w.Write([]byte(lobbyId))
}

func joinLobbyHandler(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value("id").(string)
	logger := r.Context().Value("logger").(*slog.Logger)
	state := NewGlobalStateManager(logger)

	if !r.URL.Query().Has("lobbyId") {
		logger.Debug("Missing 'lobbyId' query parameter")
		http.Error(w, "No lobbyId present in request", http.StatusBadRequest)
		return
	}

	lobbyId := r.URL.Query().Get("lobbyId")
	lobbyWithId := state.GetLobbyWithId(lobbyId)

	if lobbyWithId == nil {
		logger.Debug("No lobby with ID " + lobbyId + " found")
		http.Error(w, "Invalid lobby ID!", http.StatusBadRequest)
		return
	}

	logger = logger.With("lobbyId", lobbyId)

	player := state.GetPlayerWithId(id)
	player.SetLobby(&lobbyId)

	lobby := NewLobbyManager(lobbyWithId, logger)

	lobby.SetPlayer2(&id)
	lobby.SetGame(NewGame())

	w.WriteHeader(http.StatusOK)

	logger.Info("Broadcasting lobby update to players")
	lobby.Broadcast(&LobbyEventMessage{
		Event: GameUpdate,
		Game:  lobby.Game(),
	})
}

func leaveLobbyHandler(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value("id").(string)
	logger := r.Context().Value("logger").(*slog.Logger)
	state := NewGlobalStateManager(logger)

	player := state.GetPlayerWithId(id)

	if player.currentLobby == nil {
		http.Error(w, "Player is not in any lobby!", http.StatusBadRequest)
		return
	}

	lobbyWithId := state.GetLobbyWithId(*player.currentLobby)
	if lobbyWithId == nil {
		logger.Warn("Player " + id + " is in lobby " + *player.currentLobby + " that no longer exists. Removing lobby ID from player...")

		player.SetLobby(nil)

		http.Error(w, "Player lobby no longer exists", http.StatusBadRequest)
		return
	}

	logger = logger.With("lobbyId", lobbyWithId.LobbyId)

	lobby := NewLobbyManager(lobbyWithId, logger)

	hasOpponent := false
	if lobby.Player1() == id {
		if lobby.Player2() == nil {
			logger.Debug("Player was the only user in lobby, deleting lobby...")
			state.RemoveLobby(lobby.lobby.LobbyId)
			player.SetLobby(nil)
		} else {
			logger.Debug("Second player exists, making them the lobby owner and leaving the lobby...")
			lobby.SetPlayer1(*lobby.Player2())
			lobby.SetPlayer2(nil)

			hasOpponent = true
		}
	} else {
		logger.Debug("Identified as player 2 in lobby, leaving...")
		lobby.SetPlayer2(nil)
		player.SetLobby(nil)

		hasOpponent = true
	}

	logger.Debug("User has left lobby")

	if hasOpponent {
		opponent := state.GetPlayerWithId(lobby.Player1())
		logger.Info("Sending update to user " + opponent.id)
		opponent.SendMessage(&LobbyEventMessage{
			Event: OpponentLeft,
			Game:  lobby.Game(),
		}, logger)
	}

	w.WriteHeader(http.StatusOK)
}

func makeMoveHandler(w http.ResponseWriter, r *http.Request) {
	id := r.Context().Value("id").(string)
	logger := r.Context().Value("logger").(*slog.Logger)
	state := NewGlobalStateManager(logger)

	player := state.GetPlayerWithId(id)
	if player == nil {
		logger.Warn("ID cookie present but no player found for ID " + id)
		http.Error(w, "Player not found. Please refresh your connection.", http.StatusInternalServerError)
		return
	}

	if player.currentLobby == nil {
		logger.Debug("Player " + id + " is not in any lobby")
		http.Error(w, "Player is not in any lobby!", http.StatusBadRequest)
		return
	}

	lobbyWithId := state.GetLobbyWithId(*player.currentLobby)
	if lobbyWithId == nil {
		logger.Warn("Player " + id + " is in lobby " + *player.currentLobby + " that no longer exists. Removing lobby ID from player...")

		player.SetLobby(nil)

		http.Error(w, "Player lobby no longer exists", http.StatusBadRequest)
		return
	}

	logger = logger.With("lobbyId", lobbyWithId.LobbyId)

	lobby := NewLobbyManager(lobbyWithId, logger)

	evalGame := func() *Game {
		var playerMakingRequest turn.Turn
		if lobby.Player1() == id {
			playerMakingRequest = turn.Player1
		} else {
			playerMakingRequest = turn.Player2
		}

		game := lobby.Game()

		var from *int = nil
		rawFrom := r.URL.Query().Get("from")
		if len(rawFrom) > 0 {
			if initialPosition, err := strconv.Atoi(rawFrom); err == nil {
				from = &initialPosition
			} else {
				http.Error(w, "Invalid 'from' parameter, must be an integer", http.StatusBadRequest)
				return nil
			}
		}

		var to int
		rawTo := r.URL.Query().Get("to")
		if len(rawTo) > 0 {
			if targetPosition, err := strconv.Atoi(rawTo); err == nil {
				to = targetPosition
			} else {
				http.Error(w, "Invalid 'to' parameter, must be an integer", http.StatusBadRequest)
				return nil
			}
		} else {
			http.Error(w, "Missing required 'to' parameter, must be an integer", http.StatusBadRequest)
			return nil
		}

		newGame, err := game.EvaluateMove(playerMakingRequest, PlayerMove{from: from, to: to})
		if err != nil {
			http.Error(w, "Invalid move. Cause: "+err.Error(), http.StatusBadRequest)
			return nil
		}

		return &newGame
	}

	newGame := evalGame()

	if newGame == nil {
		return
	}

	lobby.SetGame(newGame)

	w.WriteHeader(http.StatusOK)
	logger.Info("Broadcasting updated game")
	lobby.Broadcast(&LobbyEventMessage{
		Event: GameUpdate,
		Game:  lobby.Game(),
	})
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	logger := r.Context().Value("logger").(*slog.Logger)
	state := NewGlobalStateManager(logger)

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
		logger.Info("Failed to upgrade connection: " + err.Error())
		return
	}
	//defer conn.Close()

	logger.Info("New player connected!")
	state.CreatePlayer(id, NewPlayer(id, conn))

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
