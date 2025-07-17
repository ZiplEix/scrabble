<script lang="ts">
	import { api } from '$lib/api';
	import { goto } from '$app/navigation';

	let title = $state('');
	let content = $state('');
	let type = $state('bug');
	let error = $state('');
	let success = $state(false);

	async function submitReport() {
		error = '';
		success = false;

		if (!title || !content) {
			error = 'Veuillez remplir tous les champs obligatoires.';
			return;
		}

		try {
			await api.post('/report', { title, content, type });
			success = true;
			title = '';
			content = '';
			type = 'bug';
		} catch (err: any) {
			error = err?.response?.data?.error || 'Erreur lors de l’envoi du report';
		}
	}
</script>

<main class="max-w-md mx-auto px-4 py-8">
	<h1 class="text-2xl font-bold mb-6 text-center text-gray-800">Faire un ticket</h1>

	<!-- svelte-ignore event_directive_deprecated -->
	<form on:submit|preventDefault={submitReport} class="flex flex-col gap-4">
		<input
			type="text"
			class="border rounded px-4 py-3 text-sm focus:outline-none focus:ring-2 focus:ring-green-500"
			placeholder="Titre du report"
			bind:value={title}
			required
		/>

		<textarea
			class="border rounded px-4 py-3 text-sm focus:outline-none focus:ring-2 focus:ring-green-500 resize-none h-32"
			placeholder="Décrivez le bug ou votre suggestion"
			bind:value={content}
			required
		></textarea>

		<select
			class="border rounded px-4 py-3 text-sm focus:outline-none focus:ring-2 focus:ring-green-500"
			bind:value={type}
		>
			<option value="bug">🐞 Bug</option>
			<option value="suggestion">💡 Suggestion</option>
			<option value="feedback">📢 Feedback</option>
			<option value="other">❓ Autre</option>
		</select>

		{#if error}
			<p class="text-sm text-red-600 text-center">{error}</p>
		{/if}

		{#if success}
			<p class="text-sm text-green-600 text-center">Merci pour votre retour !</p>
		{/if}

		<button
			type="submit"
			class="bg-green-600 hover:bg-green-700 text-white rounded py-3 font-semibold transition"
		>
			Envoyer le report
		</button>
	</form>
</main>
