package main

import (
	"backend/position"
	"backend/turn"
	"math"
)

type GameState string

const (
	Setup    GameState = "SETUP"
	Playing            = "PLAYING"
	GameOver           = "GAME_OVER"
)

type Game struct {
	State GameState
	Turn  turn.Turn
	Board []position.Position
}

func NewGame() *Game {
	return &Game{
		State: Setup,
		Turn:  turn.Player1,
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
}

type PlayerMove struct {
	from *int
	to   int
}

type InvalidMove string

const (
	WrongPlayer                 InvalidMove = "WRONG_PLAYER"
	TargetOutOfBounds                       = "TARGET_OUT_OF_BOUNDS"
	TargetIsNotEmpty                        = "TARGET_IS_NOT_EMPTY"
	SourceMissing                           = "SOURCE_MISSING"
	SourceOutOfBounds                       = "SOURCE_OUT_OF_BOUNDS"
	SourceDoesNotBelongToPlayer             = "SOURCE_DOES_NOT_BELONG_PLAYER"
	InvalidTarget                           = "INVALID_TARGET"
	GameIsOver                              = "GAME_IS_OVER"
)

type InvalidMoveError struct {
	cause InvalidMove
}

func (e *InvalidMoveError) Error() string {
	return string(e.cause)
}

func (game *Game) PlayerHasWon(p turn.Turn) bool {
	pos := p.AsPosition()

	hasDiagonalVictory := false
	for i := 8; i > 4; i-- {
		if game.Board[i] == pos && game.Board[0] == pos && game.Board[i-4] == pos {
			hasDiagonalVictory = true
		}
	}

	hasRowVictory := false
	for i := 1; i <= 6; i++ {
		if game.Board[i] == pos && game.Board[i+1] == pos && game.Board[i+2] == pos {
			hasRowVictory = true
		}
	}

	if game.Board[7] == pos && game.Board[8] == pos && game.Board[1] == pos {
		hasRowVictory = true
	}

	if game.Board[8] == pos && game.Board[1] == pos && game.Board[2] == pos {
		hasRowVictory = true
	}

	return hasDiagonalVictory || hasRowVictory
}

func (currentGame *Game) EvaluateMove(p turn.Turn, move PlayerMove) (Game, error) {
	existingBoard := make([]position.Position, len(currentGame.Board))
	copy(existingBoard, currentGame.Board)
	game := Game{
		State: currentGame.State,
		Turn:  currentGame.Turn,
		Board: existingBoard,
	}

	if game.State == GameOver {
		return game, &InvalidMoveError{cause: GameIsOver}
	}

	if game.Turn != p {
		return game, &InvalidMoveError{cause: WrongPlayer}
	}

	if move.to < 0 || move.to >= len(game.Board) {
		return game, &InvalidMoveError{cause: TargetOutOfBounds}
	}

	if game.Board[move.to] != position.Empty {
		return game, &InvalidMoveError{cause: TargetIsNotEmpty}
	}

	var nextPlayer turn.Turn
	if p == turn.Player1 {
		nextPlayer = turn.Player2
	} else {
		nextPlayer = turn.Player1
	}

	pos := p.AsPosition()
	if game.State == Setup {
		game.Board[move.to] = pos

		// Rare case where players set up into a winning position
		if game.PlayerHasWon(p) || game.PlayerHasWon(nextPlayer) {
			game.State = GameOver

			if game.PlayerHasWon(nextPlayer) {
				game.Turn = nextPlayer
			}

			return game, nil
		}

		game.Turn = nextPlayer

		emptyCount := 0
		for _, pos := range existingBoard {
			if pos == position.Empty {
				emptyCount++
			}
		}

		if emptyCount == 3 {
			game.State = Playing
		}
	} else if game.State == Playing {
		if move.from == nil {
			return game, &InvalidMoveError{cause: SourceMissing}
		}

		from := *move.from
		if from < 0 || from >= len(game.Board) {
			return game, &InvalidMoveError{cause: SourceOutOfBounds}
		}

		if game.Board[*move.from] != pos {
			return game, &InvalidMoveError{cause: SourceDoesNotBelongToPlayer}
		}

		isMovingToOrFromZero := move.to == 0 || from == 0
		isMovingBetweenOneAndEight := (from == 1 && move.to == 8) || (from == 8 && move.to == 1)
		isMovingOneSpace := math.Abs(float64(from-move.to)) == 1
		isValidMove := isMovingToOrFromZero || isMovingBetweenOneAndEight || isMovingOneSpace

		if !isValidMove {
			return game, &InvalidMoveError{cause: InvalidTarget}
		}

		game.Board[from] = position.Empty
		game.Board[move.to] = pos

		if game.PlayerHasWon(p) {
			game.State = GameOver
		} else {
			game.Turn = nextPlayer
		}
	}

	return game, nil
}
