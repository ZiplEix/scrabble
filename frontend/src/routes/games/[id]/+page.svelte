<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import { page } from '$app/stores';
	import { get, derived, writable, type Readable } from 'svelte/store';
	import Board from '$lib/components/Board.svelte';
	import { pendingMove } from '$lib/stores/pendingMove';
	import { dndzone } from 'svelte-dnd-action';
	import { letterValues } from '$lib/lettres_value';
	import { flip } from 'svelte/animate';
	import { cubicOut } from 'svelte/easing';
	import { user } from '$lib/stores/user';
  	import { gameStore } from '$lib/stores/game';
  	import GameHeader from '$lib/components/GameHeader.svelte';

	let gameId = $state<string | null>(null);
	let game = $state<GameInfo | null>(null);
	let error = $state('');
	let loading = $state(true);
	type RackLetter = { id: string; char: string };
	let originalRack = writable<RackLetter[]>([]);
	let showScores = $state(writable<boolean>(false));

	let sortedPlayers = $derived(game
		? [...game.players].sort((a, b) => b.score - a.score)
		: []
	)

	let moveScore: Readable<number> = derived(
		[pendingMove, page],
		([$moves, $page], set) => {
			if (!$moves.length || !game) return set(0);

			api.post(`/game/${$page.params.id}/simulate_score`, {
				letters: $moves.map(m => ({
					x: m.x,
					y: m.y,
					char: m.letter.toUpperCase()
				}))
			})
			.then(res => set(res.data.score))
			.catch(() => set(0));
		}
	)

	onMount(async () => {
		gameId = $page.params.id;
		if (!gameId) return;

		try {
			if (gameStore && $gameStore?.id === gameId) {
				game = $gameStore;
				loading = false;
				return;
			}

			const res = await api.get(`/game/${gameId}`);
			game = res.data;
			gameStore.set(game);
			console.log('Game data:', $state.snapshot(game));
			originalRack.set(game!.your_rack.split('').map((char, i) => ({
				id: `${i}-${char}-${crypto.randomUUID()}`,
				char
			})));
		} catch (e: any) {
			error = e?.response?.data?.error || 'Erreur lors du chargement de la partie';
		} finally {
			loading = false;

			if (game?.status === 'ended') {
				showScores.set(true);
			}
		}
	});

	function onPlaceLetter(x: number, y: number, cell: string) {
		const currentMoves = get(pendingMove);
		const existing = currentMoves.find((m) => m.x === x && m.y === y);

		if (existing) {
			pendingMove.set(currentMoves.filter((m) => !(m.x === x && m.y === y)));
			return;
		}

		if (cell) return;
		// placement now performed by drag-drop only; keep function for programmatic use
		return;
	}

	function cancelPendingMove() {
		const moves = get(pendingMove);
		if (!moves || moves.length === 0) {
			pendingMove.set([]);
			return;
		}

		// restore used rack letters (if any) back to originalRack
		originalRack.update(r => {
			const existingIds = new Set(r.map(i => i.id));
			const toAdd = moves.map((m) => {
				if (m.rackId) return { id: m.rackId, char: m.letter };
				return { id: `${Date.now()}-${m.letter}-${crypto.randomUUID()}`, char: m.letter };
			}).filter(item => !existingIds.has(item.id));
			return [...r, ...toAdd];
		});

		pendingMove.set([]);
	}

	function takeBackFromBoard(x: number, y: number) {
		const moves = get(pendingMove);
		const idx = moves.findIndex(m => m.x === x && m.y === y);
		if (idx === -1) return;
		const move = moves[idx];
		// restore to originalRack preserving id when available
		originalRack.update(r => {
			if (move.rackId && r.some(it => it.id === move.rackId)) return r;
			const item = move.rackId ? { id: move.rackId, char: move.letter } : { id: `${Date.now()}-${move.letter}-${crypto.randomUUID()}`, char: move.letter };
			return [...r, item];
		});
		// remove from pendingMove
		pendingMove.update(ms => ms.filter((m) => !(m.x === x && m.y === y)));
	}

	const visibleRack = derived(
		[originalRack, pendingMove],
		([$rack, $moves]) => {
			const usedIds = new Set($moves.map(m => m.rackId).filter(Boolean) as string[]);
			return $rack.filter(r => !usedIds.has(r.id));
		}
	);

	let submitting = $state(false);

	async function playMove() {
		submitting = true;
		const move = get(pendingMove);
		if (!move.length) return;

		const sorted = [...move].sort((a, b) => a.x - b.x || a.y - b.y);
		const sameRow = sorted.every((l) => l.y === sorted[0].y);
		const sameCol = sorted.every((l) => l.x === sorted[0].x);

		if (!sameRow && !sameCol) {
			alert('Les lettres doivent être alignées horizontalement ou verticalement.');
			return;
		}

		const direction = sameRow ? "H" : "V";
		const startX = sorted[0].x;
		const startY = sorted[0].y;
		let word = sorted.map(l => l.letter.toUpperCase()).join("");

		const body = {
			word,
			x: startX,
			y: startY,
			dir: direction,
			letters: move.map((m) => ({ x: m.x, y: m.y, char: m.letter.toUpperCase() })),
			score: get(moveScore)
		};

		try {
			await api.post(`/game/${gameId}/play`, body);
			const res = await api.get(`/game/${gameId}`);
			game = res.data;
			gameStore.set(game);
			originalRack.set(game!.your_rack.split('').map((char, i) => ({
				id: `${i}-${char}-${crypto.randomUUID()}`,
				char
			})));
			pendingMove.set([]);
		} catch (e: any) {
			const msg = e?.response?.data?.message || 'Erreur lors de la validation du coup.';
			alert(msg);
		} finally {
			submitting = false;
			if (game?.status === 'ended') {
				showScores.set(true);
			}
		}
	}

	async function drawNewRack() {
		const ok = confirm('Êtes-vous sûr de vouloir changer toutes vos lettres ? Cela remplacera vos lettres actuelles et passera votre tour.');
		if (!ok) return;

		try {
			const res = await api.get(`/game/${gameId}/new_rack`);
			const newRack = res.data as string[];
			console.log('Nouveau rack:', newRack);
			if (newRack.length === 0) {
				alert('Plus de lettres disponibles dans le sac.');
				return;
			}
			originalRack.set(newRack.map((char, i) => ({
				id: `${i}-${char}-${crypto.randomUUID()}`,
				char
			})));
					pendingMove.set([]);
		} catch (e: any) {
			const msg = e?.response?.data?.message || 'Erreur lors du tirage d\'un nouveau rack.';
			console.error(e);
			alert(msg);
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
			const msg = e?.response?.data?.message || 'Erreur lors du passage du tour.';
			alert(msg);
		}
	}

	function shuffleRack() {
		const rack = get(originalRack);
		for (let i = rack.length - 1; i > 0; i--) {
			const j = Math.floor(Math.random() * (i + 1));
			[rack[i], rack[j]] = [rack[j], rack[i]];
		}
		originalRack.set(rack);
	}

	async function handleRematch() {
		// Guard côté client : s'assurer que seul le créateur peut créer une revanche
		if (!game?.is_your_game) {
			alert('Seul le créateur de la partie peut créer une revanche.');
			return;
		}
		const defaultName = `${game!.name} – revanche`;
		const newName = prompt('Nom de la nouvelle partie :', defaultName);
		if (!newName) return;

		const currentUsername = get(user)?.username;
		const opponents = game!.players
			.map(p => p.username)
			.filter(u => u && u !== currentUsername);

		try {
			const res = await api.post('/game', {
				name: newName,
				players: opponents,
			});
			window.location.href = `/games/${res.data.game_id}`;
    	} catch (err: any) {
      		alert(err?.response?.data?.message || 'Impossible de créer la revanche.');
    	}
  	}
