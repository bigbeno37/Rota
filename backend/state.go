package main

import "github.com/gorilla/websocket"

var lobbies = map[string]*Lobby{}
var players = map[string]*websocket.Conn{}
