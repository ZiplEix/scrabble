<script lang="ts">
  	import type { GameSummary } from '$lib/types/game_summary';

	export let game: GameSummary;
	export let onDelete: (id: string) => Promise<void>;
	export let onRename: (id: string, currentName: string)=> Promise<void>;
	export let showTurnOf: boolean = false;
	export let showLastPlayTime: boolean = false;
	export let winning: boolean = false;

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
		onRename(game.id, game.name)
		menuOpen = false;
	}

	function deleteGame() {
		onDelete(game.id)
		menuOpen = false;
	}
</script>

<div class="relative bg-slate-100 rounded shadow-md p-4 flex flex-col gap-2">
	<a href={`/games/${game.id}`} class="flex flex-col gap-1">
		<h2 class="text-xl font-bold text-green-700">{game.name}</h2>
		{#if showLastPlayTime}
			<p class="text-xs text-gray-400">Dernier coup : {formatDate(game.last_play_time)}</p>
		{/if}
		{#if showTurnOf}
			<p class="text-sm text-gray-600">Tour de : <span class="font-medium">{game.current_turn_username}</span></p>
		{/if}
		{#if winning}
			<p class="text-sm text-gray-600">Gagnant:
				<span class="font-medium">{game.winner_username}</span>
				, le {formatDate(game.last_play_time)}
			</p>
		{/if}
	</a>

	{#if game.is_your_game}
		<div class="absolute top-2 right-2">
			<button
				aria-label="Menu"
				class="p-1 text-gray-600 rounded hover:bg-gray-100 active:bg-gray-200"
				on:click={toggleMenu}
			>
				<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" class="bi bi-three-dots-vertical" viewBox="0 0 16 16">
  					<path d="M9.5 13a1.5 1.5 0 1 1-3 0 1.5 1.5 0 0 1 3 0m0-5a1.5 1.5 0 1 1-3 0 1.5 1.5 0 0 1 3 0m0-5a1.5 1.5 0 1 1-3 0 1.5 1.5 0 0 1 3 0"/>
				</svg>
			</button>

			{#if menuOpen}
				<!-- svelte-ignore a11y_click_events_have_key_events -->
				<!-- svelte-ignore a11y_no_static_element_interactions -->
				<div
					class="absolute right-0 mt-2 w-36 bg-white border rounded shadow z-20 text-sm"
					on:click|stopPropagation
				>
					<button class="block w-full px-4 py-2 text-left hover:bg-gray-100" on:click={renameGame}>
						Renommer
					</button>
					<button class="block w-full px-4 py-2 text-left text-red-600 hover:bg-gray-100" on:click={deleteGame}>
						Supprimer
					</button>
				</div>
			{/if}
		</div>
	{/if}
</div>

<svelte:window on:click={handleClickOutside} />
