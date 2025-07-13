<script lang="ts">
	import { api } from '$lib/api';
	import { user } from '$lib/stores/user';
	import { goto } from '$app/navigation';

	let username = '';
	let password = '';
	let error = '';

	async function handleRegister() {
		error = '';
		try {
			const res = await api.post('/auth/register', { username, password });
			user.set({
				username,
				token: res.data.token,
			});
			goto('/');
		} catch (err: any) {
			error = err?.response?.data?.error || 'Register failed';
		}
	}
</script>

<h1 class="text-2xl font-bold mb-4">Register</h1>

<form on:submit|preventDefault={handleRegister} class="space-y-4">
	<input class="block border px-3 py-2 w-full rounded" placeholder="Username" bind:value={username} />
	<input class="block border px-3 py-2 w-full rounded" type="password" placeholder="Password" bind:value={password} />

	{#if error}
		<p class="text-red-500">{error}</p>
	{/if}

	<button class="bg-green-600 text-white px-4 py-2 rounded">Register</button>
</form>

<p class="mt-4">Déjà inscrit ? <a href="/login" class="text-green-600 underline">Connexion</a></p>

<style>
</style>
