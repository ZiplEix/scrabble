<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import { page } from '$app/stores';
	import { get, writable } from 'svelte/store';
	import { pendingMove } from '$lib/stores/pendingMove';
	import { user } from '$lib/stores/user';
	import { gameStore } from '$lib/stores/game';
	import { useBoardGame } from '$lib/hooks/useBoardGame.svelte';
	import GameHeader from '$lib/components/GameHeader.svelte';
	import GameSkeleton from '$lib/components/GameSkeleton.svelte';
	import GameBoard from '$lib/components/GameBoard.svelte';
	import type { GameInfo } from '$lib/types/game_infos';

	let gameId = $state<string | null>(null);
	let game = $state<GameInfo | null>(null);
	let error = $state('');
	let loading = $state(true);
	let showScores = $state(writable<boolean>(false));

	let sortedPlayers = $derived(game ? [...game.players].sort((a, b) => b.score - a.score) : []);

	const boardGame = useBoardGame({
		get simulateScoreEndpoint() { return gameId ? `/game/${gameId}/simulate_score` : ''; },
		onSubmit: async (payload) => {
			await api.post(`/game/${gameId}/play`, payload);
			const res = await api.get(`/game/${gameId}`);
			game = res.data;
			gameStore.set(game);
			if (game?.status === 'ended') showScores.set(true);
			return game!.your_rack.split('');
		}
	});

	onMount(async () => {
		gameId = $page.params.id ?? null;
		if (!gameId) return;

		try {
			if (gameStore && $gameStore?.id === gameId) {
				game = $gameStore;
			} else {
				const res = await api.get(`/game/${gameId}`);
				game = res.data;
				gameStore.set(game);
			}

			pendingMove.set([]);
			boardGame.setRackFromString(game!.your_rack);
		} catch (e: any) {
			error = e?.response?.data?.error || 'Erreur lors du chargement de la partie';
		} finally {
			loading = false;
			if (game?.status === 'ended') showScores.set(true);
		}
	});

	async function retryLoad() {
		if (!gameId) return;
		loading = true;
		error = '';
		try {
			const res = await api.get(`/game/${gameId}`);
			game = res.data;
			gameStore.set(game);
			pendingMove.set([]);
			boardGame.setRackFromString(game!.your_rack);
		} catch (e: any) {
			error = e?.response?.data?.error || 'Erreur lors du chargement de la partie';
		} finally {
			loading = false;
		}
	}

	const adminEmail = 'leroyerbaptiste@gmail.com';
	function buildAdminMailto() {
		const subject = encodeURIComponent(`Problème partie ${gameId}`);
		const body = encodeURIComponent(`ID partie: ${gameId}\nErreur: ${error}\nURL: ${typeof location !== 'undefined' ? location.href : ''}`);
		return `mailto:${adminEmail}?subject=${subject}&body=${body}`;
	}

	async function drawNewRack() {
		const ok = confirm('Êtes-vous sûr de vouloir changer toutes vos lettres ? Cela remplacera vos lettres actuelles et passera votre tour.');
		if (!ok) return;
		try {
			const res = await api.get(`/game/${gameId}/new_rack`);
			const newRack = res.data as string[];
			if (newRack.length === 0) {
				alert('Plus de lettres disponibles dans le sac.');
				return;
			}
			boardGame.setRackFromArray(newRack);
			pendingMove.set([]);
		} catch (e: any) {
			alert(e?.response?.data?.message || 'Erreur lors du tirage d\'un nouveau rack.');
		}
	}

	async function passTurn() {
		const ok = confirm('Êtes-vous sûr de vouloir passer votre tour ?');
		if (!ok) return;
		try {
			await api.post(`/game/${gameId}/pass`);
			const res = await api.get(`/game/${gameId}`);
			game = res.data;
			gameStore.set(game);
		} catch (e: any) {
			alert(e?.response?.data?.message || 'Erreur lors du passage du tour.');
		}
	}

	async function handleRematch() {
		if (!game?.is_your_game) {
			alert('Seul le créateur de la partie peut créer une revanche.');
			return;
		}
		const defaultName = `${game!.name} – revanche`;
		const newName = prompt('Nom de la nouvelle partie :', defaultName);
		if (!newName) return;

		const currentUsername = get(user)?.username;
		const opponents = game!.players.map((p) => p.username).filter((u) => u && u !== currentUsername);

		try {
			const res = await api.post('/game', { name: newName, players: opponents });
			window.location.href = `/games/${res.data.game_id}`;
		} catch (err: any) {
			alert(err?.response?.data?.message || 'Impossible de créer la revanche.');
		}
	}
</script>

