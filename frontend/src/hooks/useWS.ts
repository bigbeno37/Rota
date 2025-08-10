import {useEffect, useState} from 'react';
import type {Game} from '@/types.ts';

type LobbyEventMessage = {
	Event: 'GAME_UPDATE' | 'OPPONENT_LEFT';
	Game: Game
}

export const useWS = (onMessage: (message: LobbyEventMessage) => void) => {
	const [wsStatus, setWsStatus] = useState<{ state: 'LOADING' | 'CONNECTED' | 'CLOSED' } | {
		state: 'ERROR',
		error: any
	}>({state: 'LOADING'});

	useEffect(() => {
		const connection = new WebSocket('ws://localhost:8080/ws');

		connection.addEventListener('open', () => setWsStatus({state: 'CONNECTED'}));
		connection.addEventListener('error', e => setWsStatus({state: 'ERROR', error: e}));
		connection.addEventListener('message', e => {
			console.log('Received WS message', e.data);

			// setGame(JSON.parse(e.data));
			onMessage(JSON.parse(e.data));
		});
		connection.addEventListener('close', () => {
			setWsStatus({ state: 'CLOSED' });
		})
	}, []);

	return wsStatus;
}