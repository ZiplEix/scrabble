<script lang="ts">
    import { specialCells } from '$lib/cells';
	import { pendingMove } from '$lib/stores/pendingMove';
	import { createEventDispatcher } from 'svelte';

	export let game: any;
	export let onPlaceLetter: (x: number, y: number, cell: string) => void;
</script>

<div class="grid grid-cols-15 gap-[1px] border border-amber-500 w-full max-w-[95vw] mx-auto bg-amber-500">
	{#each game.board as row, y}
		{#each row as cell, x}
			{@const key = `${y},${x}`}
			{@const type = specialCells.get(key)}
			{@const pending = $pendingMove.find(p => p.x === x && p.y === y)}
			{@const displayed = cell || pending?.letter || type}
			{@const isPlacedLetter = cell !== "" && !pending}
			{@const bg = isPlacedLetter
				? "bg-yellow-100 text-yellow-800 font-bold rounded"
				: type === "TW" ? "bg-red-400 text-white"
				: type === "DW" || type === "★" ? "bg-pink-300"
				: type === "TL" ? "bg-blue-500 text-white"
				: type === "DL" ? "bg-blue-200"
				: "bg-green-200"}

			<!-- svelte-ignore a11y_click_events_have_key_events -->
			<!-- svelte-ignore a11y_no_static_element_interactions -->
			 <div
				class={`aspect-square w-full text-center text-sm flex items-center justify-center border border-gray-300
					${bg}
					${pending ? 'bg-yellow-100 text-red-700 font-extrabold rounded' : ''}
					cursor-pointer select-none font-bold`}
				on:click={() => {
					onPlaceLetter(x, y, cell);
				}}
			>
				{displayed}
			</div>
		{/each}
	{/each}
</div>

<style>
	:global(.grid-cols-15) {
		grid-template-columns: repeat(15, minmax(0, 1fr));
	}
</style>
