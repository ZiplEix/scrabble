<script lang="ts">
	import { api } from '$lib/api';
	import { goto } from '$app/navigation';

	let name = '';
	let error = '';
	let loading = false;

	async function createGame() {
		error = '';
		if (name.trim().length === 0) {
			error = 'Le nom de la partie est obligatoire';
			return;
		}

		loading = true;
		try {
			const res = await api.post('/game', { name });
			const gameId = res.data.game_id;
			goto(`/games/${gameId}`);
		} catch (e: any) {
			error = e?.response?.data?.error || 'Erreur lors de la création de la partie';
		} finally {
			loading = false;
		}
	}
</script>

<h1 class="text-2xl font-bold mb-6">Créer une nouvelle partie</h1>

<form on:submit|preventDefault={createGame} class="max-w-md space-y-4">
	<div>
		<label for="name" class="block font-semibold mb-1">Nom de la partie</label>
		<input
			id="name"
			type="text"
			bind:value={name}
			placeholder="Nom de la partie"
			class="block w-full border rounded px-3 py-2"
			autocomplete="off"
		/>
	</div>

	{#if error}
		<p class="text-red-600">{error}</p>
	{/if}

	<button
		type="submit"
		class="bg-green-600 text-white px-4 py-2 rounded disabled:opacity-50"
		disabled={loading}
	>
		{loading ? 'Création...' : 'Créer la partie'}
	</button>
</form>
