<script lang='ts'>
	import { goto } from "$app/navigation";
	import { page } from "$app/stores";
	import { api } from "$lib/api";
	import { gameStore } from "$lib/stores/game";
	import { onMount } from "svelte";
	import { get, writable } from "svelte/store";
	import GameMenu from "$lib/components/GameMenu.svelte";
	import UserLink from "$lib/components/UserLink.svelte";
	import type { GameInfo } from "$lib/types/game_infos";

    let game: GameInfo | null = $state<GameInfo | null>(null);
	let error = $state('');
	let loading = $state(true);
    let showScore = writable<boolean>(false);

	onMount(async () => {
		const id = $page.params.id;
		const stored = get(gameStore);
		if (stored?.id === id) {
			game = stored;
			loading = false;
			return;
		}

		try {
			const res = await api.get<GameInfo>(`/game/${id}`);
			game = res.data;
			gameStore.set(res.data);
		} catch (e: any) {
			error = e?.response?.data?.message || 'Impossible de charger l’historique.';
		} finally {
			loading = false;
		}
	})

	function getUsername(pid: number) {
		return game?.players.find(p => p.id === pid)?.username ?? '–';
	}

	function backToGame() {
		if (game) goto(`/games/${game.id}`);
		else goto('/');
	}
</script>

{#if loading}
  	<p class="mt-8 text-center text-gray-600">Chargement de l’historique…</p>
{:else if error}
  	<p class="mt-8 text-center text-red-600">{error}</p>
{:else if game}
	<!-- Conteneur gradient comme la page de jeu -->
	<div class="flex flex-col overflow-hidden bg-gradient-to-b from-emerald-50 to-white" style="min-height: 100dvh;">
		<!-- Header aligné GameHeader (ligne 1) -->
		<header class="px-3 pt-2 pb-2">
			<div class="flex items-center w-full justify-between gap-2">
				<div class="flex items-center gap-2 min-w-0">
					<button
						class="p-2 rounded-lg hover:bg-white/60 ring-1 ring-black/5 bg-white/40 backdrop-blur-sm"
						aria-label="Retour à la partie"
						onclick={backToGame}
						title="Retour"
					>
						<svg width="18" height="18" viewBox="0 0 24 24" fill="none" aria-hidden="true">
							<path d="M4 12h16M10 6l-6 6 6 6" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
						</svg>
					</button>
					<h2 class="text-base font-semibold text-gray-900 truncate" title={game.name}>{game.name}</h2>
				</div>

				<div class="flex items-center gap-2">
					<button
						class="hidden sm:flex items-center gap-1 px-2.5 h-8 rounded-lg bg-emerald-600/90 hover:bg-emerald-600 text-white text-[12px] font-medium shadow-sm ring-1 ring-emerald-700/30"
						onclick={() => showScore.set(true)}
						title="Voir le classement"
						aria-label="Voir le classement"
					>
						<span>🏆</span>
						<span>Scores</span>
					</button>
					<GameMenu showScores={showScore} gameId={game.id} />
				</div>
			</div>
			<div class="mt-2 h-[3px] rounded-full bg-gradient-to-r from-emerald-200 via-emerald-500/40 to-emerald-200"></div>
		</header>

		<!-- Contenu -->
		<main class="flex-1 px-3 pb-[max(16px,env(safe-area-inset-bottom))]">
			<h1 class="text-lg font-semibold text-gray-800 text-center my-3">Historique de “{game.name}”</h1>

			{#each game.moves as move, idx}
				<article class="rounded-2xl ring-1 ring-black/5 bg-white shadow p-4 space-y-3 mb-3">
					<!-- Ligne 1: index + mot/dir + score/pass en chip -->
					<div class="flex items-center justify-between gap-2">
						<div class="flex items-center gap-2 min-w-0">
							<span class="inline-flex items-center px-2 py-0.5 rounded-full bg-gray-100 text-[11px] text-gray-700 ring-1 ring-black/5">#{idx + 1}</span>
							{#if !move.move.type}
								<span class="font-semibold text-gray-900 truncate" title={move.move.word}>{move.move.word}</span>
								<span class="inline-flex items-center px-2 py-0.5 rounded-full bg-blue-100 text-blue-800 text-[10px] uppercase ring-1 ring-blue-300/60">{move.move.dir}</span>
							{/if}
						</div>
						<div>
							{#if move.move.score}
								<span class="inline-flex items-center gap-1 px-2.5 py-1 rounded-full bg-emerald-100 text-emerald-800 text-[11px] ring-1 ring-emerald-300/60 tabular-nums font-semibold">+{move.move.score}</span>
							{:else}
								<span class="inline-flex items-center gap-1 px-2.5 py-1 rounded-full bg-amber-100 text-amber-800 text-[11px] ring-1 ring-amber-300/60">Pass</span>
							{/if}
						</div>
					</div>

					<!-- Ligne 2: joueur + heure + date en petites infos -->
									<div class="flex flex-wrap items-center gap-2 text-xs text-gray-600">
										<span class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full bg-gray-50 ring-1 ring-black/5">
											Par <UserLink id={move.player_id} username={getUsername(move.player_id)} />
										</span>
						<span class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full bg-gray-50 ring-1 ring-black/5">
							{new Date(move.played_at).toLocaleTimeString('fr-FR', { hour: '2-digit', minute: '2-digit' })}
						</span>
						<span class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full bg-gray-50 ring-1 ring-black/5">
							{new Date(move.played_at).toLocaleDateString('fr-FR')}
						</span>
					</div>

					<!-- Ligne 3: positions des lettres avec tuiles -->
					{#if move.move.letters?.length}
						<div class="flex flex-wrap gap-1 mt-1">
							{#each move.move.letters as l}
								<div
									class="grid place-items-center w-7 h-7 rounded bg-amber-100 text-amber-900 font-semibold ring-1 ring-amber-300/50 shadow-sm"
									title={`(${l.x},${l.y})`}
								>
									<span class="text-[12px] leading-none">{l.char}</span>
								</div>
							{/each}
						</div>
					{/if}
				</article>
			{/each}

			{#if game.moves.length === 0}
				<p class="text-center text-gray-500 mt-8">Aucun coup joué pour l’instant.</p>
			{/if}
		</main>
	</div>
{/if}
