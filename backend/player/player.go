package player

import "backend/position"

type Player string

const (
	Player1 Player = "PLAYER_1"
	Player2        = "PLAYER_2"
)

func (p Player) AsPosition() position.Position {
	if p == Player1 {
		return position.Player1
	} else {
		return position.Player2
	}
}
