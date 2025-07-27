package main

import (
	"backend/player"
	"backend/position"
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strconv"
)

type GameState string

const (
	Setup    GameState = "SETUP"
	Playing            = "PLAYING"
	GameOver           = "GAME_OVER"
)

type Game struct {
	State GameState
	Turn  player.Player
	Board []position.Position
}

type Lobby struct {
	Player1 string
	Player2 string
	Game    Game
}

var lobbies = map[string]*Lobby{}
var players = map[string]*websocket.Conn{}

func broadcastLobby(lobby *Lobby) {
	player1, player2 := players[lobby.Player1], players[lobby.Player2]
	game := lobby.Game

	player1.WriteJSON(game)
	player2.WriteJSON(game)
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Middleware func(http.Handler) http.Handler

func CreateStack(xs ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(xs) - 1; i >= 0; i-- {
			x := xs[i]
			next = x(next)
		}

		return next
	}
}

func MyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Hello from My Middleware!")
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), "foo", "bar")))
	})
}

func MyMiddleware2(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("The context is", r.Context().Value("foo"))
		next.ServeHTTP(w, r)
	})
}

func main() {
	http.HandleFunc("POST /api/create-lobby", func(w http.ResponseWriter, r *http.Request) {
		idCookie, err := r.Cookie("id")

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("ID cookie is not present. Connect to the WebSocket server first!"))
			return
		}

		lobbyId := uuid.NewString()
		lobbies[lobbyId] = &Lobby{
			Player1: idCookie.Value,
		}

		w.Write([]byte(lobbyId))
	})

	http.HandleFunc("POST /api/join-lobby", func(w http.ResponseWriter, r *http.Request) {
		idCookie, err := r.Cookie("id")

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("ID cookie is not present. Connect to the WebSocket server first!"))
			return
		}

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

		lobby.Player2 = idCookie.Value
		lobby.Game = Game{
			State: Setup,
			Turn:  player.Player1,
			Board: []position.Position{
				position.Empty,
				position.Empty,
				position.Empty,
				position.Empty,
				position.Empty,
				position.Empty,
				position.Empty,
				position.Empty,
				position.Empty,
			},
		}

		w.WriteHeader(http.StatusOK)
		broadcastLobby(lobby)
	})

	http.HandleFunc("/api/make-move", func(w http.ResponseWriter, r *http.Request) {
		idCookie, err := r.Cookie("id")

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("ID cookie is not present. Connect to the WebSocket server first!"))
			return
		}

		var lobby *Lobby
		id := idCookie.Value
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

		var currentPlayer player.Player
		if lobby.Player1 == id {
			currentPlayer = player.Player1
		} else {
			currentPlayer = player.Player2
		}

		game := &lobby.Game
		if (game.Turn == player.Player1 && lobby.Player1 != id) || (game.Turn == player.Player2 && lobby.Player2 != id) {
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte("It is not your Turn!"))
			fmt.Println("Invalid turn", game.Turn, "id", id, "players:", lobby.Player1, lobby.Player2)
			return
		}

		to := r.URL.Query().Get("to")

		targetPosition, _ := strconv.Atoi(to)

		board := game.Board
		if board[targetPosition] != position.Empty {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Target position must be empty!"))
			return
		}

		if game.State == Setup {
			if currentPlayer == player.Player1 {
				board[targetPosition] = position.Player1
				game.Turn = player.Player2
			} else {
				board[targetPosition] = position.Player2
				game.Turn = player.Player1
			}

			emptyCount := 0
			for _, pos := range board {
				if pos == position.Empty {
					emptyCount++
				}
			}

			if emptyCount == 3 {
				game.State = Playing
			}
		} else if game.State == Playing {
			from := r.URL.Query().Get("from")
			initialPosition, _ := strconv.Atoi(from)

			if currentPlayer == player.Player1 {
				if board[initialPosition] != position.Player1 {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte("Initial position must belong to one of your pieces!"))
					return
				}

				board[initialPosition] = position.Empty
				board[targetPosition] = position.Player1
				game.Turn = player.Player2
			} else if currentPlayer == player.Player2 {
				if board[initialPosition] != position.Player2 {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte("Initial position must belong to one of your pieces!"))
					return
				}

				board[initialPosition] = position.Empty
				board[targetPosition] = position.Player2
				game.Turn = player.Player1
			}
		}

		// TODO: Implement defensive checks
		//if len(to) == 0 {
		//	w.WriteHeader(http.StatusBadRequest)
		//	w.Write([]byte("Expected a \"to\" query parameter!"))
		//	return
		//}
		//
		//targetPosition, err := strconv.Atoi(to)
		//
		//if err != nil {
		//	w.WriteHeader(http.StatusBadRequest)
		//	w.Write([]byte("\"to\" must be an integer!"))
		//	return
		//}
		//
		//if targetPosition < 0 || targetPosition > 8 {
		//	w.WriteHeader(http.StatusBadRequest)
		//	w.Write([]byte("\"to\" must be an integer between 0 and 8!"))
		//	return
		//}
		//
		//if lobby.Game.State == Setup {
		//
		//}

		w.WriteHeader(http.StatusOK)
		broadcastLobby(lobby)
	})

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
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
	})

	fmt.Println("WebSocket server listening on port 8080!")
	//log.Fatal(http.ListenAndServe(
	//	":8080",
	//	CreateStack(
	//		MyMiddleware,
	//		MyMiddleware2,
	//	)(http.DefaultServeMux),
	//))
	log.Fatal(http.ListenAndServe(
		":8080",
		nil,
	))
}
