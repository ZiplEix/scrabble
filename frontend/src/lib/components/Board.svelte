<script lang="ts">
    import { specialCells } from '$lib/cells';
	import { pendingMove } from '$lib/stores/pendingMove';
	import { createEventDispatcher } from 'svelte';

	export let game: any;

	const dispatch = createEventDispatcher();
</script>

<div class="grid grid-cols-15 gap-[1px] border border-gray-400 w-full max-w-[95vw] mx-auto bg-black">
	{#each game.board as row, y}
		{#each row as cell, x}
			{@const key = `${y},${x}`}
			{@const type = specialCells.get(key)}
			{@const pending = $pendingMove.find(p => p.x === x && p.y === y)}
			{@const displayed = cell || pending?.letter || type}
			{@const isPlacedLetter = cell !== "" && !pending}
			{@const bg = isPlacedLetter
				? "bg-yellow-100 text-yellow-800 font-bold"
				: type === "TW" ? "bg-red-500 text-white"
				: type === "DW" || type === "★" ? "bg-pink-400 text-white"
				: type === "TL" ? "bg-blue-800 text-white"
				: type === "DL" ? "bg-blue-300"
				: "bg-white"}

			<!-- svelte-ignore a11y_click_events_have_key_events -->
			<!-- svelte-ignore a11y_no_static_element_interactions -->
			 <div
				class={`aspect-square w-full text-center text-sm flex items-center justify-center border border-gray-300
					${bg}
					${pending ? 'bg-red-200 text-red-700 font-extrabold' : ''}
					cursor-pointer select-none font-bold`}
				on:click={() => {
					dispatch('placeLetter', { x, y, cell });
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
