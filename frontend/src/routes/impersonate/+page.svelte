<script lang="ts">
	import { api } from '$lib/api';
	import { user } from '$lib/stores/user';
	import { goto } from '$app/navigation';

	let username = '';
	let error = '';

	async function handleImpersonate() {
		error = '';
		try {
			const res = await api.get(`/auth/connect-as?user=${username}`);
			const userNameToStore = username.trim().toLowerCase();
			user.set({ username: userNameToStore, token: res.data.token });
			goto('/');
		} catch (err: any) {
			error = err?.response?.data?.message || 'Échec de la connexion';
		}
	}
</script>

<main class="max-w-sm mx-auto px-4 py-8">
	<h1 class="text-2xl font-bold mb-6 text-center text-gray-800">Impersonate</h1>

	<form on:submit|preventDefault={handleImpersonate} class="flex flex-col gap-4">
		<input
			class="border rounded px-4 py-3 text-sm focus:outline-none focus:ring-2 focus:ring-green-500"
			type="text"
			placeholder="Nom d'utilisateur"
			bind:value={username}
			required
		/>

		{#if error}
			<p class="text-sm text-red-600 text-center">{error}</p>
		{/if}

		<button
			type="submit"
			class="bg-green-600 hover:bg-green-700 text-white rounded py-3 font-semibold transition"
		>
			Impersonate
		</button>
	</form>
</main>
