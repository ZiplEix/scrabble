<script lang="ts">
  	import type { GameSummary } from '$lib/types/game_summary';
	import { createEventDispatcher } from 'svelte';

	export let game: GameSummary;

	const dispatch = createEventDispatcher();

	let menuOpen = false;

	function toggleMenu(e: MouseEvent) {
		e.preventDefault();
		e.stopPropagation();
		menuOpen = !menuOpen;
	}

	// Close menu on outside click
	function handleClickOutside() {
		menuOpen = false;
	}

	// Format date lisible
	function formatDate(d: string) {
		const date = new Date(d);
		return date.toLocaleString('fr-FR', {
			year: 'numeric',
			month: 'short',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit',
		});
	}

	function renameGame() {
		dispatch('rename', { id: game.id, currentName: game.name });
		menuOpen = false;
	}

	function deleteGame() {
		dispatch('delete', { id: game.id });
		menuOpen = false;
	}
</script>

<div class="relative border rounded p-4 hover:bg-gray-100 flex items-center justify-between">
	<div class="flex-grow cursor-pointer" on:click={() => window.location.href = `/games/${game.id}`}>
		<h2 class="text-lg font-semibold text-blue-600 hover:underline">{game.name}</h2>
		<p class="text-sm text-gray-600 mb-1">Tour de : <strong>{game.current_turn_username}</strong></p>
		<p class="text-xs text-gray-500">Dernier coup joué : {formatDate(game.last_play_time)}</p>
	</div>

	{#if game.is_your_game}
		<div class="relative">
			<button
				aria-label="Menu options"
				class="p-2 hover:bg-gray-300 rounded"
				on:click={toggleMenu}
				type="button"
			>
				<!-- Icône 3 points verticaux -->
				<svg class="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
					<path d="M10 6a2 2 0 110-4 2 2 0 010 4zm0 4a2 2 0 110-4 2 2 0 010 4zm0 4a2 2 0 110-4 2 2 0 010 4z" />
				</svg>
			</button>

			{#if menuOpen}
				<!-- svelte-ignore a11y_click_events_have_key_events -->
				<!-- svelte-ignore a11y_no_static_element_interactions -->
				<div
					class="absolute right-0 mt-1 w-40 bg-white border rounded shadow-lg z-10"
					on:click|stopPropagation
				>
					<button
						class="block w-full px-4 py-2 text-left hover:bg-gray-200"
						on:click={renameGame}
					>
						Renommer
					</button>
					<button
						class="block w-full px-4 py-2 text-left text-red-600 hover:bg-gray-200"
						on:click={deleteGame}
					>
						Supprimer
					</button>
				</div>
			{/if}
		</div>
	{/if}
</div>

<!-- Fermer menu au clic hors du composant -->
<svelte:window on:click={handleClickOutside} />
