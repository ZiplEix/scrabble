<script lang="ts">
    import { goto } from "$app/navigation";
    import { getGames, getTodayPuzzle, deleteGame, renameGame } from "$lib/api";
    import GameList from "$lib/components/GameList.svelte";
    import { user } from "$lib/stores/user";
    import type { GameSummary } from "$lib/types/game_summary";
    import { onMount } from "svelte";

    let games: GameSummary[] = $state<GameSummary[]>([]);
    let tab: 'myturn' | 'ongoing' | 'finished' = $state('myturn');
    let q = $state('');
	let showDailyChallenge = $state(false);

    onMount(async () => {
        if ($user) {
            try {
				const [gamesList, puzzleToday] = await Promise.all([
					getGames(),
					getTodayPuzzle()
				]);

				games = gamesList;
				showDailyChallenge = puzzleToday?.has_player_attempted === false;
            } catch (err) {
                console.error('Erreur en récupérant les parties', err);
				showDailyChallenge = false;
            }
        }
    });

    let myTurnGames = $derived(games.filter(g => g.current_turn_username === $user?.username && g.status === 'ongoing'));

    let ongoingGame = $derived(games.filter(g => g.status === 'ongoing' && g.current_turn_username !== $user?.username));

    let finishedGames = $derived(games.filter(g => g.status === 'ended'));

    function createGame() {
        goto('/games/new');
    }

    async function onDelete(id: string) {
        if (!confirm('Voulez-vous vraiment supprimer cette partie ?')) return;
        try {
            await deleteGame(id);
            games = games.filter(g => g.id !== id);
        } catch (err) {
            alert('Erreur lors de la suppression de la partie');
        }
    }

    function renameGamePrompt(currentName: string): Promise<string | null> {
        const newName = prompt('Entrez le nouveau nom de la partie', currentName);
        return Promise.resolve(newName && newName.trim() !== '' ? newName.trim() : null);
    }

    async function onRename(id: string, currentName: string) {
        const newName = await renameGamePrompt(currentName);
        if (!newName) return;
        try {
            await renameGame(id, newName);
            games = games.map(g => (g.id === id ? { ...g, name: newName } : g));
        } catch (err) {
            alert('Erreur lors du renommage de la partie');
        }
    }
</script>

