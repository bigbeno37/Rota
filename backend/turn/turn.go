package turn

import "backend/position"

type Turn string

const (
	Player1 Turn = "PLAYER_1"
	Player2      = "PLAYER_2"
)

func (t Turn) AsPosition() position.Position {
	if t == Player1 {
		return position.Player1
	} else {
		return position.Player2
	}
}
