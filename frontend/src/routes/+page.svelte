<script lang="ts">
	import { onMount } from 'svelte';
	import { user } from '$lib/stores/user';
	import { api } from '$lib/api';
	import GameList from '$lib/components/GameList.svelte';
	import { goto } from '$app/navigation';
	import type { GameSummary } from '$lib/types/game_summary';

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

	async function onDeleteGame(id: string) {
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

	async function onRenameGame(id: string, currentName: string) {
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

<!-- NAVBAR -->
<nav class="flex justify-between items-center px-6 py-4 bg-white shadow-md">
	<div class="text-2xl font-bold text-green-700 tracking-tight">🧩 Scrabble</div>
	<div class="flex space-x-6 items-center">
		{#if !$user}
			<a href="/login" class="bg-green-600 text-white px-4 py-2 rounded hover:bg-green-700">Connexion</a>
		{:else}
			<a href="/login" class="bg-green-600 text-white px-4 py-2 rounded hover:bg-green-700">Jouer</a>
		{/if}
	</div>
</nav>

<main class="max-w-4xl mx-auto px-6 py-10">
	{#if !$user}
		<section class="text-center mt-12">
			<h1 class="text-4xl font-extrabold mb-4 text-gray-800">Bienvenue sur Scrabble en ligne</h1>
			<p class="text-gray-600 text-lg mb-6">
				Jouez au Scrabble en ligne avec vos amis, gratuitement, sans pub !
			</p>
			<div class="flex justify-center gap-4">
				<a href="/login" class="bg-green-600 text-white px-6 py-2 rounded hover:bg-green-700">
					Connexion
				</a>
				<a href="/register" class="bg-gray-200 text-gray-700 px-6 py-2 rounded hover:bg-gray-300">
					Inscription
				</a>
			</div>
		</section>
	{:else}
		<div class="mb-6">
			<h2 class="text-xl font-semibold mb-2">Mes parties en cours</h2>
			<GameList
				{games}
				onDelete={onDeleteGame}
				onRename={onRenameGame}
			/>
		</div>
		<button on:click={createGame} class="bg-green-600 text-white px-4 py-2 rounded hover:bg-green-700">
			Créer une nouvelle partie
		</button>
	{/if}
</main>
