<script lang='ts'>
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { api } from "$lib/api";
  import { gameStore } from "$lib/stores/game";
  import { onMount } from "svelte";
  import { get } from "svelte/store";

    let game: GameInfo | null = $state<GameInfo | null>(null);
	let error = $state('');
	let loading = $state(true);

	onMount(async () => {
		const id = $page.params.id;
		const stored = get(gameStore);
		if (stored?.id === id) {
			game = stored;
			loading = false;
			return;
		}

		try {
			const res = await api.get<GameInfo>(`/game/${id}`);
			game = res.data;
			gameStore.set(res.data);
		} catch (e: any) {
			error = e?.response?.data?.message || 'Impossible de charger l’historique.';
		} finally {
			loading = false;
		}
	})

	function getUsername(pid: number) {
		return game?.players.find(p => p.id === pid)?.username ?? '–';
	}

	function backToGame() {
		if (game) goto(`/games/${game.id}`);
		else goto('/');
	}
</script>

{#if loading}
  	<p class="mt-8 text-center text-gray-600">Chargement de l’historique…</p>
{:else if error}
  	<p class="mt-8 text-center text-red-600">{error}</p>
{:else if game}
	<div class="px-4 pt-6">
		<button
			class="text-sm text-blue-600 hover:underline mb-4 flex items-center"
			onclick={backToGame}
		>
			← Retour à la partie
		</button>
  	</div>

  	<div class="px-4 pt-6 pb-12 space-y-4">
		<h1 class="text-xl font-bold text-gray-800 text-center mb-4">
			Historique de “{game.name}”
    	</h1>

		{#each game.moves as move, idx}
			<article class="bg-white rounded-lg shadow-lg p-4 space-y-2">
				<!-- En-tête : numéro + mot + score -->
				<div class="flex items-center justify-between">
					<div class="flex items-baseline space-x-2">
						<span class="text-sm text-gray-500">#{idx + 1}</span>
						<span class="font-semibold text-gray-800">{move.move.word}</span>
						<span
							class="text-xs uppercase px-2 py-0.5 bg-blue-100 text-blue-800 rounded"
						>
							{move.move.dir}
						</span>
					</div>
					<span class="text-sm font-bold text-green-600">
						+{move.move.score}
					</span>
				</div>

				<!-- Joueur + date -->
				<div class="flex flex-wrap justify-between text-xs text-gray-600">
					<span>Par <strong>{getUsername(move.player_id)}</strong></span>
					<span>
						{new Date(move.played_at).toLocaleTimeString('fr-FR', {
							hour: '2-digit',
							minute: '2-digit'
						})}
					</span>
					<span class="w-full mt-1">
						{new Date(move.played_at).toLocaleDateString('fr-FR')}
					</span>
				</div>

				<!-- Positions des lettres -->
				<div class="flex flex-wrap gap-1 mt-2">
					{#each move.move.letters as l}
						<div
							class="flex items-center justify-center w-8 h-8 bg-gray-100 rounded"
							title={`(${l.x},${l.y})`}
						>
							<span class="text-sm font-mono">{l.char}</span>
						</div>
					{/each}
				</div>
			</article>
		{/each}

		{#if game.moves.length === 0}
			<p class="text-center text-gray-500 mt-8">Aucun coup joué pour l’instant.</p>
		{/if}
  	</div>
{/if}
