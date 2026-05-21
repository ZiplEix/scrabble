<script lang="ts">
	import { onMount } from 'svelte';
	import HeaderBar from '$lib/components/HeaderBar.svelte';
	import { api } from '$lib/api';
	import type { PuzzleHistory } from '$lib/types/puzzle';

	let loading = $state(true);
	let error = $state<string | null>(null);
	let history = $state<PuzzleHistory[]>([]);

	onMount(async () => {
		await loadHistory();
	});

	async function loadHistory() {
		try {
			loading = true;
			error = null;
			const res = await api.get('/puzzles?limit=50&offset=0');
			history = res.data ?? [];
		} catch (e: any) {
			error = e?.response?.data?.message || 'Erreur lors du chargement de l\'historique';
		} finally {
			loading = false;
		}
	}

	function getLevelLabel(level: number): string {
		const labels: Record<number, string> = {
			0: 'Infini',
			1: 'Facile',
			2: 'Moyen',
			3: 'Difficile'
		};
		return labels[level] || 'Inconnu';
	}

	function getTimeoutText(level: number): string {
		const timeouts: Record<number, string> = {
			1: '3 min',
			2: '5 min',
			3: '7 min'
		};
		return timeouts[level] || '?';
	}
</script>

<HeaderBar title="Historique des puzzles" back={true} />

<main class="max-w-2xl mx-auto px-4 py-6">
	{#if loading}
		<div class="text-center py-8">
			<p class="text-gray-600">Chargement...</p>
		</div>
	{:else if error}
		<div class="rounded-lg bg-red-50 p-4 ring-1 ring-red-200 mb-4">
			<p class="text-red-700 font-medium">{error}</p>
		</div>
	{:else if history.length === 0}
		<div class="rounded-lg bg-amber-50 p-6 ring-1 ring-amber-200 text-center">
			<p class="text-amber-700 font-medium">Aucun puzzle complété pour l'instant</p>
			<a href="/puzzles" class="inline-flex mt-4 px-4 py-2 bg-emerald-600 text-white rounded-lg font-medium hover:bg-emerald-700">
				Jouer le puzzle du jour
			</a>
		</div>
	{:else}
		<div class="space-y-3">
			{#each history as item (item.id)}
				<div class="rounded-lg border border-gray-200 p-4 hover:shadow-md transition">
					<div class="flex items-start justify-between">
						<div class="flex-1">
							<p class="text-sm text-gray-600">
								{new Date(item.puzzle_date).toLocaleDateString('fr-FR', {
									weekday: 'long',
									year: 'numeric',
									month: 'long',
									day: 'numeric'
								})} • {getLevelLabel(item.level)}
							</p>
							{#if item.has_attempted && item.player_attempt}
								<p class="text-lg font-bold text-emerald-700 mt-1">
									Score: {item.player_attempt.score}
								</p>
								<p class="text-xs text-gray-600 mt-1">
									Temps: {Math.floor(item.player_attempt.time_used_secs / 60)}m {item.player_attempt.time_used_secs % 60}s
									• Rang: #{item.player_attempt.rank_today}
								</p>
							{:else}
								<p class="text-gray-600 italic text-sm mt-1">Non tenté</p>
							{/if}
						</div>
						<div class="ml-4">
							{#if item.has_attempted}
								<a href={`/puzzles/${item.id}`} class="inline-flex px-3 py-1 bg-blue-100 text-blue-700 rounded text-xs font-medium hover:bg-blue-200">
									Voir détails
								</a>
							{:else}
								<a href={`/puzzles/${item.id}`} class="inline-flex px-3 py-1 bg-emerald-100 text-emerald-700 rounded text-xs font-medium hover:bg-emerald-200">
									Jouer
								</a>
							{/if}
						</div>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</main>
