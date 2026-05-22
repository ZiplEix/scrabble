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

<div class="relative rounded-3xl bg-white border border-stone-200/60 shadow-sm p-4 hover:shadow-md transition-all group">
	<!-- Link to game page -->
	<a href={`/games/${game.id}`} class="flex flex-col gap-2.5">
		
		<!-- Row 1: Game name + Status badge -->
		<div class="flex items-center justify-between pr-6 gap-2">
			<h3 class="text-sm font-extrabold text-stone-800 truncate" title={game.name}>
				{game.name}
			</h3>
			
			{#if showTurnOf}
				<!-- Status bubble -->
				<span class="shrink-0 text-[10px] font-bold px-2 py-0.5 rounded-full ring-1 {game.current_turn_username === game.winner_username ? 'bg-amber-50 text-brand-gold-hover ring-brand-gold/20' : 'bg-brand-emerald-light text-brand-emerald ring-brand-emerald/10'}">
					{game.current_turn_username}
				</span>
			{/if}
		</div>

		<!-- Row 2: Secondary info -->
		<div class="flex flex-col gap-1 text-[11px] text-stone-500 font-medium">
			{#if showLastPlayTime}
				<div class="flex items-center gap-1">
					<span>🕒</span>
					<span>Joué le {formatDate(game.last_play_time)}</span>
				</div>
			{/if}

			{#if winning}
				<div class="flex items-center gap-1 bg-amber-50/50 border border-brand-gold/10 rounded-xl px-2.5 py-1.5 mt-1 text-stone-700">
					<span class="text-xs">👑</span>
					<span>Gagné par <strong class="text-brand-gold-hover">{game.winner_username}</strong></span>
				</div>
			{/if}
		</div>

	</a>

	<!-- Premium Menu Option -->
	{#if game.is_your_game}
		<div class="absolute top-3 right-3 z-10">
			<button
				aria-label="Menu"
				class="w-8 h-8 flex items-center justify-center text-stone-400 rounded-full hover:bg-stone-100 active:scale-90 transition-all"
				onclick={toggleMenu}
			>
				<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" class="w-4 h-4" viewBox="0 0 16 16">
  					<path d="M9.5 13a1.5 1.5 0 1 1-3 0 1.5 1.5 0 0 1 3 0m0-5a1.5 1.5 0 1 1-3 0 1.5 1.5 0 0 1 3 0m0-5a1.5 1.5 0 1 1-3 0 1.5 1.5 0 0 1 3 0"/>
				</svg>
			</button>

			{#if menuOpen}
				<!-- svelte-ignore a11y_click_events_have_key_events -->
				<!-- svelte-ignore a11y_no_static_element_interactions -->
				<div
					class="absolute right-0 mt-1 w-32 bg-white/95 backdrop-blur-md border border-stone-200/80 rounded-2xl shadow-xl z-30 text-xs py-1.5 overflow-hidden"
					onclick={(e) => e.stopPropagation()}
				>
					<button class="block w-full px-4 py-2 text-left hover:bg-stone-50 font-semibold text-stone-700" onclick={renameGame}>
						Renommer
					</button>
					<button class="block w-full px-4 py-2 text-left text-red-600 hover:bg-red-50 font-semibold border-t border-stone-100" onclick={deleteGame}>
						Supprimer
					</button>
				</div>
			{/if}
		</div>
	{/if}
</div>

<svelte:window onclick={handleClickOutside} />
