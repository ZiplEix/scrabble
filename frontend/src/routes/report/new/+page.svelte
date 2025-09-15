<script lang="ts">
	import { api } from '$lib/api';
	import { goto } from '$app/navigation';
	import HeaderBar from '$lib/components/HeaderBar.svelte';

	let title = $state('');
	let content = $state('');
	let type = $state('bug');
	let error = $state('');
	let success = $state(false);

	async function submitReport() {
		error = '';
		success = false;

		content = content.trim();
		title = title.trim();

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

<HeaderBar title="Nouveau ticket" back={true} />
<main class="max-w-md mx-auto px-4 py-8">
	<h1 class="text-2xl font-bold mb-4 text-center text-gray-800">Faire un ticket</h1>

	<div class="mx-auto rounded-sm ring-1 ring-black/5 bg-white shadow p-4">
		<!-- svelte-ignore event_directive_deprecated -->
		<form on:submit|preventDefault={submitReport} class="flex flex-col gap-4">
			<input
				type="text"
				class="w-full bg-white rounded-lg px-4 py-3 text-sm placeholder-gray-400 shadow-sm ring-1 ring-black/5 focus:outline-none focus:ring-2 focus:ring-emerald-500 transition"
				placeholder="Titre du report"
				bind:value={title}
				required
			/>

			<textarea
				class="w-full bg-white rounded-lg px-4 py-3 text-sm placeholder-gray-400 shadow-sm ring-1 ring-black/5 focus:outline-none focus:ring-2 focus:ring-emerald-500 resize-none h-32 transition"
				placeholder="Décrivez le bug ou votre suggestion"
				bind:value={content}
				required
			></textarea>

			<select
				class="w-full bg-white rounded-lg px-4 py-3 text-sm placeholder-gray-400 shadow-sm ring-1 ring-black/5 focus:outline-none focus:ring-2 focus:ring-emerald-500 transition"
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
	</div>
</main>
