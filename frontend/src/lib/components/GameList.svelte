<script lang="ts">
	import GameItem from './GameItem.svelte';
  	import type { GameSummary } from '$lib/types/game_summary';

	export let games: GameSummary[];
	export let onDelete: (id: string) => Promise<void>;
	export let onRename: (id: string, currentName: string)=> Promise<void>;
</script>

{#if !games || !games.length || games.length === 0}
	<p class="text-center text-gray-500 italic mt-6">Aucune partie en cours.</p>
{:else}
	<div class="flex flex-col gap-3">
		{#each games as game}
			<GameItem
				{game}
				{onDelete}
				{onRename}
			/>
		{/each}
	</div>
{/if}
