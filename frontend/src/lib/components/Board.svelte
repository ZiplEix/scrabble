<script lang="ts">
    import { specialCells } from '$lib/cells';
  import { letterValues } from '$lib/lettres_value';
	import { pendingMove } from '$lib/stores/pendingMove';

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
				? "bg-white text-yellow-800 rounded font-bold"
				: type === "TW" ? "bg-red-500 text-white text-xs"
				: type === "DW" || type === "★" ? "bg-pink-300 text-xs"
				: type === "TL" ? "bg-blue-700 text-white text-xs"
				: type === "DL" ? "bg-blue-400 text-xs"
				: "bg-green-600"}

			<!-- svelte-ignore a11y_click_events_have_key_events -->
			<!-- svelte-ignore a11y_no_static_element_interactions -->
			<div
				class={`relative aspect-square w-full text-center flex items-center justify-center border border-gray-300
					cursor-pointer select-none overflow-hidden
					${bg}
					${pending ? 'bg-white text-red-700 font-extrabold rounded' : ''}`}
				on:click={() => {
					onPlaceLetter(x, y, cell);
				}}
			>
				<span>{displayed}</span>

				{#if /^[A-Z]$/.test(displayed)}
					<span class="absolute bottom-[-1.5px] right-[0px] text-[8px] text-gray-600">
						{letterValues[displayed]}
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
