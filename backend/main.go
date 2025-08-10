package main

import (
	"context"
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
	http.HandleFunc("POST /api/create-lobby", createLobbyHandler)
	http.HandleFunc("POST /api/join-lobby", joinLobbyHandler)
	http.HandleFunc("/api/make-move", makeMoveHandler)
	http.HandleFunc("/ws", wsHandler)

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
