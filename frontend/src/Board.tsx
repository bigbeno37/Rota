import type {Game} from '@/types.ts';

export const Board = (props: {
	game: Game,
	disabled: boolean,
	activePosition: number,
	onPositionClicked: (position: number) => void
}) => {
	const circleStyle = 'rounded-full border border-black h-24 w-24 cursor-pointer hover:bg-gray-300';

	const Circle = ({position}: { position: number }) => {
		const player1Styling = 'bg-blue-300';
		const player2Styling = 'bg-purple-300';
		const activeStyling = 'border-4 border-red-700';

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
			<Circle position={1}/>
			<div className="col-span-4"></div>

			<div className="col-span-9"></div>

			<div className="col-span-2"></div>
			<Circle position={8}/>
			<div className="col-span-3"></div>
			<Circle position={2}/>
			<div className="col-span-2"></div>

			<div className="col-span-9"></div>

			<Circle position={7}/>
			<div className="col-span-3"></div>
			<Circle position={0}/>
			<div className="col-span-3"></div>
			<Circle position={3}/>

			<div className="col-span-9"></div>

			<div className="col-span-2"></div>
			<Circle position={6}/>
			<div className="col-span-3"></div>
			<Circle position={4}/>
			<div className="col-span-2"></div>

			<div className="col-span-9"></div>

			<div className="col-span-4"></div>
			<Circle position={5}/>
			<div className="col-span-4"></div>
		</div>
	);
};