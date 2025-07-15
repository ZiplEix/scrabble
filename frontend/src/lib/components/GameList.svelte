<script lang="ts">
	import GameItem from './GameItem.svelte';
  	import type { GameSummary } from '$lib/types/game_summary';

	export let games: GameSummary[];
	export let onDelete: (id: string) => Promise<void>;
	export let onRename: (id: string, currentName: string)=> Promise<void>;
</script>

{#if games.length === 0}
	<p class="text-gray-600 italic">Aucune partie en cours.</p>
{:else}
	<div class="space-y-3">
		{#each games as game}
			<GameItem
				{game}
				{onDelete}
				{onRename}
			/>
		{/each}
	</div>
{/if}
