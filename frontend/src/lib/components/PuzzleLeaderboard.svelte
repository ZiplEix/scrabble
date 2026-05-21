<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import type { PuzzleDailyLeaderboard } from '$lib/types/puzzle';

	interface Props {
		puzzleId: string;
	}

	let { puzzleId }: Props = $props();

	let loading = $state(true);
	let error = $state<string | null>(null);
	let leaderboard = $state<PuzzleDailyLeaderboard[]>([]);

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
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
{/if}
