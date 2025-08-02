<script lang="ts">
	import GameItem from './GameItem.svelte';
  	import type { GameSummary } from '$lib/types/game_summary';

	let { games, onDelete, onRename, placeholder, showTurnOf = false, showLastPlayTime = false, winning = false }: {
		games: GameSummary[];
		onDelete: (id: string) => Promise<void>;
		onRename: (id: string, currentName: string) => Promise<void>;
		placeholder: string;
		showTurnOf: boolean;
		showLastPlayTime: boolean;
		winning: boolean;
	} = $props();

	let onScreenGameLimit = $state(3);

	let onScreenGame = $derived(games.length > onScreenGameLimit ? games.slice(0, onScreenGameLimit) : games);
</script>

<div class="mb-6">
	{#if !games || !games.length || games.length === 0}
		<p class="text-center text-gray-500 italic">{placeholder}</p>
	{:else}
		<div class="flex flex-col gap-3 mb-4">
			{#each onScreenGame as game}
				<GameItem
					{game}
					{onDelete}
					{onRename}
					{showTurnOf}
					{showLastPlayTime}
					{winning}
				/>
			{/each}
		</div>
		<!-- {#if games.length > 3}
			<button class="text-blue-500 hover:underline" onclick={() => alert('Voir plus de parties...')}>
				Afficher plus
			</button>
		{/if} -->
	{/if}
</div>
