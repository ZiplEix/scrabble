<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import { page } from '$app/stores';
	import { get, derived } from 'svelte/store';
	import Board from '$lib/components/Board.svelte';
	import { pendingMove, selectedLetter } from '$lib/stores/pendingMove';
  	import { computeWordValue } from '$lib/lettres_value';

	let gameId = '';

	let game = $state<GameInfo | null>(null);
	let error = $state('');
	let loading = $state(true);
	let originalRack: string[] = [];
	let showScores = $state(false);

	let moveScore = derived(pendingMove, (moves) => {
		return computeWordValue(moves);
	});

	onMount(async () => {
		gameId = $page.params.id;
		if (!gameId) return;

		try {
			const res = await api.get(`/game/${gameId}`);
			game = res.data;
			originalRack = [...game!.your_rack];
		} catch (e: any) {
			error = e?.response?.data?.error || 'Erreur lors du chargement de la partie';
		} finally {
			loading = false;
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
	const visibleRack = derived(placedLetters, (used) => {
		const usedCopy = [...used];
		return originalRack.filter((l) => {
			const idx = usedCopy.indexOf(l);
			if (idx !== -1) {
				usedCopy.splice(idx, 1);
				return false;
			}
			return true;
		});
	});

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

		let word = ""
		for (let i = 0; i < sorted.length; i++) {
			word += sorted[i].letter.toUpperCase();
		}

		const body = {
			word,
			x: startX,
			y: startY,
			dir: direction,
			letters: move.map((m) => ({
				x: m.x,
				y: m.y,
				char: m.letter.toUpperCase()
			})),
			score: get(moveScore)
		};

		try {
			await api.post(`/game/${gameId}/play`, body);

			const res = await api.get(`/game/${gameId}`);
			game = res.data;
			originalRack = [...game!.your_rack];
			pendingMove.set([]);
			selectedLetter.set(null);
		} catch (e: any) {
			const msg = e?.response?.data?.message || 'Erreur lors de la validation du coup.';
			alert(msg);
		}
	}
</script>

{#if loading}
	<p class="text-center mt-8">Chargement...</p>
{:else if error}
	<p class="text-center text-red-600 mt-8">{error}</p>
{:else if game}
	<!-- Header: nom + tour + bouton classement -->
	<div class="px-4 pt-3 pb-1 w-full flex justify-between items-center">
		<div>
			<h2 class="text-lg font-semibold text-gray-800">{game.name}</h2>
			<p class="text-sm text-gray-600">Tour de : <strong>{game.current_turn_username}</strong></p>
		</div>
		<!-- Actions: classement + report -->
		<div class="flex flex-col items-end gap-2">
			<button
				class="text-xs bg-gray-200 px-3 py-1 rounded shadow hover:bg-gray-300"
				onclick={() => showScores = true}
			>
				Classement 🏆
			</button>
			<a
				href="/report"
				class="text-xs bg-gray-200 px-3 py-1 rounded shadow hover:bg-gray-300 text-center"
			>
				🛠️ Suggestions / raporter un bug
			</a>
		</div>
	</div>

	<!-- Plateau + rack -->
	<div class="flex flex-col items-center justify-center w-full gap-2 px-2"
    	style="min-height: calc(100vh - 220px);">
		<!-- Plateau -->
		<div class="max-w-[95vw] w-full aspect-square">
			<Board
				game={game}
				{onPlaceLetter}
			/>
		</div>

		<!-- Rack -->
		<div class="flex justify-center gap-1 mt-2 flex-wrap max-w-[95vw]">
			{#each $visibleRack as letter}
				<!-- svelte-ignore a11y_click_events_have_key_events -->
				<div
					role="button"
					tabindex="0"
					class="w-10 h-10 rounded shadow text-center text-lg font-bold flex items-center justify-center border cursor-pointer
						{ $selectedLetter === letter ? 'bg-yellow-400 border-yellow-600' : 'bg-yellow-100 border-yellow-400' }"
					onclick={() => onSelectLetter(letter)}
				>
					{letter}
				</div>
			{/each}
		</div>
	</div>

	<!-- Actions sticky -->
	<div class="fixed bottom-0 left-0 w-full bg-white border-t shadow-inner px-4 py-3 flex justify-between items-center">
		<span class="text-sm font-medium">Score : <strong>{$moveScore}</strong></span>
		<div class="flex gap-3">
			<button class="bg-green-600 text-white px-4 py-2 rounded shadow" onclick={playMove}>
				Valider
			</button>
			<button class="bg-gray-400 text-white px-4 py-2 rounded shadow" onclick={() => {
				pendingMove.set([]); selectedLetter.set(null);
			}}>
				Annuler
			</button>
		</div>
	</div>

	<!-- Modal classement -->
	{#if showScores}
		<div class="fixed inset-0 bg-black/30 backdrop-blur-sm flex items-center justify-center z-50">
			<div class="bg-white rounded-lg shadow-lg w-[90vw] max-w-sm p-6">
				<h3 class="text-lg font-semibold mb-4 text-center">Classement</h3>
				<div class="flex flex-col gap-2">
					{#each game.players as player}
						<div class="flex justify-between items-center p-2 rounded border
							{player.id === game.current_turn ? 'bg-green-100 border-green-400' : 'bg-gray-50'}">
							<span>{player.username}</span>
							<span class="font-bold">{player.score}</span>
						</div>
					{/each}
				</div>
				<button class="mt-6 w-full bg-gray-300 py-2 rounded hover:bg-gray-400" onclick={() => showScores = false}>
					Fermer
				</button>
			</div>
		</div>
	{/if}
{/if}
