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
	import { downloadGameResultImage } from '$lib/utils/canvas_share';

	// Import modular subviews
	import GameChat from '$lib/components/GameChat.svelte';
	import GameHistoryView from '$lib/components/GameHistoryView.svelte';
	import GameDictionaryView from '$lib/components/GameDictionaryView.svelte';

	let gameId = $state<string | null>(null);
	let game = $state<GameInfo | null>(null);
	let error = $state('');
	let loading = $state(true);
	let showScores = $state(writable<boolean>(false));

	// Tab states
	let activeTab = $state<'board' | 'chat' | 'history' | 'dictionary'>('board');
	let hasNewMessages = $state(false);

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
		return `mailto:${adminEmail}?subject=${subject}?body=${body}`;
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
			<div class="mx-auto rounded-3xl bg-white/95 backdrop-blur-md ring-1 ring-black/5 shadow-xl p-6 text-center border border-white/60">
				<div class="mx-auto w-12 h-12 rounded-full bg-amber-100 flex items-center justify-center text-amber-700 mb-3 text-xl">
					⚠️
				</div>
				<h3 class="text-lg font-bold text-stone-800">Impossible de charger la partie</h3>
				<p class="mt-2 text-xs text-stone-600 leading-relaxed">{error}</p>
				<p class="mt-1 text-xs text-stone-500">Si le problème persiste, contactez l'administrateur.</p>
				<div class="mt-5 flex gap-2 justify-center">
					<button class="px-4 py-2 bg-brand-emerald text-white text-xs font-bold rounded-xl shadow-md hover:bg-brand-emerald-hover active:scale-95 transition-all cursor-pointer" onclick={retryLoad}>Réessayer</button>
					<a href="/" class="px-4 py-2 bg-white ring-1 ring-stone-200 text-stone-700 text-xs font-bold rounded-xl hover:bg-stone-50 active:scale-95 transition-all">Accueil</a>
					<a href={buildAdminMailto()} class="px-4 py-2 bg-white ring-1 ring-stone-200 text-stone-700 text-xs font-bold rounded-xl hover:bg-stone-50 active:scale-95 transition-all">Signaler</a>
				</div>
			</div>
		</div>
	</div>
{:else if game}
	<div class="flex flex-col overflow-hidden bg-radial from-brand-cream to-[#f4eee2]" style="height: 100dvh;">
		
		<!-- Custom Header -->
		<GameHeader {game} {showScores} gameId={gameId!} />

		<!-- Unified Sub-navigation Tabs -->
		<div class="mx-4 my-2 max-w-md sm:mx-auto z-10 shrink-0 select-none w-[calc(100%-2rem)] sm:w-full">
			<div class="glass-card p-1 rounded-2xl shadow-md">
				<div class="grid grid-cols-4 gap-1">
					<button 
						onclick={() => activeTab = 'board'} 
						class="py-1.5 text-[10px] font-bold rounded-xl text-center transition-all active:scale-[0.98] {activeTab === 'board' ? 'bg-white text-brand-emerald shadow-sm' : 'text-stone-500 hover:text-stone-700'}"
					>
						Plateau
					</button>
					<button 
						onclick={() => { activeTab = 'chat'; hasNewMessages = false; }} 
						class="py-1.5 text-[10px] font-bold rounded-xl text-center transition-all active:scale-[0.98] relative {activeTab === 'chat' ? 'bg-white text-brand-emerald shadow-sm' : 'text-stone-500 hover:text-stone-700'}"
					>
						Chat
						{#if hasNewMessages}
							<span class="absolute top-1.5 right-3 w-1.5 h-1.5 bg-brand-gold rounded-full gold-glow animate-pulse"></span>
						{/if}
					</button>
					<button 
						onclick={() => activeTab = 'history'} 
						class="py-1.5 text-[10px] font-bold rounded-xl text-center transition-all active:scale-[0.98] {activeTab === 'history' ? 'bg-white text-brand-emerald shadow-sm' : 'text-stone-500 hover:text-stone-700'}"
					>
						Historique
					</button>
					<button 
						onclick={() => activeTab = 'dictionary'} 
						class="py-1.5 text-[10px] font-bold rounded-xl text-center transition-all active:scale-[0.98] {activeTab === 'dictionary' ? 'bg-white text-brand-emerald shadow-sm' : 'text-stone-500 hover:text-stone-700'}"
					>
						Dictionnaire
					</button>
				</div>
			</div>
		</div>

		<!-- Container View -->
		<main class="flex-1 flex flex-col min-h-0 overflow-hidden">
			{#if activeTab === 'board'}
				<!-- Render Game Board -->
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
							class="h-12 px-2 flex flex-col items-center justify-center text-[11px] font-bold text-stone-600 active:scale-[0.95] transition-all cursor-pointer"
							onclick={drawNewRack}
							aria-label="Échanger les lettres"
						>
							<svg class="w-4.5 h-4.5 mb-0.5 text-stone-500" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" d="M19.5 12c0-1.232-.046-2.453-.138-3.662a4.006 4.006 0 00-3.7-3.7 48.656 48.656 0 00-7.324 0 4.006 4.006 0 00-3.7 3.7C4.547 9.547 4.5 10.768 4.5 12s.047 2.453.138 3.662a4.006 4.006 0 003.7 3.7 48.656 48.656 0 007.324 0 4.006 4.006 0 003.7-3.7c.092-1.209.138-2.43.138-3.662z"/>
								<path stroke-linecap="round" stroke-linejoin="round" d="M9 10.5l3 3 3-3"/>
							</svg>
							<span>Échanger</span>
						</button>
						<!-- Passer -->
						<button
							class="h-12 px-2 flex flex-col items-center justify-center text-[11px] font-bold text-stone-600 active:scale-[0.95] transition-all cursor-pointer"
							onclick={passTurn}
							aria-label="Passer le tour"
						>
							<svg class="w-4.5 h-4.5 mb-0.5 text-stone-500" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" d="M15.75 5.25v13.5m-7.5-13.5v13.5"/>
							</svg>
							<span>Passer</span>
						</button>
					{/snippet}
				</GameBoard>
			{:else if activeTab === 'chat'}
				<!-- Render In-Game Chat -->
				<GameChat 
					gameId={gameId!} 
					{game} 
					onNewMessage={() => { if (activeTab !== 'chat') hasNewMessages = true; }} 
				/>
			{:else if activeTab === 'history'}
				<!-- Render Log timeline -->
				<GameHistoryView {game} />
			{:else if activeTab === 'dictionary'}
				<!-- Render Larousse meanings -->
				<GameDictionaryView {game} />
			{/if}
		</main>
	</div>

	<!-- Modal classement -->
	{#if $showScores}
		<div class="fixed inset-0 bg-black/40 backdrop-blur-sm flex items-center justify-center z-50 p-4">
			<div class="bg-white rounded-3xl shadow-xl w-full max-w-sm p-6 border border-stone-200/50">
				{#if game?.status === 'ended'}
					<h3 class="text-xl font-extrabold mb-2 text-center text-stone-800">
						🎉 Partie terminée<br />
						Vainqueur : <span class="text-brand-emerald font-black">{game.winner_username}</span>
					</h3>
					<p class="text-center text-xs text-stone-500 mb-5">
						Terminée le {new Date(game.ended_at!).toLocaleString('fr-FR', { dateStyle: 'short', timeStyle: 'short' })}
					</p>
				{:else}
					<h3 class="text-xl font-extrabold mb-5 text-center text-stone-800">Classement actuel</h3>
				{/if}
				
				<div class="flex flex-col gap-2">
					{#each (game.status === 'ended' ? sortedPlayers : game.players) as player, i}
						{@const playerClass = game.status === 'ended' ? 'bg-stone-50 border-stone-200/60' : player.id === game.current_turn ? 'bg-brand-emerald-light/60 border-brand-emerald/20 ring-2 ring-brand-emerald/10' : 'bg-stone-50 border-stone-200/60'}
						<a
							class="flex justify-between items-center p-3 rounded-2xl border font-bold text-xs text-stone-800 {playerClass}"
							href={player.id !== $user?.id ? `/user/${player.id}` : undefined}
						>
							<span class="flex items-center gap-1.5">
								{#if game.status === 'ended'}
									<span class="text-[10px] text-stone-400 font-black">#{i + 1}</span>
								{/if}
								{player.username}
							</span>
							<span class="font-black text-brand-emerald tabular-nums">{player.score} pts</span>
						</a>
					{/each}
				</div>

				<div class="mt-6 flex flex-col gap-2">
					{#if game?.status === 'ended'}
						<button
							class="w-full bg-amber-500 hover:bg-amber-600 text-white py-3 rounded-2xl font-bold shadow-md active:scale-95 transition-all cursor-pointer text-xs flex items-center justify-center gap-2"
							onclick={() => game && downloadGameResultImage(game)}
						>
							<span>📸</span> Partager le plateau (PNG)
						</button>
					{/if}
					<div class="flex gap-2 w-full">
						{#if game?.status === 'ended' && game?.is_your_game}
							<button class="flex-1 bg-brand-emerald text-white py-3 rounded-2xl font-bold hover:bg-brand-emerald-hover shadow-md active:scale-95 transition-all cursor-pointer text-xs" onclick={handleRematch}>
								Rejouer
							</button>
						{/if}
						<button class="flex-1 bg-stone-100 text-stone-600 py-3 rounded-2xl font-bold hover:bg-stone-200 active:scale-95 transition-all cursor-pointer text-xs" onclick={() => showScores.set(false)}>
							Fermer
						</button>
					</div>
				</div>
			</div>
		</div>
	{/if}
{/if}

<style>
	.no-scrollbar::-webkit-scrollbar { display: none; }
	.no-scrollbar { -ms-overflow-style: none; scrollbar-width: none; }
</style>
