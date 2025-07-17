<script lang="ts">
    import { goto } from "$app/navigation";
    import { api } from "$lib/api";
    import GameList from "$lib/components/GameList.svelte";
  	import { user } from "$lib/stores/user";
    import type { GameSummary } from "$lib/types/game_summary";
  	import { onMount } from "svelte";

    let games: GameSummary[] = [];

	onMount(async () => {
		if ($user) {
			try {
				const res = await api.get('/game');
				games = res.data.games;
			} catch (err) {
				console.error('Erreur en récupérant les parties', err);
			}
		}
	});

    function createGame() {
		goto('/games/new');
	}

    async function onDelete(id: string) {
		if (!confirm('Voulez-vous vraiment supprimer cette partie ?')) return;
		try {
			await api.delete(`/game/${id}`);
			games = games.filter(g => g.id !== id);
		} catch (err) {
			alert('Erreur lors de la suppression de la partie');
		}
	}

	function renameGamePrompt(currentName: string): Promise<string | null> {
		const newName = prompt('Entrez le nouveau nom de la partie', currentName);
		return Promise.resolve(newName && newName.trim() !== '' ? newName.trim() : null);
	}

    async function onRename(id: string, currentName: string) {
		const newName = await renameGamePrompt(currentName);
		if (!newName) return;
		try {
			await api.put(`/game/${id}/rename`, { new_name: newName });
			games = games.map(g => (g.id === id ? { ...g, name: newName } : g));
		} catch (err) {
			alert('Erreur lors du renommage de la partie');
		}
	}
</script>

<div class="mb-6">
    <h2 class="text-xl font-semibold mb-2">Mes parties en cours</h2>
    <GameList
        {games}
        {onDelete}
        {onRename}
    />
</div>
<button on:click={createGame} class="bg-green-600 text-white px-4 py-2 rounded hover:bg-green-700">
    Créer une nouvelle partie
</button>