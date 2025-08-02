<script lang="ts">
    import { goto } from "$app/navigation";
    import { api } from "$lib/api";
    import GameList from "$lib/components/GameList.svelte";
  	import { user } from "$lib/stores/user";
    import type { GameSummary } from "$lib/types/game_summary";
  	import { onMount } from "svelte";

    let games: GameSummary[] = $state<GameSummary[]>([]);

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

	let myTurnGames = $derived(games.filter(g => g.current_turn_username === $user?.username && g.status === 'ongoing'));

	let ongoingGame = $derived(games.filter(g => g.status === 'ongoing' && g.current_turn_username !== $user?.username));

	let finishedGames = $derived(games.filter(g => g.status === 'ended'));

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
	<h2 class="text-xl font-semibold mb-2">À mon tour</h2>
	<GameList
        games={myTurnGames}
        {onDelete}
        {onRename}
		placeholder="Aucune partie à jouer pour le moment."
		showTurnOf={true}
		showLastPlayTime={true}
		winning={false}
    />

	<h2 class="text-xl font-semibold mb-2">Mes parties en cours</h2>
    <GameList
        games={ongoingGame}
        {onDelete}
        {onRename}
		placeholder="Aucune partie en cours."
		showTurnOf={true}
		showLastPlayTime={true}
		winning={false}
    />
	<h2 class="text-xl font-semibold mb-2 mt-6">Mes parties terminées</h2>
	<GameList
		games={finishedGames}
		{onDelete}
		placeholder="Aucune partie terminée."
		{onRename}
		winning={true}
		showTurnOf={false}
		showLastPlayTime={false}
	/>
</div>

<button onclick={createGame} class="bg-green-600 text-white px-4 py-2 rounded hover:bg-green-700">
    Créer une nouvelle partie
</button>
