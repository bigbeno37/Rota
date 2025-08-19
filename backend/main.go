package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"log"
	"net/http"
	"os"
	"strings"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	redisUrl := "localhost:6379"
	if redisUrlEnvVar, exists := os.LookupEnv("REDIS_URL"); exists == true {
		redisUrl = redisUrlEnvVar
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: redisUrl,
	})

	redisError := rdb.Ping(ctx).Err()
	if redisError != nil {
		log.Fatal("Failed to connect to redis: " + redisError.Error())
	}

	fmt.Println("Connected to redis!")

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
		CreateStack(
			WithIdMiddleware,
			AddIdToLoggerMiddleware,
		)(authenticatedMux),
	)

	app := http.NewServeMux()
	app.Handle(
		"/",
		CreateStack(
			WithLoggerMiddleware,
			WithRedisMiddleware(rdb),
		)(mainMux),
	)

	port := "8080"
	if portNum, exists := os.LookupEnv("PORT"); exists == true {
		port = portNum
	}

	production := false
	if productionValue, exists := os.LookupEnv("production"); exists == true {
		production = strings.ToLower(productionValue) == "true"
	}

	if production {
		fmt.Println("Production server listening on 0.0.0.0:" + port)
		log.Fatal(http.ListenAndServe(
			"0.0.0.0:"+port,
			app,
		))
	} else {
		fmt.Println("Development server listening on localhost:8080")
		log.Fatal(http.ListenAndServe(
			"localhost:8080",
			app,
		))
	}
}
