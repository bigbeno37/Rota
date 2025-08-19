package main

import (
	"backend/turn"
	"context"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"net/http"
	"strconv"
)

func WatchWithRetries(ctx context.Context, executeWatch func() error, retryLimit int) error {
	for i := 0; i < retryLimit; i++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		err := executeWatch()
		if err == nil {
			return nil
		}

		if errors.Is(err, redis.TxFailedErr) {
			continue
		}

		return err
	}

	return redis.TxFailedErr
}

func StrAsJson(str string) string {
	return "\"" + str + "\""
}

func createLobbyHandler(w http.ResponseWriter, r *http.Request) {
	logger := GetLoggerFromContext(r.Context())
	id := GetIdFromContext(r.Context())
	rdb := GetRedisFromContext(r.Context())

	lobbyId, _ := gonanoid.Generate("abcdefghijklmnopqrstuvwxyz0123456789", 8)

	logger = logger.With(slog.String("lobbyId", lobbyId))

	tx := func(tx *redis.Tx) error {
		_, err := tx.TxPipelined(r.Context(), func(pipe redis.Pipeliner) error {
			logger.Info("Creating lobby...")
			err := pipe.JSONSet(r.Context(), "lobby:"+lobbyId, "$", Lobby{
				LobbyId: lobbyId,
				Player1: id,
			}).Err()

			if err != nil {
				return err
			}

			logger.Info("Adding player to lobby...")
			err = pipe.JSONSet(r.Context(), "player:"+id, "$.CurrentLobby", StrAsJson(lobbyId)).Err()

			if err != nil {
				return err
			}

			return nil
		})

		return err
	}

	err := WatchWithRetries(
		ctx,
		func() error {
			return rdb.Watch(r.Context(), tx, "player:"+id)
		},
		5,
	)

	if err != nil {
		logger.Warn("There was an error creating the lobby: " + err.Error())
		http.Error(w, "There was an error creating the lobby", http.StatusInternalServerError)
		return
	}

	w.Write([]byte(lobbyId))
}

