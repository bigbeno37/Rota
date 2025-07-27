import {useEffect, useRef, useState} from 'react';
import {Button} from '@/components/ui/button.tsx';
import {Label} from '@/components/ui/label.tsx';
import {Input} from '@/components/ui/input.tsx';
import {useMutation} from '@tanstack/react-query';

function throwIfNotOk(fetchCall: Promise<Response>) {
	return fetchCall
		.then((response) => {
			return response.text()
				.then(text => {
					if (!response.ok) {
						throw new Error(text);
					}

					return text
				});
		})
}

type Game = {
	State: 'SETUP' | 'PLAYING' | 'GAME_OVER',
	Turn: 'PLAYER_1' | 'PLAYER_2',
	Board: Array<'PLAYER_1' | 'PLAYER_2' | 'EMPTY'>
}

const Board = (props: { game: Game, disabled: boolean, activePosition: number, onPositionClicked: (position: number) => void }) => {
	const circleStyle = 'rounded-full border border-black h-24 w-24 cursor-pointer hover:bg-gray-300';

	const Circle = ({ position }: { position: number }) => {
		const player1Styling = 'bg-blue-300';
		const player2Styling = 'bg-purple-300';
		const activeStyling = 'border-4 border-red-700'

		const handleClick = () => {
			props.onPositionClicked(position);
		};

		return (
			<button
				className={`${circleStyle} ${props.game.Board[position] === 'PLAYER_1' ? player1Styling : props.game.Board[position] === 'PLAYER_2' ? player2Styling : ''} ${props.activePosition === position && activeStyling}`}
				disabled={props.disabled}
				onClick={handleClick}
			></button>
		);
	};

	return (
		<div className="grid grid-cols-9 grid-rows-9 gap-4 w-4xl">
			<div className="col-span-4"></div>
			<Circle position={1} />
			<div className="col-span-4"></div>

			<div className="col-span-9"></div>

			<div className="col-span-2"></div>
			<Circle position={8} />
			<div className="col-span-3"></div>
			<Circle position={2} />
			<div className="col-span-2"></div>

			<div className="col-span-9"></div>

			<Circle position={7} />
			<div className="col-span-3"></div>
			<Circle position={0} />
			<div className="col-span-3"></div>
			<Circle position={3} />

			<div className="col-span-9"></div>

			<div className="col-span-2"></div>
			<Circle position={6} />
			<div className="col-span-3"></div>
			<Circle position={4} />
			<div className="col-span-2"></div>

			<div className="col-span-9"></div>

			<div className="col-span-4"></div>
			<Circle position={5} />
			<div className="col-span-4"></div>
		</div>
	)
};

export function App() {
	const ws = useRef<WebSocket>(null);

	const [wsStatus, setWsStatus] = useState<{ state: 'LOADING' | 'CONNECTED' | 'CLOSED' } | {
		state: 'ERROR',
		error: any
	}>({state: 'LOADING'});
	const [game, setGame] = useState<Game | null>(null);

	useEffect(() => {
		const connection = new WebSocket('ws://localhost:8080/ws');

		connection.addEventListener('open', () => setWsStatus({state: 'CONNECTED'}));
		connection.addEventListener('error', e => setWsStatus({state: 'ERROR', error: e}));
		connection.addEventListener('message', e => {
			console.log('Received WS message', e.data);

			setGame(JSON.parse(e.data));
		});
		connection.addEventListener('close', () => {
			setWsStatus({ state: 'CLOSED' });
		})
	}, []);

	const [playerState, setPlayerState] = useState<'MAIN_MENU' | 'IN_LOBBY'>('MAIN_MENU');
	const [lobbyId, setLobbyId] = useState<string | null>(null);

	const createLobbyMutation = useMutation({
		mutationFn: () => {
			return throwIfNotOk(fetch('/api/create-lobby', {
				method: 'POST',
			}));
		}
	});

	const joinLobbyMutation = useMutation({
		mutationFn: () => {
			return throwIfNotOk(fetch(`/api/join-lobby?lobbyId=${lobbyId}`, {
				method: 'POST',
			}));
		}
	});

	const disableControls = createLobbyMutation.isPending || joinLobbyMutation.isPending;

	const makeMoveMutation = useMutation({
		mutationFn: (opts: { from?: number, to: number }) => {
			return throwIfNotOk(fetch(`/api/make-move?from=${opts.from ?? -1}&to=${opts.to}`, {
				method: 'POST'
			}));
		},
		onError: () => {
			setActivePosition(-1);
		}
	});

	// Clear any error messages when the board has been updated
	useEffect(() => {
		makeMoveMutation.reset();
	}, [game]);

	const handleCreateLobbyClicked = async () => {
		const lobbyId = await createLobbyMutation.mutateAsync();
		setLobbyId(lobbyId);
		setPlayerState('IN_LOBBY');
	};

	const handleJoinLobbyClicked = async () => {
		await joinLobbyMutation.mutateAsync();
		setPlayerState('IN_LOBBY');
	};

	const [activePosition, setActivePosition] = useState<number>(-1)


	const handlePositionClicked = async (position: number) => {
		if (game?.State === 'PLAYING') {
			if (activePosition === -1) {
				setActivePosition(position);
			} else {
				makeMoveMutation.mutateAsync({ from: activePosition, to: position });
			}
		} else {
			makeMoveMutation.mutate({ to: position });
		}
	};

	// Remove any active positions when receiving board updates
	useEffect(() => {
		setActivePosition(-1);
	}, [game]);

	if (wsStatus.state === 'LOADING') {
		return <p>Loading...</p>;
	}

	if (wsStatus.state === 'CLOSED') {
		return <p>WebSocket connection closed. Please refresh the page.</p>;
	}

	if (wsStatus.state === 'ERROR') {
		return <p>There was an error connecting to the WebSocket server: {JSON.stringify(wsStatus.error)}</p>;
	}

	if (playerState === 'MAIN_MENU') {
		return (
			<div className="flex flex-col gap-8 w-[900px] m-4">
				{createLobbyMutation.isError && <p>{''+createLobbyMutation.error}</p>}
				<Button
					variant="outline"
					disabled={disableControls}
					onClick={handleCreateLobbyClicked}
				>
					Create Lobby
				</Button>
				<div className="flex flex-col gap-2">
					<Label htmlFor="lobby-id">Lobby ID: </Label>
					<Input
						id="lobby-id"
						type="text"
						disabled={disableControls}
						value={lobbyId ?? ''}
						onChange={e => setLobbyId(e.target.value)}
					/>
					<Button disabled={disableControls || !lobbyId} onClick={handleJoinLobbyClicked}>Join Lobby</Button>
				</div>
			</div>
		);
	} else if (playerState === 'IN_LOBBY') {
		return (
			<>
				{!game && (<p>Waiting for opponent to join. Lobby ID is: {lobbyId}</p>)}
				{game && (<>
					<p>{JSON.stringify(game)}</p>
					<p>This is the {game.State} phase.</p>
					<p>It is {game.Turn}'s turn.</p>
					{makeMoveMutation.isError && <p className="text-red-700">Invalid move! {''+makeMoveMutation.error}</p>}
					{makeMoveMutation.isPending && <p>Submitting move...</p>}
					<Board
						disabled={makeMoveMutation.isPending}
						game={game}
						activePosition={activePosition}
						onPositionClicked={handlePositionClicked}
					/>
				</>)}
			</>
		);
	}

	return <p>Invalid player state: {playerState}</p>;
}