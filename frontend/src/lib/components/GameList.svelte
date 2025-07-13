<script lang="ts">
	import GameItem from './GameItem.svelte';

	export let games: {
		id: string;
		name: string;
		current_turn_username: string;
	}[];
</script>

{#if games.length === 0}
	<p class="text-gray-600 italic">Aucune partie en cours.</p>
{:else}
	<div class="space-y-3">
		{#each games as game}
			<GameItem {game} />
		{/each}
	</div>
{/if}
