# Rota
“Rota” is a common modern name for an easy strategy game played
on a round board. The Latin name is probably Terni Lapilli (“Three
pebbles”). Many boards survive, both round and rectangular, and
there must have been variations in play.

This is a very simple game, and Rota is especially appropriate to play
with young family members and friends! Unlike Tic Tac Toe, Rota
avoids a tie.

# Board
A center circle is surrounded by 8 equidistant circles, each connecting to their neighbours and the central circle. 

## RULES
**Goal:** The first player to place three game pieces in a row across the center or in the circle of the board wins
- Roll a die or flip a coin to determine who starts. The higher number
plays first
- Players take turns placing one piece on the board in any open spot 
- After all the pieces are on the board, a player moves one piece each
turn onto the next empty spot (along spokes or circle)

A player may not:
- Skip a turn, even if the move forces you to lose the game
- Jump over another piece
- Move more than one space
- Land on a space with a piece already on it
- Knock a piece off a space

# Development
I have been developing this project with Node 24.4.1 and Go 1.23.11, so your mileage may vary with earlier versions.

To run this app, navigate into the `backend` directory, and run `go mod tidy` to download all Go-related dependencies.
Once complete, run `go run main.go` to get the WebSocket server started.

To start the frontend, navigate to the `frontend` directory and run `npm i` to install all related dependencies.
Once complete, run `npm run dev` to get the server started.

Unless configured otherwise, the app shall now be running at `http://localhost:5173`, with the Go HTTP/WebSocket server started on `http://localhost:8080` and `ws://localhost:8080` respectively.