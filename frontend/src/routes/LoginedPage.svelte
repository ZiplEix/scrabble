<script lang="ts">
    import { goto } from "$app/navigation";
    import { api } from "$lib/api";
    import GameList from "$lib/components/GameList.svelte";
    import { user } from "$lib/stores/user";
    import type { GameSummary } from "$lib/types/game_summary";
    import { onMount } from "svelte";

    let games: GameSummary[] = $state<GameSummary[]>([]);
    let tab: 'myturn' | 'ongoing' | 'finished' = $state('myturn');
    let q = $state('');

    onMount(async () => {
        if ($user) {
            try {
                const res = await api.get('/game');
                games = res.data.games;
            } catch (err) {
                console.error('Erreur en récupérant les parties', err);
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
            await api.delete(`/game/${id}`);
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
            await api.put(`/game/${id}/rename`, { new_name: newName });
            games = games.map(g => (g.id === id ? { ...g, name: newName } : g));
        } catch (err) {
            alert('Erreur lors du renommage de la partie');
        }
    }
</script>

<div class="min-h-[100dvh]">
	<!-- HEADER (centré, sans bordure) -->
	<header class="px-4 pt-5 text-center">
		<h1 class="text-3xl font-extrabold tracking-tight bg-gradient-to-r from-emerald-600 to-green-500 bg-clip-text text-transparent">
			Scrabble
		</h1>
		<p class="mt-1 text-[12px] text-gray-600">Joue avec tes amis, partout.</p>
	</header>

	<!-- HERO + CTA avec dégradé vert -->
	<section class="px-4 pt-4 pb-3">
		<div class="max-w-2xl mx-auto">
			<div class="rounded-2xl bg-emerald-50 ring-1 ring-black/5 p-4">
				<div class="flex items-center justify-between gap-3">
					<div class="min-w-0">
						<h2 class="text-lg font-bold text-gray-900 truncate">Bonjour {$user?.username ?? 'joueur'} 👋</h2>
						<p class="text-[12px] text-gray-700">Reprenez une partie ou lancez-en une nouvelle.</p>
					</div>
					<button onclick={createGame} class="inline-flex items-center gap-2 bg-green-600 text-white p-2 rounded-full shadow hover:bg-green-700">
						<svg width="18" height="18" viewBox="0 0 24 24" fill="none" aria-hidden="true"><path d="M12 5v14M5 12h14" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg>
						<span class="hidden sm:inline">Nouvelle</span>
					</button>
				</div>
			</div>
		</div>
	</section>

	<!-- SEGMENTED CONTROL + SEARCH -->
	<section class="px-4 pt-4 gap-2 flex flex-col mb-2">
		<div class="w-full max-w-2xl mx-auto">
			<div class="w-full rounded-full bg-white ring-1 ring-black/5 p-1 shadow-sm">
				<div class="grid grid-cols-3 gap-1">
					<button
						class="w-full text-center px-3 py-1.5 text-sm rounded-full transition {tab==='myturn' ? 'bg-emerald-600 text-white shadow font-bold' : 'text-gray-700 hover:bg-gray-50'}"
						aria-pressed={tab==='myturn'}
						onclick={() => tab='myturn'}
					>
						À mon tour
					</button>
					<button
						class="w-full text-center px-3 py-1.5 text-sm rounded-full transition {tab==='ongoing' ? 'bg-emerald-600 text-white shadow font-bold' : 'text-gray-700 hover:bg-gray-50'}"
						aria-pressed={tab==='ongoing'}
						onclick={() => tab='ongoing'}
					>
						En cours
					</button>
					<button
						class="w-full text-center px-3 py-1.5 text-sm rounded-full transition {tab==='finished' ? 'bg-emerald-600 text-white shadow font-bold' : 'text-gray-700 hover:bg-gray-50'}"
						aria-pressed={tab==='finished'}
						onclick={() => tab='finished'}
					>
						Terminées
					</button>
				</div>
			</div>
		</div>
		<div class="w-full max-w-2xl mx-auto">
			<div class="relative">
				<svg class="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-gray-500" width="16" height="16" viewBox="0 0 24 24" fill="none" aria-hidden="true">
					<circle cx="11" cy="11" r="7" stroke="currentColor" stroke-width="2" />
					<path d="M20 20l-3.5-3.5" stroke="currentColor" stroke-width="2" stroke-linecap="round" />
				</svg>
				<input
					class="w-full h-10 rounded-full bg-white ring-1 ring-black/5 shadow-sm pr-3 pl-9 text-sm placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-emerald-500/50"
					placeholder="Rechercher une partie..."
					bind:value={q}
					type="search"
				/>
			</div>
		</div>
	</section>

	<!-- LIST -->
	<section class="px-4 py-3">
		<div class="max-w-2xl mx-auto">
			{#if tab === 'myturn'}
				<GameList
					games={myTurnGames.filter(g => g.name.toLowerCase().includes(q.toLowerCase()))}
					{onDelete}
					{onRename}
					placeholder="Aucune partie à jouer pour le moment."
					showTurnOf={true}
					showLastPlayTime={true}
					winning={false}
				/>
			{:else if tab === 'ongoing'}
				<GameList
					games={ongoingGame.filter(g => g.name.toLowerCase().includes(q.toLowerCase()))}
					{onDelete}
					{onRename}
					placeholder="Aucune partie en cours."
					showTurnOf={true}
					showLastPlayTime={true}
					winning={false}
				/>
			{:else}
				<GameList
					games={finishedGames.filter(g => g.name.toLowerCase().includes(q.toLowerCase()))}
					{onDelete}
					placeholder="Aucune partie terminée."
					{onRename}
					winning={true}
					showTurnOf={false}
					showLastPlayTime={false}
				/>
			{/if}
		</div>
	</section>

	<!-- Actions globales (hors parties) -->
	<div class="fixed right-4 bottom-[calc(env(safe-area-inset-bottom)+16px)] flex flex-col items-end gap-2 z-40">
		<!-- Secondaires -->
		<a href="/report" title="Reports" class="inline-flex items-center justify-center w-11 h-11 rounded-full bg-white ring-1 ring-black/5 shadow hover:bg-gray-50" aria-label="Reports">
			<svg width="18" height="18" viewBox="0 0 24 24" fill="none" aria-hidden="true"><path d="M4 4h16v16H4z" stroke="currentColor" stroke-width="2"/><path d="M8 8h8M8 12h8M8 16h6" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg>
		</a>
		<a href="/me" title="Profil" class="inline-flex items-center justify-center w-11 h-11 rounded-full bg-white ring-1 ring-black/5 shadow hover:bg-gray-50" aria-label="Profil">
			<svg width="18" height="18" viewBox="0 0 24 24" fill="none" aria-hidden="true"><path d="M12 12a5 5 0 1 0 0-10 5 5 0 0 0 0 10Z" stroke="currentColor" stroke-width="2"/><path d="M4 20a8 8 0 0 1 16 0" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg>
		</a>
		<!-- Principale: créer une partie -->
		<button
			onclick={createGame}
			class="inline-flex items-center justify-center w-14 h-14 rounded-full bg-green-600 text-white shadow-lg hover:bg-green-700 active:scale-95 transition"
			style="padding-bottom: calc(env(safe-area-inset-bottom) / 4);"
			aria-label="Créer une nouvelle partie"
			title="Nouvelle partie"
		>
			<svg width="22" height="22" viewBox="0 0 24 24" fill="none" aria-hidden="true"><path d="M12 5v14M5 12h14" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg>
		</button>
	</div>
</div>
