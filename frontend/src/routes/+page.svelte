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
		<GameList {games} />
	</div>

	<button on:click={createGame} class="bg-green-600 text-white px-4 py-2 rounded">
		Créer une nouvelle partie
	</button>
{/if}
