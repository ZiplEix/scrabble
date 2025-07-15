<script lang="ts">
	import { onMount } from 'svelte';
	import { user } from '$lib/stores/user';
	import { api } from '$lib/api';
	import GameList from '$lib/components/GameList.svelte';
	import { goto } from '$app/navigation';

	let games: {
		id: string;
		name: string;
		current_turn_username: string;
		last_play_time: string;
		is_your_game: boolean;
	}[] = [];

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

	async function deleteGame(id: string) {
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

	async function renameGame({ id, currentName }: { id: string; currentName: string }) {
		const newName = await renameGamePrompt(currentName);
		if (!newName) return;

		try {
			await api.put(`/game/${id}`, { name: newName });
			// Met à jour localement
			games = games.map(g => (g.id === id ? { ...g, name: newName } : g));
		} catch (err) {
			alert('Erreur lors du renommage de la partie');
		}
	}
</script>

<h1 class="text-2xl font-bold mb-4">Bienvenue sur Scrabble en ligne</h1>

{#if !$user}
	<p class="text-gray-700">
		<a href="/login" class="text-blue-600 underline">Connexion</a> ou
		<a href="/register" class="text-blue-600 underline">Inscription</a> pour commencer à jouer !
	</p>
{:else}
	<div class="mb-6">
		<h2 class="text-xl font-semibold mb-2">Mes parties en cours</h2>
		<GameList
			{games}
			on:delete={(e) => deleteGame(e.detail.id)}
			on:rename={(e) => renameGame(e.detail)}
		/>
	</div>

	<button on:click={createGame} class="bg-green-600 text-white px-4 py-2 rounded">
		Créer une nouvelle partie
	</button>
{/if}