</script>

{#if loading}
	<p class="text-center mt-8">Chargement...</p>
{:else if error}
	<p class="text-center text-red-600 mt-8">{error}</p>
{:else if game}
	<!-- Main colonne sans fixed -->
	<div class="flex flex-col overflow-hidden"
		style="height: calc(100dvh - var(--nav-h, 72px));"
	>
		<!-- Header compact avec menu -->
		<GameHeader game={game} showScores={showScores} gameId={gameId!} />

		<!-- Zone centrale qui prend le reste -->
		<main class="flex-1 flex flex-col min-h-0 overflow-hidden">

			<!-- Board : centré verticalement et horizontalement -->
			<div class="flex-1 grid place-items-center px-2 min-h-0">
				<!-- Carré qui s’adapte à la hauteur dispo ET à la largeur -->
				<div class="mx-auto"
					style="width: min(95vw, 100%); height: min(95vw, 100%);"
				>
					<Board {game} {onPlaceLetter} onTakeFromBoard={takeBackFromBoard} onDropFromRack={(char, x, y, id) => {
						console.log('[page] onDropFromRack', { char, x, y, id });
						// called when an item from the rack is dropped on the board via svelte-dnd-action
						const cell = game!.board[y][x];
						if (cell) {
							// target occupied -> restore tile into originalRack so it doesn't disappear
							originalRack.update(r => {
								// if we already have the id in the rack, do nothing
								if (id && r.some(it => it.id === id)) return r;
								// re-add with same id when available, otherwise create a new id
								const newItem = id ? { id, char } : { id: `${Date.now()}-${char}-${crypto.randomUUID()}`, char };
								return [...r, newItem];
							});
							return; // do not add to pendingMove
						}
						// add to pendingMove and tag which rack item was used
						pendingMove.update(moves => {
							const filtered = moves.filter((m) => !(m.x === x && m.y === y));
							return [...filtered, { x, y, letter: char, rackId: id }];
						});
						// remove from originalRack by id
						originalRack.update(r => r.filter(item => item.id !== id));
					}} />
				</div>
			</div>

			<!-- Rack (dernier, bord à bord) -->
			<div class="bg-white/95 backdrop-blur supports-backdrop-blur:border-y border-y flex-none mb-4">
				<div class="px-2 pt-2 pb-[max(env(safe-area-inset-bottom),10px)]">
					<div class="mx-auto max-w-[95vw] overflow-x-auto no-scrollbar">
						<div
							class="flex gap-1 whitespace-nowrap justify-center"
							use:dndzone={{
								items: $visibleRack,
								flipDurationMs: 150,
								dropFromOthersDisabled: false,
								dragDisabled: false,
							}}
							onconsider={({ detail }) => originalRack.set(detail.items)}
							onfinalize={({ detail }) => originalRack.set(detail.items)}
						>
							{#each $visibleRack as item (item.id)}
								<!-- svelte-ignore a11y_click_events_have_key_events -->
								<div
									role="button"
									tabindex="0"
									draggable="true"
									ondragstart={(e) => {
										// Try to use dataTransfer but also set a global fallback because some dnd libs intercept dataTransfer
										e.dataTransfer?.setData('text/plain', JSON.stringify({ char: item.char }));
										try { e.dataTransfer!.effectAllowed = 'move'; e.dataTransfer!.dropEffect = 'move'; } catch (err) {}
										(window as any).__draggedTile = { char: item.char, id: item.id };
										(window as any).__dndActive = true;
										// stop propagation so svelte-dnd-action doesn't intercept this native drag
										try { e.stopPropagation(); } catch (err) {}
									}}
									ondragend={() => { try { (window as any).__draggedTile = null; (window as any).__dndActive = false; } catch(e){} }}
									class="relative inline-flex w-11 h-11 rounded-lg text-center text-lg font-bold items-center justify-center border cursor-pointer bg-yellow-100 border-yellow-400"
									animate:flip={{ duration: 200, easing: cubicOut }}
								>
									{item.char}
									<span class="absolute bottom-0.5 right-1 text-[10px] font-normal text-gray-600">
										{letterValues[item.char]}
									</span>
								</div>
							{/each}
						</div>
					</div>
				</div>
			</div>

			<!-- Cluster d’actions (en flux, juste au-dessus du rack) -->
			<div class="px-3 pb-4 flex-none">
				<div class="mx-auto max-w-[640px]">
					<div class="w-full rounded-2xl bg-white/95 backdrop-blur-md shadow-lg ring-1 ring-black/5">
						<div class="grid grid-cols-4 divide-x">
							<!-- Annuler -->
							<button
								class="h-12 px-2 flex flex-col items-center justify-center text-[12px] font-medium disabled:opacity-40 active:scale-[0.98] transition"
								onclick={cancelPendingMove}
								disabled={$moveScore <= 0 || get(pendingMove).length === 0}
								aria-label="Annuler le coup en cours"
							>
								<svg width="18" height="18" viewBox="0 0 24 24" fill="none" aria-hidden="true">
									<path d="M9 16l-4-4 4-4" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
									<path d="M20 20a8 8 0 0 0-8-8H5" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
								</svg>
								<span>Annuler</span>
							</button>

							<!-- Passer -->
							<button
								class="h-12 px-2 flex flex-col items-center justify-center text-[12px] font-medium active:scale-[0.98] transition"
								onclick={shuffleRack}
								aria-label="Passer le tour"
							>
								<svg width="18" height="18" viewBox="0 0 24 24" version="1.1" xml:space="preserve" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink">
									<path d="M8.7,14.2C8,14.7,7.1,15,6.2,15H4c-0.6,0-1,0.4-1,1s0.4,1,1,1h2.2c1.3,0,2.6-0.4,3.7-1.2c0.4-0.3,0.5-1,0.2-1.4    C9.7,13.9,9.1,13.8,8.7,14.2z"/>
									<path d="M13,10.7c0.3,0,0.6-0.1,0.8-0.3C14.5,9.5,15.6,9,16.8,9h0.8l-0.3,0.3c-0.4,0.4-0.4,1,0,1.4c0.2,0.2,0.5,0.3,0.7,0.3    s0.5-0.1,0.7-0.3l2-2c0.1-0.1,0.2-0.2,0.2-0.3c0.1-0.2,0.1-0.5,0-0.8c-0.1-0.1-0.1-0.2-0.2-0.3l-2-2c-0.4-0.4-1-0.4-1.4,0    s-0.4,1,0,1.4L17.6,7h-0.8c-1.8,0-3.4,0.8-4.6,2.1c-0.4,0.4-0.3,1,0.1,1.4C12.5,10.7,12.8,10.7,13,10.7z"/>
									<path d="M20.7,15.3l-2-2c-0.4-0.4-1-0.4-1.4,0s-0.4,1,0,1.4l0.3,0.3h-1.5c-1.6,0-2.9-0.9-3.6-2.3l-1.2-2.4C10.3,8.3,8.2,7,5.9,7H4    C3.4,7,3,7.4,3,8s0.4,1,1,1h1.9c1.6,0,2.9,0.9,3.6,2.3l1.2,2.4c1,2.1,3.1,3.4,5.4,3.4h1.5l-0.3,0.3c-0.4,0.4-0.4,1,0,1.4    c0.2,0.2,0.5,0.3,0.7,0.3s0.5-0.1,0.7-0.3l2-2C21.1,16.3,21.1,15.7,20.7,15.3z"/>
								</svg>
								<span>Mélanger</span>
							</button>

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

							{#if $moveScore <= 0 || get(pendingMove).length === 0}
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
							{:else}
								<!-- Valider (CTA) -->
								<button
									class="relative h-12 px-2 flex flex-col items-center justify-center text-[12px] font-semibold text-white bg-green-600 rounded-r-2xl active:scale-[0.98] transition disabled:opacity-60 disabled:bg-green-600/70"
									onclick={playMove}
									disabled={$moveScore <= 0 || get(pendingMove).length === 0 || submitting}
									aria-label="Valider le coup"
								>
									{#if submitting}
										<!-- spinner -->
										<svg class="animate-spin" width="18" height="18" viewBox="0 0 24 24" fill="none" aria-hidden="true">
											<circle cx="12" cy="12" r="10" stroke="rgba(255,255,255,0.4)" stroke-width="4"></circle>
											<path d="M22 12a10 10 0 0 1-10 10" stroke="white" stroke-width="4" stroke-linecap="round"></path>
										</svg>
									{:else}
										<svg width="18" height="18" viewBox="0 0 24 24" fill="none" aria-hidden="true">
											<path d="M20 7l-9 9-4-4" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
										</svg>
									{/if}
									<span class="mt-1">{submitting ? 'Envoi...' : 'Valider'}</span>

									<!-- Badge score -->
									<span class="absolute -top-2 -right-2 text-[10px] px-2 py-0.5 rounded-full bg-white text-green-700 shadow ring-1 ring-black/5">
										{$moveScore}
									</span>
								</button>
							{/if}
						</div>
					</div>
				</div>
			</div>
		</main>
	</div>

	<!-- Modal classement -->
	{#if $showScores}
		<div class="fixed inset-0 bg-black/30 backdrop-blur-sm flex items-center justify-center z-50">
			<div class="bg-white rounded-lg shadow-lg w-[90vw] max-w-sm p-6">
				{#if game?.status === 'ended'}
					<h3 class="text-lg font-semibold mb-2 text-center">
						🎉 Partie terminée
						<br />
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
					{@const playerClass = game.status === "ended" ? "bg-gray-50" : player.id === game.current_turn ? "bg-green-100 border-green-400" : "bg-gray-50"}
						<div class="flex justify-between items-center p-2 rounded border
							{playerClass}">
							<span>
								{#if game.status === 'ended'}{i+1}.&nbsp;{/if}
								{player.username}
							</span>
							<span class="font-bold">{player.score}</span>
						</div>
					{/each}
				</div>
				<div class="mt-6 flex gap-2">
					{#if game?.status === 'ended' && game?.is_your_game}
						<button
							class="flex-1 bg-blue-500 text-white py-2 rounded hover:bg-blue-600"
							onclick={handleRematch}
						>
				 			Rejouer
				 		</button>
         			{/if}
         			<button
           				class="flex-1 bg-gray-300 py-2 rounded hover:bg-gray-400"
           				onclick={() => showScores.set(false)}
         			>
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
