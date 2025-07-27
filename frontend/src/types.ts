export type Game = {
	State: 'SETUP' | 'PLAYING' | 'GAME_OVER',
	Turn: 'PLAYER_1' | 'PLAYER_2',
	Board: Array<'PLAYER_1' | 'PLAYER_2' | 'EMPTY'>
}