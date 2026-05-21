<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import type { PuzzleDailyLeaderboard } from '$lib/types/puzzle';
	import type { GameInfo } from '$lib/types/game_infos';
	import Board from '$lib/components/Board.svelte';

	interface Props {
		puzzleId: string;
		puzzleBoard: string[][];
	}

	let { puzzleId, puzzleBoard }: Props = $props();

	let loading = $state(true);
	let error = $state<string | null>(null);
	let leaderboard = $state<PuzzleDailyLeaderboard[]>([]);
	let selectedEntry = $state<PuzzleDailyLeaderboard | null>(null);
	let previewGame = $derived(selectedEntry ? toPreviewGame(selectedEntry) : null);

	onMount(async () => {
		try {
			loading = true;
			const res = await api.get(`/puzzles/${puzzleId}/leaderboard?limit=50&offset=0`);
			leaderboard = res.data ?? [];
		} catch (e: any) {
			error = e?.response?.data?.message || 'Erreur lors du chargement du classement';
		} finally {
			loading = false;
		}
	});

	function getMedalEmoji(rank: number): string {
		if (rank === 1) return '🥇';
		if (rank === 2) return '🥈';
		if (rank === 3) return '🥉';
		return '';
	}

	function formatTime(seconds: number): string {
		const mins = Math.floor(seconds / 60);
		const secs = seconds % 60;
		return `${mins}m ${secs.toString().padStart(2, '0')}s`;
	}

	function toPreviewGame(entry: PuzzleDailyLeaderboard): GameInfo {
		const board = puzzleBoard.map((row) => [...row]);
		for (const w of entry.words_played ?? []) {
			const [sx, sy] = w.position.split(',').map((n) => Number.parseInt(n, 10));
			if (!Number.isFinite(sx) || !Number.isFinite(sy)) continue;
			const horizontal = w.direction === 'horizontal';
			for (let i = 0; i < w.word.length; i++) {
				const x = horizontal ? sx + i : sx;
				const y = horizontal ? sy : sy + i;
				if (x < 0 || x >= 15 || y < 0 || y >= 15) break;
				const ch = w.word[i]?.toUpperCase?.() ?? '';
				if (!ch) continue;
				if (!board[y][x] || board[y][x] === ch) board[y][x] = ch;
			}
		}

		return {
			id: puzzleId,
			name: `Tentative de ${entry.username}`,
			board,
			your_rack: '',
			players: [],
			moves: [],
			current_turn: 0,
			current_turn_username: '',
			status: 'ended',
			remaining_letters: 0,
			is_your_game: false,
			pass_count: 0
		};
	}
</script>

{#if loading}
	<div class="text-center py-6">
		<p class="text-gray-600">Chargement du classement...</p>
	</div>
{:else if error}
	<div class="rounded-lg bg-red-50 p-3 ring-1 ring-red-200">
		<p class="text-red-700 text-sm">{error}</p>
	</div>
{:else if leaderboard.length === 0}
	<div class="text-center py-6">
		<p class="text-gray-600">Aucune tentative pour ce puzzle</p>
	</div>
{:else}
	<div class="overflow-x-auto">
		<table class="w-full text-sm">
			<thead>
				<tr class="border-b border-gray-300">
					<th class="text-left py-2 px-3 font-bold text-gray-900">Rang</th>
					<th class="text-left py-2 px-3 font-bold text-gray-900">Joueur</th>
					<th class="text-right py-2 px-3 font-bold text-gray-900">Score</th>
					<th class="text-right py-2 px-3 font-bold text-gray-900">Temps</th>
					<th class="text-right py-2 px-3 font-bold text-gray-900">Grille</th>
				</tr>
			</thead>
			<tbody>
				{#each leaderboard as entry, idx}
					<tr class={idx % 2 === 0 ? 'bg-white' : 'bg-gray-50'}>
						<td class="py-3 px-3 font-bold text-gray-900">
							{getMedalEmoji(entry.rank)}
							#{entry.rank}
						</td>
						<td class="py-3 px-3 text-gray-900">{entry.username}</td>
						<td class="py-3 px-3 text-right font-bold text-emerald-600">{entry.score}</td>
						<td class="py-3 px-3 text-right text-gray-600 text-xs">{formatTime(entry.time_used)}</td>
						<td class="py-3 px-3 text-right">
							{#if entry.words_played && entry.words_played.length > 0}
								<button
									type="button"
									class="px-2 py-1 rounded bg-emerald-100 text-emerald-700 text-xs font-medium hover:bg-emerald-200"
									onclick={() => (selectedEntry = entry)}
								>
									Voir
								</button>
							{:else}
								<span class="text-xs text-gray-400">-</span>
							{/if}
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>

	{#if selectedEntry && previewGame}
		<div class="mt-4 rounded-lg bg-white p-3 ring-1 ring-gray-200">
			<div class="flex items-center justify-between mb-2">
				<p class="text-sm font-semibold text-gray-900">Grille de {selectedEntry.username}</p>
				<button
					type="button"
					class="text-xs px-2 py-1 rounded bg-gray-100 text-gray-700 hover:bg-gray-200"
					onclick={() => (selectedEntry = null)}
				>
					Fermer
				</button>
			</div>
			<div class="mx-auto w-full max-w-[min(95vw,520px)]">
				<div class="mx-auto rounded-sm ring-1 ring-black/5 bg-white shadow p-2" style="width: min(95vw, 100%); height: min(95vw, 100%);">
					<Board
						game={previewGame}
						onPlaceLetter={() => {}}
						onTakeFromBoard={() => {}}
					/>
				</div>
			</div>
		</div>
	{/if}
{/if}
