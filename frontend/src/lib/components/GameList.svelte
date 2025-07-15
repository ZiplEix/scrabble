<script lang="ts">
  	import { createEventDispatcher } from 'svelte';
	import GameItem from './GameItem.svelte';
  	import type { GameSummary } from '$lib/types/game_summary';

	export let games: GameSummary[];

	const dispatch = createEventDispatcher();
</script>

{#if games.length === 0}
	<p class="text-gray-600 italic">Aucune partie en cours.</p>
{:else}
	<div class="space-y-3">
		{#each games as game}
			<GameItem
				{game}
				on:delete={(e) => dispatch('delete', e.detail)}
				on:rename={(e) => dispatch('rename', e.detail)}
			/>
		{/each}
	</div>
{/if}
