<script lang="ts">
    import { derived } from 'svelte/store';
	import { specialCells } from '$lib/cells';
	import { pendingMove } from '$lib/stores/pendingMove';
  	import { letterValues } from '$lib/lettres_value';

	let { game, onPlaceLetter }: {
		game: GameInfo | null;
		onPlaceLetter: (x: number, y: number, cell: string) => void;
	} = $props();

	type DisplayCell = {
		x: number;
		y: number;
		char: string;
		points: number | null;
		className: string;
	};

	const lastMoveCoords = (() => {
		if (!game?.moves?.length) return [];
		const last = game.moves[game.moves.length - 1];
		return last
			? (last.move.letters).map(m => ({ x: m.x, y: m.y }))
			: [];
	})();

	const computedBoard = derived(pendingMove, ($pendingMove) => {
		if (!game) return [];

		return game.board.map((row: string[], y: number) =>
			row.map((cell: string, x: number): DisplayCell => {
				const key = `${y},${x}`;
				const special = specialCells.get(key);
				const pending = $pendingMove.find((p) => p.x === x && p.y === y);
				const displayed = cell || pending?.letter || special || '';
				const isPlacedLetter = cell !== "" && !pending;
				const inLastTurn = lastMoveCoords.some(p => p.x === x && p.y === y);

				let base = "relative aspect-square w-full text-center flex items-center justify-center border border-gray-300 cursor-pointer select-none overflow-hidden";
				let color = "";

				if (isPlacedLetter) {
					color = inLastTurn
						? "bg-orange-200 text-yellow-800 rounded font-bold"  // dernier coup
						: "bg-white text-yellow-800 rounded font-bold";     // anciens coups
				} else if (pending) {
					color = "bg-white text-red-700 font-extrabold rounded";
				} else {
					switch (special) {
						case "TW":
							color = "bg-red-500 text-white text-xs";
							break;
						case "DW":
						case "★":
							color = "bg-pink-300 text-xs";
							break;
						case "TL":
							color = "bg-blue-700 text-white text-xs";
							break;
						case "DL":
							color = "bg-blue-400 text-xs";
							break;
						default:
							color = "bg-green-600";
							break;
					}
				}

				return {
					x, y,
					char: displayed,
					points: /^[A-Z]$/.test(displayed) ? letterValues[displayed] : null,
					className: `${base} ${color}`
				};
			})
		);
	});
</script>

<div class="grid grid-cols-15 gap-[1px] border border-amber-500 w-full max-w-[95vw] mx-auto bg-amber-500">
	{#each $computedBoard as row}
		{#each row as cell (cell.x + '-' + cell.y)}
			<!-- svelte-ignore a11y_click_events_have_key_events -->
			<!-- svelte-ignore a11y_no_static_element_interactions -->
			<div
				class={cell.className}
				onclick={() => onPlaceLetter(cell.x, cell.y, game!.board[cell.y][cell.x])}
			>
				<span>{cell.char}</span>

				{#if cell.points !== null}
					<span class="absolute bottom-[-1.5px] right-[0px] text-[8px] text-gray-600">
						{cell.points}
					</span>
				{/if}
			</div>
		{/each}
	{/each}
</div>


<style>
	:global(.grid-cols-15) {
		grid-template-columns: repeat(15, minmax(0, 1fr));
	}
</style>