<div class="px-4 py-6 max-w-xl mx-auto flex flex-col gap-6">
	
	<!-- HEADER -->
	<header class="text-center">
		<h1 class="text-3xl font-black tracking-tight bg-gradient-to-r from-brand-emerald to-brand-gold bg-clip-text text-transparent select-none">
			SCRABBLE
		</h1>
		<p class="text-[11px] font-semibold text-stone-500 uppercase tracking-widest mt-1">Scrabble Club</p>
	</header>

	<!-- HERO + CTA CARD -->
	<section class="w-full">
		<div class="rounded-3xl glass-card border border-white/60 p-5 shadow-xl relative overflow-hidden">
			<!-- Spark decorations -->
			<div class="absolute -top-6 -right-6 w-20 h-20 rounded-full bg-brand-gold/10 blur-lg"></div>
			<div class="absolute -bottom-6 -left-6 w-20 h-20 rounded-full bg-brand-emerald/10 blur-lg"></div>

			<div class="flex items-center justify-between gap-4 relative z-10">
				<div class="min-w-0">
					<h2 class="text-xl font-extrabold text-stone-800 truncate">Bonjour {$user?.username ?? 'joueur'} 👋</h2>
					<p class="text-xs text-stone-500 mt-1 leading-normal">Reprenez une grille ou lancez un défi à un proche.</p>
				</div>
				<button 
					onclick={createGame} 
					class="shrink-0 flex items-center justify-center w-12 h-12 rounded-2xl bg-brand-emerald text-white shadow-lg shadow-brand-emerald/20 hover:bg-brand-emerald-hover active:scale-95 transition-all"
					aria-label="Nouvelle partie"
					title="Créer une nouvelle partie"
				>
					<svg class="w-6 h-6" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15"/>
					</svg>
				</button>
			</div>
		</div>

		<!-- Daily challenge banner -->
		{#if showDailyChallenge}
			<a
				href="/puzzles"
				class="mt-4 block rounded-3xl bg-gradient-to-r from-brand-gold-light to-amber-50/70 border border-brand-gold/20 p-4 hover:border-brand-gold/40 shadow-md active:scale-[0.99] transition-all"
				aria-label="Accéder au défi quotidien"
			>
				<div class="flex items-center justify-between gap-3">
					<div class="min-w-0">
						<div class="flex items-center gap-1.5">
							<span class="text-xs">🏆</span>
							<p class="text-[10px] font-bold uppercase tracking-wider text-brand-gold-hover">Défi du jour</p>
						</div>
						<p class="text-sm font-extrabold text-stone-800 mt-1">Votre entraînement quotidien attend !</p>
						<p class="text-[11px] text-stone-500 mt-0.5">Faites votre meilleure tentative et comparez votre score.</p>
					</div>
					<div class="shrink-0 flex items-center justify-center w-10 h-10 rounded-full bg-white text-brand-gold border border-brand-gold/30 shadow-sm">
						<svg class="w-5 h-5" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" d="M13.5 4.5L21 12m0 0l-7.5 7.5M21 12H3"/>
						</svg>
					</div>
				</div>
			</a>
		{/if}
	</section>

	<!-- SEGMENT CONTROL & SEARCH -->
	<section class="flex flex-col gap-3">
		<!-- Segment Controls (Rounded Tab) -->
		<div class="rounded-2xl bg-stone-200/50 p-1 border border-stone-200/20">
			<div class="grid grid-cols-3 gap-1">
				<button
					class="py-2.5 px-3 text-xs font-bold rounded-xl text-center transition-all {tab === 'myturn' ? 'bg-white text-brand-emerald shadow-md' : 'text-stone-600 hover:text-stone-800'}"
					aria-pressed={tab === 'myturn'}
					onclick={() => tab = 'myturn'}
				>
					À mon tour
				</button>
				<button
					class="py-2.5 px-3 text-xs font-bold rounded-xl text-center transition-all {tab === 'ongoing' ? 'bg-white text-brand-emerald shadow-md' : 'text-stone-600 hover:text-stone-800'}"
					aria-pressed={tab === 'ongoing'}
					onclick={() => tab = 'ongoing'}
				>
					En cours
				</button>
				<button
					class="py-2.5 px-3 text-xs font-bold rounded-xl text-center transition-all {tab === 'finished' ? 'bg-white text-brand-emerald shadow-md' : 'text-stone-600 hover:text-stone-800'}"
					aria-pressed={tab === 'finished'}
					onclick={() => tab = 'finished'}
				>
					Terminées
				</button>
			</div>
		</div>

		<!-- Search Bar -->
		<div class="relative">
			<svg class="pointer-events-none absolute left-4 top-1/2 -translate-y-1/2 text-stone-400 w-4 h-4" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24">
				<path stroke-linecap="round" stroke-linejoin="round" d="M21 21l-5.197-5.197m0 0A7.5 7.5 0 105.196 5.196a7.5 7.5 0 0010.607 10.607z"/>
			</svg>
			<input
				class="w-full h-11 bg-white/70 border border-stone-200/80 rounded-2xl pr-4 pl-11 text-xs placeholder-stone-400 focus:outline-none focus:ring-2 focus:ring-brand-emerald/40 focus:border-brand-emerald shadow-sm transition-all"
				placeholder="Rechercher une partie par nom..."
				bind:value={q}
				type="search"
			/>
		</div>
	</section>

	<!-- GAMES LIST -->
	<section class="flex-1">
		{#if tab === 'myturn'}
			<GameList
				games={myTurnGames.filter(g => g.name.toLowerCase().includes(q.toLowerCase()))}
				{onDelete}
				{onRename}
				placeholder="Aucune partie en attente de votre coup ! ✨"
				showTurnOf={true}
				showLastPlayTime={true}
				winning={false}
			/>
		{:else if tab === 'ongoing'}
			<GameList
				games={ongoingGame.filter(g => g.name.toLowerCase().includes(q.toLowerCase()))}
				{onDelete}
				{onRename}
				placeholder="Aucune autre partie en cours en ce moment."
				showTurnOf={true}
				showLastPlayTime={true}
				winning={false}
			/>
		{:else}
			<GameList
				games={finishedGames.filter(g => g.name.toLowerCase().includes(q.toLowerCase()))}
				{onDelete}
				placeholder="Aucune partie terminée pour l'instant."
				{onRename}
				winning={true}
				showTurnOf={false}
				showLastPlayTime={false}
			/>
		{/if}
	</section>

</div>
