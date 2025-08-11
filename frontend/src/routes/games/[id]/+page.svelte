<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import { page } from '$app/stores';
	import { get, derived, writable, type Readable } from 'svelte/store';
	import Board from '$lib/components/Board.svelte';
	import { pendingMove, selectedLetter } from '$lib/stores/pendingMove';
	import { letterValues } from '$lib/lettres_value';
	import { dndzone } from 'svelte-dnd-action';
	import { flip } from 'svelte/animate';
	import { cubicOut } from 'svelte/easing';
	import { goto } from '$app/navigation';
	import { user } from '$lib/stores/user';
  	import { gameStore } from '$lib/stores/game';
  	import GameMenu from '$lib/components/GameMenu.svelte';

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

	function onSelectLetter(letter: string) {
		const current = get(selectedLetter);
		selectedLetter.set(current === letter ? null : letter);
	}

	function onPlaceLetter(x: number, y: number, cell: string) {
		const currentMoves = get(pendingMove);
		const existing = currentMoves.find((m) => m.x === x && m.y === y);

		if (existing) {
			pendingMove.set(currentMoves.filter((m) => !(m.x === x && m.y === y)));
			return;
		}

		if (cell) return;
		const letter = get(selectedLetter);
		if (!letter) return;

		pendingMove.update((moves) => {
			const filtered = moves.filter((m) => !(m.x === x && m.y === y));
			return [...filtered, { x, y, letter }];
		});
		selectedLetter.set(null);
	}

	const placedLetters = derived(pendingMove, (moves) => moves.map((m) => m.letter));
	const visibleRack = derived(
		[originalRack, placedLetters],
		([$rack, $used]) => {
			const usedCopy = [...$used];
			return $rack.filter((l) => {
				const idx = usedCopy.indexOf(l.char);
				if (idx !== -1) {
					usedCopy.splice(idx, 1);
					return false;
				}
				return true;
			});
		}
	);

	async function playMove() {
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
			selectedLetter.set(null);
		} catch (e: any) {
			const msg = e?.response?.data?.message || 'Erreur lors de la validation du coup.';
			alert(msg);
		} finally {
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
			selectedLetter.set(null);
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

	async function handleRematch() {
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
      		goto(`/games/${res.data.game_id}`);
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
	<!-- Header compact avec menu -->
	<GameMenu game={game} showScores={showScores} gameId={gameId!} />

	<!-- Board + actions -->
	<div class="relative flex-1 w-full px-2 pb-[max(env(safe-area-inset-bottom),72px)]">
		<div class="mx-auto max-w-[95vw] w-full aspect-square">
			<Board {game} {onPlaceLetter} />
		</div>
	</div>

	<!-- Action cluster -->
	<div
		class="fixed left-0 right-0 z-30 px-3"
		style="bottom: calc(env(safe-area-inset-bottom) + 92px)"
	>
		<div class="mx-auto max-w-[640px]">
			<div class="w-full rounded-2xl bg-white/95 backdrop-blur-md shadow-lg ring-1 ring-black/5">
				<div class="grid grid-cols-4 divide-x">
					<!-- Annuler -->
					<button
						class="h-12 px-2 flex flex-col items-center justify-center text-[12px] font-medium disabled:opacity-40 active:scale-[0.98] transition"
						onclick={() => { pendingMove.set([]); selectedLetter.set(null); }}
						disabled={$moveScore <= 0 || get(pendingMove).length === 0}
						aria-label="Annuler le coup en cours"
					>
						<svg width="18" height="18" viewBox="0 0 24 24" fill="none" aria-hidden="true">
							<path d="M9 14l-4-4 4-4" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
							<path d="M20 20a8 8 0 0 0-8-8H5" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
						</svg>
						<span>Annuler</span>
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

					<!-- Valider (CTA) -->
					<button
						class="relative h-12 px-2 flex flex-col items-center justify-center text-[12px] font-semibold text-white bg-green-600 rounded-r-2xl active:scale-[0.98] transition disabled:opacity-60 disabled:bg-green-600/70"
						onclick={playMove}
						disabled={$moveScore <= 0 || get(pendingMove).length === 0}
						aria-label="Valider le coup"
					>
						<svg width="18" height="18" viewBox="0 0 24 24" fill="none" aria-hidden="true">
							<path d="M20 7l-9 9-4-4" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
						</svg>
						<span>Valider</span>

						<!-- Badge score -->
						<span class="absolute -top-2 -right-2 text-[10px] px-2 py-0.5 rounded-full bg-white text-green-700 shadow ring-1 ring-black/5">
							{$moveScore}
						</span>
					</button>
				</div>
			</div>
		</div>
	</div>

	<!-- Rack dock -->
	<div class="fixed left-0 right-0 bottom-0 z-40 bg-white/95 backdrop-blur supports-backdrop-blur:border-t border-t shadow-inner">
		<div class="px-2 pt-2 pb-[max(env(safe-area-inset-bottom),10px)]">
			<div class="mx-auto max-w-[95vw] overflow-x-auto no-scrollbar">
				<div
					class="flex gap-1 whitespace-nowrap justify-center"
					use:dndzone={{
						items: $visibleRack,
						flipDurationMs: 150,
						dropFromOthersDisabled: true,
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
							class="relative inline-flex w-11 h-11 rounded-lg shadow text-center text-lg font-bold items-center justify-center border cursor-pointer
							{ $selectedLetter === item.char ? 'bg-yellow-400 border-yellow-600' : 'bg-yellow-100 border-yellow-400' }"
							onclick={() => onSelectLetter(item.char)}
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
					{#if game?.status === 'ended'}
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
