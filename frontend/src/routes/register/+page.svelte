<script lang="ts">
	import { api } from '$lib/api';
	import { user } from '$lib/stores/user';
	import { goto } from '$app/navigation';

	let username = '';
	let password = '';
	let error = '';
	let showPassword = $state(false);

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

	function togglePasswordVisibility() {
		showPassword = !showPassword;
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

		<div class="relative">
			<input
				class="w-full border rounded px-4 py-3 text-sm focus:outline-none focus:ring-2 focus:ring-green-500"
				type={showPassword ? 'text' : 'password'}
				placeholder="Mot de passe"
				bind:value={password}
				required
			/>
			<button
				type="button"
				on:click={togglePasswordVisibility}
				class="absolute inset-y-0 right-0 px-3 flex items-center text-gray-500 hover:text-gray-700"
				aria-label={showPassword ? 'Cacher le mot de passe' : 'Montrer le mot de passe'}
			>
				{#if showPassword}
					<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none"
						viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
						<path stroke-linecap="round" stroke-linejoin="round"
							d="M13.875 18.825A10.05 10.05 0 0112 19c-5.523
								0-10-4.477-10-10a9.96 9.96 0
								012.277-6.176m1.59-1.594A9.956
								9.956 0 0112 3c5.523 0 10 4.477
								10 10 0 2.042-.602 3.94-1.64 5.52M15
								12a3 3 0 11-6 0 3 3 0 016 0z"/>
						<path stroke-linecap="round" stroke-linejoin="round"
							d="M3 3l18 18"/>
					</svg>
				{:else}
					<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none"
						viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
						<path stroke-linecap="round" stroke-linejoin="round"
							d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/>
						<path stroke-linecap="round" stroke-linejoin="round"
							d="M2.458 12C3.732 7.943
								7.523 5 12 5c4.477 0 8.268 2.943
								9.542 7-1.274 4.057-5.065
								7-9.542 7-4.477 0-8.268-2.943-9.542-7z"/>
					</svg>
				{/if}
			</button>
    	</div>

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
