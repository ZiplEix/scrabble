<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import { page } from '$app/stores';
	import { get, derived } from 'svelte/store';
	import Board from '$lib/components/Board.svelte';
	import { pendingMove, selectedLetter } from '$lib/stores/pendingMove';

	let gameId = '';

	let game = $state<GameInfo | null>(null);
	let error = $state('');
	let loading = $state(true);
	let originalRack: string[] = [];

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
</script>

{#if loading}
	<p class="text-center mt-8">Chargement...</p>
{:else if error}
	<p class="text-center text-red-600 mt-8">{error}</p>
{:else if game}
	<div class="flex flex-col items-center gap-4 mt-6 px-2">
		<h2 class="text-xl font-semibold">{game.name}</h2>
		<p>Tour de : <strong>{game.current_turn_username}</strong></p>

		<Board game={game} on:placeLetter={(e) => onPlaceLetter(e.detail.x, e.detail.y, e.detail.cell)} />

		<!-- Rack -->
		<div class="flex justify-center gap-2 mt-6 flex-wrap max-w-[95vw]">
			{#each $visibleRack as letter}
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

		<!-- Actions -->
		<div class="flex gap-4 mt-6">
			<button
				class="px-4 py-2 bg-green-600 text-white rounded shadow"
				onclick={() => {
					console.log('Coup à valider :', get(pendingMove));
					// Appeler API ici
				}}
			>
				Valider le coup
			</button>

			<button
				class="px-4 py-2 bg-gray-400 text-white rounded shadow"
				onclick={() => {
					pendingMove.set([]);
					selectedLetter.set(null);
				}}
			>
				Annuler
			</button>
		</div>
	</div>
{/if}
