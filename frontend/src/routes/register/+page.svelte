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
			user.set({ username, token: res.data.token });
			goto('/');
		} catch (err: any) {
			error = err?.response?.data?.message || 'Échec de l’inscription';
		}
	}
</script>

<main class="max-w-sm mx-auto px-4 py-8">
	<h1 class="text-2xl font-bold mb-6 text-center text-gray-800">Créer un compte</h1>

	<form on:submit|preventDefault={handleRegister} class="flex flex-col gap-4">
		<div>
			<input
				class="border rounded px-4 py-3 text-sm w-full focus:outline-none focus:ring-2 focus:ring-green-500"
				type="text"
				placeholder="Nom d'utilisateur"
				bind:value={username}
				required
			/>
			<p class="text-xs text-gray-500 mt-1">
				Ce nom sera visible par les autres joueurs et utilisé pour vous inviter à des parties.
			</p>
		</div>

		<input
			class="border rounded px-4 py-3 text-sm focus:outline-none focus:ring-2 focus:ring-green-500"
			type="password"
			placeholder="Mot de passe"
			bind:value={password}
			required
		/>

		{#if error}
			<p class="text-sm text-red-600 text-center">{error}</p>
		{/if}

		<button
			type="submit"
			class="bg-green-600 hover:bg-green-700 text-white rounded py-3 font-semibold transition"
		>
			S’inscrire
		</button>
	</form>


	<p class="mt-6 text-sm text-center text-gray-600">
		Déjà inscrit ?
		<a href="/login" class="text-green-600 font-medium underline">Connexion</a>
	</p>
</main>
