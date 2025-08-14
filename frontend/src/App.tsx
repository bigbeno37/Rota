import {useEffect, useState} from 'react';
import {Button} from '@/components/ui/button.tsx';
import {Label} from '@/components/ui/label.tsx';
import {Input} from '@/components/ui/input.tsx';
import {useMutation} from '@tanstack/react-query';
import {throwIfNotOk} from '@/utils.ts';
import {Board} from '@/Board.tsx';
import type {Game} from '@/types.ts';
import {useWS} from '@/hooks/useWS.ts';

export function App() {
	const [game, setGame] = useState<Game | null>(null);
	const wsStatus = useWS(message => {
		if (message.Event === 'GAME_UPDATE') {
			setGame(message.Game);
		} else if (message.Event === 'OPPONENT_LEFT') {
			setGame(null);
		}
	})

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

	const leaveLobbyMutation = useMutation({
		mutationFn: () => {
			return throwIfNotOk(fetch('/api/leave-lobby', {
				method: 'POST'
			}))
		},
		onSuccess: () => {
			setGame(null);
			setPlayerState('MAIN_MENU');
		}
	});

	const handleLeaveLobbyClicked = () => {
		leaveLobbyMutation.mutate();
	}

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
			<div className="flex flex-col items-center justify-center w-full h-full">
				<div className="flex flex-col gap-8 w-[900px]">
					<span className="text-center text-4xl">Rota</span>
					{createLobbyMutation.isError && <p>{'' + createLobbyMutation.error}</p>}
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
						<Button disabled={disableControls || !lobbyId} onClick={handleJoinLobbyClicked}>Join
							Lobby</Button>
					</div>
				</div>
			</div>
		);
	} else if (playerState === 'IN_LOBBY') {
		return (
			<>
				<Button
					disabled={leaveLobbyMutation.isPending}
					onClick={handleLeaveLobbyClicked}
				>
					Leave game
				</Button>
				{!game && (<p>Waiting for opponent to join. Lobby ID is: {lobbyId}</p>)}
				{game && (<>
					<p>This is the {game.State} phase.</p>
					{game.State === 'PLAYING' ? (<p>It is {game.Turn}'s turn.</p>) : game.State === 'GAME_OVER' ? (
						<p>{game.Turn} won!</p>) : null}

					{makeMoveMutation.isError &&
                        <p className="text-red-700">Invalid move! {'' + makeMoveMutation.error}</p>}
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