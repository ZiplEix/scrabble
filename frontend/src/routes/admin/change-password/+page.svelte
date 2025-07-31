<script lang="ts">
	import { api } from '$lib/api';
	import { user } from '$lib/stores/user';
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { get } from 'svelte/store';

	let username = '';
	let newPassword = '';
	let error = '';
	let success = '';

	onMount(() => {
		const current = get(user);
		if (!current?.token) {
			goto('/');
		}
	});

	async function handleChange() {
		error = '';
		success = '';
		try {
			await api.post(
				'/auth/change-password',
				{ username, new_password: newPassword },
			);
			success = 'Mot de passe changé avec succès !';
			username = '';
			newPassword = '';
		} catch (err: any) {
			error = err?.response?.data?.message || 'Échec du changement de mot de passe';
		}
	}
</script>

<main class="max-w-sm mx-auto px-4 py-8">
	<h1 class="text-2xl font-bold mb-6 text-center text-gray-800">Changer le mot de passe</h1>

	<form on:submit|preventDefault={handleChange} class="flex flex-col gap-4">
		<input
			type="text"
			placeholder="Nom d’utilisateur"
			bind:value={username}
			required
			class="border rounded px-4 py-3 text-sm focus:outline-none focus:ring-2 focus:ring-green-500"
		/>

		<input
			type="password"
			placeholder="Nouveau mot de passe"
			bind:value={newPassword}
			required
			minlength="4"
			class="border rounded px-4 py-3 text-sm focus:outline-none focus:ring-2 focus:ring-green-500"
		/>

		{#if error}
			<p class="text-sm text-red-600 text-center">{error}</p>
		{/if}

		{#if success}
			<p class="text-sm text-green-600 text-center">{success}</p>
		{/if}

		<button
			type="submit"
			class="bg-green-600 hover:bg-green-700 text-white rounded py-3 font-semibold transition"
		>
			Valider
		</button>
	</form>
</main>
