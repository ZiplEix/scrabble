<script lang="ts">
	import { onMount } from 'svelte';
	import HeaderBar from '$lib/components/HeaderBar.svelte';
	import { api } from '$lib/api';
	import PuzzleLeaderboard from '$lib/components/PuzzleLeaderboard.svelte';
	import type { PuzzleInfo } from '$lib/types/puzzle';
	import { page } from '$app/stores';

	let loading = $state(true);
	let error = $state<string | null>(null);
	let puzzle = $state<PuzzleInfo | null>(null);

	onMount(async () => {
		const puzzleId = $page.params.id;
		try {
			loading = true;
			const res = await api.get(`/puzzles/${puzzleId}`);
			puzzle = res.data;
		} catch (e: any) {
			error = e?.response?.data?.message || 'Erreur lors du chargement du puzzle';
		} finally {
			loading = false;
		}
	});

	function getLevelLabel(level: number): string {
		const labels: Record<number, string> = { 0: 'Infini', 1: 'Facile', 2: 'Moyen', 3: 'Difficile' };
		return labels[level] || 'Inconnu';
	}

	function normalizeBoard(raw: any): string[][] {
		if (!Array.isArray(raw)) {
			return Array.from({ length: 15 }, () => Array.from({ length: 15 }, () => ''));
		}
		return raw.map((row: any) => {
			if (!Array.isArray(row)) return Array.from({ length: 15 }, () => '');
			return row.map((cell: any) => (typeof cell === 'string' ? cell : ''));
		});
	}
</script>

{#snippet historyButton()}
	<a
		href="/puzzles/history"
		class="inline-flex items-center gap-2 rounded-full bg-white px-3 py-1.5 text-sm font-medium text-gray-700 ring-1 ring-black/5 shadow-sm hover:bg-gray-50"
		aria-label="Voir l'historique des puzzles"
	>
		<svg width="16" height="16" viewBox="0 0 24 24" fill="none" aria-hidden="true">
			<path d="M12 8v5l3 2" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
			<path d="M3.05 11A9 9 0 1 1 6 17.3" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
			<path d="M3 4v7h7" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
		</svg>
		<span>Historique</span>
	</a>
{/snippet}

<HeaderBar title="Détail du puzzle" back={true} right={historyButton} />

<main class="max-w-2xl mx-auto px-4 py-6">
	{#if loading}
		<div class="text-center py-8">
			<p class="text-gray-600">Chargement...</p>
		</div>
	{:else if error}
		<div class="rounded-lg bg-red-50 p-4 ring-1 ring-red-200 mb-4">
			<p class="text-red-700 font-medium">{error}</p>
		</div>
	{:else if puzzle}
		<div class="mb-6">
			<h2 class="text-2xl font-bold text-gray-900">Puzzle {puzzle.puzzle_date}</h2>
			<p class="text-gray-600 mt-1">{getLevelLabel(puzzle.level)} • Timeout: {puzzle.timeout_seconds}s</p>
		</div>

		{#if puzzle.has_player_attempted}
			<div class="rounded-lg bg-emerald-50 p-4 ring-1 ring-emerald-200 mb-6">
				<p class="text-emerald-700 font-medium">✓ Vous avez complété ce puzzle</p>
			</div>
		{:else}
			<div class="rounded-lg bg-amber-50 p-4 ring-1 ring-amber-200 mb-6">
				<p class="text-amber-700 font-medium">Ce puzzle n'a pas encore été tenté par vous</p>
			</div>
		{/if}

		<div class="rounded-lg border border-gray-200 p-6 bg-gray-50">
			<h3 class="text-lg font-bold text-gray-900 mb-4">Classement du jour</h3>
			<PuzzleLeaderboard puzzleId={puzzle.id} puzzleBoard={normalizeBoard(puzzle.board)} />
		</div>
	{/if}
</main>
