<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import type { PuzzleInfo } from '$lib/types/puzzle';

	interface Props {
		puzzle: PuzzleInfo;
		disabled?: boolean;
	}

	let { puzzle, disabled = false }: Props = $props();

	const dispatch = createEventDispatcher<{
		submit: { wordsPlayed: any[] }; // time_used calculé côté serveur
	}>();

	let rack = $state<string[]>(puzzle.available_letters.split(''));
	let board = $state<(string | null)[][]>(initializeBoard());
	let score = $state(0);
	let selectedTiles = $state<{ row: number; col: number }[]>([]);

	function initializeBoard(): (string | null)[][] {
		const grid: (string | null)[][] = [];
		for (let i = 0; i < 15; i++) {
			grid[i] = Array(15).fill(null);
		}
		return grid;
	}

	function placeTile(row: number, col: number, tileIdx: number) {
		if (disabled || board[row][col] !== null) return;

		const tile = rack[tileIdx];
		if (!tile) return;

		board[row][col] = tile;
		rack.splice(tileIdx, 1);
		selectedTiles.push({ row, col });

		// Recalculate score
		calculateScore();
	}

	function removeTile(row: number, col: number) {
		if (disabled) return;

		if (board[row][col] !== null) {
			rack.push(board[row][col]!);
			board[row][col] = null;

			selectedTiles = selectedTiles.filter((t) => !(t.row === row && t.col === col));

			// Recalculate score
			calculateScore();
		}
	}

	function calculateScore() {
		// Simplified scoring: count unique tiles placed
		score = selectedTiles.length * 10; // Basic scoring
	}

	function handleSubmit() {
		const wordsPlayed = extractPlacedWords();
		// On ne passe plus timeUsed — le serveur le calcule depuis started_at
		dispatch('submit', { wordsPlayed });
	}

	function extractPlacedWords(): any[] {
		// TODO: Implement word extraction logic
		return selectedTiles.map((tile, idx) => ({
			word: 'WORD', // Placeholder
			position: `${tile.row},${tile.col}`,
			direction: 'horizontal'
		}));
	}

	function clearBoard() {
		// Return all tiles to rack
		for (let i = 0; i < 15; i++) {
			for (let j = 0; j < 15; j++) {
				if (board[i][j] !== null) {
					rack.push(board[i][j]!);
					board[i][j] = null;
				}
			}
		}
		selectedTiles = [];
		score = 0;
	}
</script>

<div class="space-y-6">
	<!-- Board Grid -->
	<div class="rounded-lg border border-gray-200 bg-white p-4 overflow-x-auto">
		<div class="inline-block">
			<div class="grid gap-0.5" style="grid-template-columns: repeat(15, 40px);">
				{#each board as row, rowIdx}
					{#each row as cell, colIdx}
						<button
							class="w-10 h-10 border border-gray-300 rounded text-center font-bold text-sm flex items-center justify-center transition
									{cell !== null ? 'bg-emerald-100 text-emerald-900' : 'bg-white hover:bg-gray-50'}
									{disabled ? 'cursor-not-allowed opacity-60' : 'cursor-pointer'}"
							on:click={() => removeTile(rowIdx, colIdx)}
							{disabled}
						>
							{cell || ''}
						</button>
					{/each}
				{/each}
			</div>
		</div>
	</div>

	<!-- Rack (Available Tiles) -->
	<div class="rounded-lg border border-gray-200 bg-gray-50 p-4">
		<p class="text-sm font-bold text-gray-900 mb-3">Pièces disponibles ({rack.length})</p>
		<div class="flex flex-wrap gap-2">
			{#each rack as tile, idx}
				<button
					class="w-12 h-12 rounded-lg bg-amber-400 text-white font-bold text-lg flex items-center justify-center shadow hover:shadow-lg transition {disabled
						? 'opacity-50 cursor-not-allowed'
						: 'cursor-move hover:bg-amber-500'}"
					on:click={() => placeTile(0, 0, idx)}
					draggable={!disabled}
					{disabled}
				>
					{tile}
				</button>
			{/each}
		</div>
	</div>

	<!-- Score & Actions -->
	<div class="rounded-lg border border-gray-200 bg-white p-4">
		<div class="flex items-center justify-between mb-4">
			<div>
				<p class="text-gray-600 text-sm">Score estimé</p>
				<p class="text-3xl font-bold text-emerald-700">{score}</p>
			</div>
			<div class="flex gap-2">
				<button
					class="px-4 py-2 bg-gray-200 text-gray-900 rounded-lg font-medium hover:bg-gray-300 transition {disabled
						? 'opacity-50 cursor-not-allowed'
						: ''}"
					on:click={clearBoard}
					{disabled}
				>
					Réinitialiser
				</button>
				<button
					class="px-4 py-2 bg-emerald-600 text-white rounded-lg font-medium hover:bg-emerald-700 transition {disabled
						? 'opacity-50 cursor-not-allowed'
						: ''}"
					on:click={handleSubmit}
					{disabled}
				>
					Soumettre
				</button>
			</div>
		</div>
	</div>
</div>