func joinLobbyHandler(w http.ResponseWriter, r *http.Request) {
	id := GetIdFromContext(r.Context())
	logger := GetLoggerFromContext(r.Context())
	rdb := GetRedisFromContext(r.Context())

	if !r.URL.Query().Has("lobbyId") {
		logger.Debug("Missing 'lobbyId' query parameter")
		http.Error(w, "No lobbyId present in request", http.StatusBadRequest)
		return
	}

	lobbyId := r.URL.Query().Get("lobbyId")
	lobbyJson, err := rdb.JSONGet(r.Context(), "lobby:"+lobbyId).Result()

	if err != nil {
		logger.Warn("There was an error fetching lobbies: " + err.Error())
		http.Error(w, "There was an error fetching lobbies", http.StatusInternalServerError)
		return
	}

	if len(lobbyJson) == 0 {
		logger.Debug("No lobby with ID " + lobbyId + " found")
		http.Error(w, "Invalid lobby ID!", http.StatusBadRequest)
		return
	}

	var lobby Lobby
	json.Unmarshal([]byte(lobbyJson), &lobby)

	logger = logger.With("lobbyId", lobbyId)

	var updatedLobby Lobby

	tx := func(tx *redis.Tx) error {
		_, err := tx.TxPipelined(r.Context(), func(pipe redis.Pipeliner) error {
			err = pipe.JSONSet(r.Context(), "player:"+id, "$.CurrentLobby", StrAsJson(lobbyId)).Err()

			if err != nil {
				return err
			}

			err = pipe.JSONSet(r.Context(), "lobby:"+lobbyId, "$.Player2", StrAsJson(id)).Err()

			if err != nil {
				return err
			}

			err = pipe.JSONSet(r.Context(), "lobby:"+lobbyId, "$.Game", NewGame()).Err()

			if err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			return err
		}

		updatedLobbyJson, err := tx.JSONGet(r.Context(), "lobby:"+lobbyId).Result()
		if err != nil {
			return err
		}

		json.Unmarshal([]byte(updatedLobbyJson), &updatedLobby)

		return nil
	}

	err = WatchWithRetries(ctx, func() error {
		return rdb.Watch(r.Context(), tx, "player:"+id, "lobby:"+lobbyId)
	}, 5)

	if err != nil {
		logger.Warn("There was an error joining lobby: " + err.Error())
		http.Error(w, "There was an error joining lobby", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

	logger.Info("Broadcasting lobby update to players")
	message, _ := json.Marshal(LobbyEventMessage{
		Event: GameUpdate,
		Game:  updatedLobby.Game,
	})

	err = rdb.Publish(r.Context(), "player:"+updatedLobby.Player1, message).Err()

	if err != nil {
		logger.Warn("There was an error publishing the lobby update: " + err.Error())
	}

	err = rdb.Publish(r.Context(), "player:"+*updatedLobby.Player2, message).Err()

	if err != nil {
		logger.Warn("There was an error publishing the lobby update: " + err.Error())
	}
}

func leaveLobbyHandler(w http.ResponseWriter, r *http.Request) {
	id := GetIdFromContext(r.Context())
	rdb := GetRedisFromContext(r.Context())
	logger := GetLoggerFromContext(r.Context())

	hasOpponent := false
	var sendUpdateToPlayerId *string
	tx := func(tx *redis.Tx) error {
		playerJson, err := tx.JSONGet(r.Context(), "player:"+id).Result()

		if err != nil {
			return err
		}

		var player Player
		json.Unmarshal([]byte(playerJson), &player)

		if player.CurrentLobby == nil {
			return nil
		}

		lobbyJson, err := tx.JSONGet(r.Context(), "lobby:"+*player.CurrentLobby).Result()
		if err != nil {
			return err
		}

		if len(lobbyJson) == 0 {
			logger.Warn("Player is in lobby " + *player.CurrentLobby + " that no longer exists. Removing lobby ID from player...")
			tx.JSONSet(r.Context(), "player:"+id, "$.CurrentLobby", nil)

			return nil
		}

		var lobby Lobby
		json.Unmarshal([]byte(lobbyJson), &lobby)

		_, err = tx.TxPipelined(r.Context(), func(pipe redis.Pipeliner) error {
			hasOpponent = false
			if lobby.Player1 == id {
				if lobby.Player2 == nil {
					logger.Debug("Player was the only user in lobby, deleting lobby...")
					err = pipe.JSONDel(r.Context(), "lobby:"+lobby.LobbyId, "$").Err()
					if err != nil {
						return err
					}

					err = pipe.JSONSet(r.Context(), "player:"+id, "$.CurrentLobby", nil).Err()
					if err != nil {
						return err
					}
				} else {
					logger.Debug("Second player exists, making them the lobby owner and leaving the lobby...")
					err = pipe.JSONSet(r.Context(), "lobby:"+lobby.LobbyId, "$.Player1", StrAsJson(*lobby.Player2)).Err()
					if err != nil {
						return err
					}

					err = pipe.JSONSet(r.Context(), "lobby:"+lobby.LobbyId, "$.Player2", nil).Err()
					if err != nil {
						return err
					}

					hasOpponent = true
					sendUpdateToPlayerId = lobby.Player2
				}
			} else {
				logger.Debug("Identified as player 2 in lobby, leaving...")
				err = pipe.JSONSet(r.Context(), "lobby:"+lobby.LobbyId, "$.Player2", nil).Err()
				if err != nil {
					return err
				}

				err = pipe.JSONSet(r.Context(), "player:"+id, "$.CurrentLobby", nil).Err()
				if err != nil {
					return err
				}

				hasOpponent = true
				sendUpdateToPlayerId = &lobby.Player1
			}

			return nil
		})

		return err
	}

	err := WatchWithRetries(r.Context(), func() error {
		return rdb.Watch(r.Context(), tx, "player:"+id)
	}, 5)

	if err != nil {
		logger.Warn("There was an error leaving player lobby: " + err.Error())
		http.Error(w, "There was an error leaving player lobby", http.StatusInternalServerError)
		return
	}

	logger.Debug("User has left lobby")

	if hasOpponent && sendUpdateToPlayerId != nil {
		logger.Info("Sending update to user " + *sendUpdateToPlayerId)
		message, _ := json.Marshal(LobbyEventMessage{
			Event: OpponentLeft,
			Game:  nil,
		})

		err = rdb.Publish(r.Context(), "player:"+*sendUpdateToPlayerId, message).Err()

		if err != nil {
			logger.Warn("There was an error publishing \"opponent left\" message: " + err.Error())
		}
	}

	w.WriteHeader(http.StatusOK)
}

type MakeMoveValidationError struct {
	cause string
}

func (e MakeMoveValidationError) Error() string {
	return e.cause
}

func makeMoveHandler(w http.ResponseWriter, r *http.Request) {
	id := GetIdFromContext(r.Context())
	logger := GetLoggerFromContext(r.Context())
	rdb := GetRedisFromContext(r.Context())

	existingPlayerJson, err := rdb.JSONGet(r.Context(), "player:"+id).Result()

	if err != nil {
		logger.Warn("Unable to fetch player from redis: " + err.Error())
		http.Error(w, "Unable to fetch player from redis", http.StatusInternalServerError)
		return
	}

	var existingPlayer Player
	json.Unmarshal([]byte(existingPlayerJson), &existingPlayer)

	if existingPlayer.CurrentLobby == nil {
		http.Error(w, "The player is not in a lobby!", http.StatusBadRequest)
		return
	}

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

	tx := func(tx *redis.Tx) error {
		playerJson, err := tx.JSONGet(r.Context(), "player:"+id).Result()

		if err != nil {
			return err
		}

		var player Player
		json.Unmarshal([]byte(playerJson), &player)

		if player.CurrentLobby == nil {
			return MakeMoveValidationError{cause: "NOT_IN_LOBBY"}
		}

		if *player.CurrentLobby != *existingPlayer.CurrentLobby {
			return MakeMoveValidationError{cause: "CONCURRENT_EDIT"}
		}

		lobbyJson, err := tx.JSONGet(r.Context(), "lobby:"+*player.CurrentLobby).Result()
		if err != nil {
			return err
		}

		if len(lobbyJson) == 0 {
			tx.JSONSet(r.Context(), "player:"+id, "$.CurrentLobby", nil)
			return MakeMoveValidationError{cause: "NOT_IN_LOBBY"}
		}

		var lobby Lobby
		json.Unmarshal([]byte(lobbyJson), &lobby)

		logger = logger.With("lobbyId", lobby.LobbyId)

		if lobby.Player2 == nil {
			return MakeMoveValidationError{cause: "WAITING_FOR_PLAYER_2"}
		}

		var playerMakingRequest turn.Turn
		if lobby.Player1 == id {
			playerMakingRequest = turn.Player1
		} else {
			playerMakingRequest = turn.Player2
		}

		game := lobby.Game

		newGame, err := game.EvaluateMove(playerMakingRequest, PlayerMove{from: from, to: to})
		if err != nil {
			return err
		}

		tx.JSONSet(r.Context(), "lobby:"+lobby.LobbyId, "$.Game", newGame)

		return nil
	}

	err = WatchWithRetries(r.Context(), func() error {
		return rdb.Watch(r.Context(), tx, "player:"+id, "lobby:"+*existingPlayer.CurrentLobby)
	}, 5)

	if err != nil {
		if errors.Is(err, MakeMoveValidationError{}) {
			http.Error(w, "Unable to make move: "+err.Error(), http.StatusBadRequest)
		} else {
			logger.Warn("Error making move caused by: " + err.Error())
			http.Error(w, "Error making move: "+err.Error(), http.StatusInternalServerError)
		}

		return
	}

	w.WriteHeader(http.StatusOK)
	logger.Info("Broadcasting updated game")

	lobbyJson, err := rdb.JSONGet(r.Context(), "lobby:"+*existingPlayer.CurrentLobby).Result()

	if err != nil {
		logger.Warn("Unable to fetch lobby from redis: " + err.Error())
		return
	}

	var lobby Lobby
	json.Unmarshal([]byte(lobbyJson), &lobby)

	message, _ := json.Marshal(LobbyEventMessage{
		Event: GameUpdate,
		Game:  lobby.Game,
	})

	err = rdb.Publish(r.Context(), "player:"+lobby.Player1, message).Err()
	if err != nil {
		logger.Warn("There was an error publishing lobby update: " + err.Error())
	}

	err = rdb.Publish(r.Context(), "player:"+*lobby.Player2, message).Err()
	if err != nil {
		logger.Warn("There was an error publishing lobby update: " + err.Error())
	}
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	logger := GetLoggerFromContext(r.Context())
	redis := GetRedisFromContext(r.Context())

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

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()

	logger.Info("New player connected!")
	err = redis.JSONSet(r.Context(), "player:"+id, "$", &Player{
		Id:           id,
		CurrentLobby: nil,
	}).Err()

	if err != nil {
		logger.Warn("Failed to set player: " + err.Error())
	}

	pubsub := redis.Subscribe(r.Context(), "player:"+id)
	defer pubsub.Close()
	ch := pubsub.Channel()

	for {
		select {
		case <-done:
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			if err := conn.WriteMessage(websocket.TextMessage, []byte(msg.Payload)); err != nil {
				logger.Warn("Failed to send player: " + err.Error())
				return
			}
		}
	}
}