{#if loading}
	<GameSkeleton />
{:else if error}
	<div class="flex-1 grid place-items-center px-3" style="min-height: calc(100dvh - 64px);">
		<div class="w-full max-w-[640px]">
			<div class="mx-auto rounded-2xl bg-white/90 backdrop-blur-md ring-1 ring-black/5 shadow-lg p-6 text-center">
				<div class="mx-auto w-12 h-12 rounded-full bg-amber-100/60 flex items-center justify-center text-amber-700 mb-3">
					<svg width="20" height="20" viewBox="0 0 24 24" fill="none" aria-hidden="true"><path d="M12 9v4" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/><path d="M12 17h.01" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>
				</div>
				<h3 class="text-lg font-semibold text-gray-800">Impossible de charger la partie</h3>
				<p class="mt-2 text-sm text-gray-600">{error}</p>
				<p class="mt-1 text-sm text-gray-600">Si le problème persiste, merci d'ouvrir un ticket ou de contacter un administrateur.</p>
				<div class="mt-4 flex gap-2 justify-center">
					<button class="px-4 py-2 bg-emerald-100 text-emerald-800 rounded-md hover:bg-emerald-200" onclick={retryLoad}>Réessayer</button>
					<a href="/" class="px-4 py-2 bg-white ring-1 ring-black/5 text-gray-700 rounded-md hover:bg-gray-50">Accueil</a>
					<a href={buildAdminMailto()} class="px-4 py-2 bg-white ring-1 ring-black/5 text-gray-700 rounded-md hover:bg-gray-50">Signaler</a>
				</div>
			</div>
		</div>
	</div>
{:else if game}
	<div class="flex flex-col overflow-hidden bg-gradient-to-b from-emerald-50 to-white" style="height: 100dvh;">
		<GameHeader {game} {showScores} gameId={gameId!} />

		<main class="flex-1 flex flex-col min-h-0 overflow-hidden">
			<GameBoard
				{game}
				visibleRack={boardGame.visibleRack}
				originalRack={boardGame.originalRack}
				moveScore={boardGame.moveScore}
				submitting={boardGame.submitting()}
				onDropFromRack={(char, x, y, id) => boardGame.dropFromRack(char, x, y, id, game!.board[y]?.[x] ?? '')}
				onTakeFromBoard={boardGame.takeBackFromBoard}
				onCancelPendingMove={boardGame.cancelPendingMove}
				onShuffleRack={boardGame.shuffleRack}
				onPlayMove={boardGame.playMove}
			>
				{#snippet extraIdleActions()}
					<!-- Échanger -->
					<button
						class="h-12 px-2 flex flex-col items-center justify-center text-[12px] font-medium active:scale-[0.98] transition"
						onclick={drawNewRack}
						aria-label="Échanger les lettres"
					>
						<svg width="18" height="18" viewBox="0 0 24 24" fill="none" aria-hidden="true">
							<path d="M4 7h11l-3-3M20 17H9l3 3" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
						</svg>
						<span>Échanger</span>
					</button>
					<!-- Passer -->
					<button
						class="h-12 px-2 flex flex-col items-center justify-center text-[12px] font-medium active:scale-[0.98] transition"
						onclick={passTurn}
						aria-label="Passer le tour"
					>
						<svg width="18" height="18" viewBox="0 0 24 24" fill="none" aria-hidden="true">
							<path d="M5 12h14" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
							<path d="M12 5v14" stroke="currentColor" stroke-width="2" stroke-linecap="round" opacity=".15"/>
						</svg>
						<span>Passer</span>
					</button>
				{/snippet}
			</GameBoard>
		</main>
	</div>

	<!-- Modal classement -->
	{#if $showScores}
		<div class="fixed inset-0 bg-black/30 backdrop-blur-sm flex items-center justify-center z-50">
			<div class="bg-white rounded-lg shadow-lg w-[90vw] max-w-sm p-6">
				{#if game?.status === 'ended'}
					<h3 class="text-lg font-semibold mb-2 text-center">
						🎉 Partie terminée<br />
						Vainqueur : <span class="text-green-600 font-bold">{game.winner_username}</span>
					</h3>
					<p class="text-center text-sm text-gray-600 mb-4">
						Terminée le {new Date(game.ended_at!).toLocaleString('fr-FR')}
					</p>
				{:else}
					<h3 class="text-lg font-semibold mb-4 text-center">Classement</h3>
				{/if}
				<div class="flex flex-col gap-2">
					{#each (game.status === 'ended' ? sortedPlayers : game.players) as player, i}
						{@const playerClass = game.status === 'ended' ? 'bg-gray-50' : player.id === game.current_turn ? 'bg-green-100 border-green-400' : 'bg-gray-50'}
						<a
							class="flex justify-between items-center p-2 rounded border {playerClass}"
							href={player.id !== $user?.id ? `/user/${player.id}` : undefined}
						>
							<span>{#if game.status === 'ended'}{i + 1}.&nbsp;{/if}{player.username}</span>
							<span class="font-bold">{player.score}</span>
						</a>
					{/each}
				</div>
				<div class="mt-6 flex gap-2">
					{#if game?.status === 'ended' && game?.is_your_game}
						<button class="flex-1 bg-blue-500 text-white py-2 rounded hover:bg-blue-600" onclick={handleRematch}>
							Rejouer
						</button>
					{/if}
					<button class="flex-1 bg-gray-300 py-2 rounded hover:bg-gray-400" onclick={() => showScores.set(false)}>
						Fermer
					</button>
				</div>
			</div>
		</div>
	{/if}
{/if}

<style>
	.no-scrollbar::-webkit-scrollbar { display: none; }
	.no-scrollbar { -ms-overflow-style: none; scrollbar-width: none; }
</style>
