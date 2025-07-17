<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { api } from '$lib/api';
	import { goto } from '$app/navigation';

	let report: any = $state(null);
	let error = $state('');
	let loading = $state(true);
	let reportId = $state('');

	const statuses = ['open', 'in_progress', 'resolved', 'rejected'];
	const types = ['bug', 'suggestion', 'feedback'];

	onMount(async () => {
		reportId = $page.params.id;
		if (!reportId) {
			error = 'Aucun ID de report fourni.';
			loading = false;
			return;
		}

		try {
			const res = await api.get(`/report/${reportId}`);
			report = res.data;
		} catch (err: any) {
			error = err?.response?.data?.error || 'Impossible de charger le report.';
		} finally {
			loading = false;
		}
	});

	async function updateReport() {
		try {
			await api.patch(`/report/${report.id}`, {
				title: report.title,
				content: report.content,
				status: report.status,
				type: report.type,
			});
			alert('Report mis à jour');
			goto('/report');
		} catch (err) {
			alert('Erreur lors de la mise à jour');
			console.error(err);
		}
	}
</script>

{#if loading}
	<p class="text-center mt-8">Chargement...</p>
{:else if error}
	<p class="text-center text-red-600 mt-8">{error}</p>
{:else}
	<main class="max-w-sm mx-auto px-2 py-8">
		<h1 class="text-2xl font-bold mb-6 text-center text-gray-800">
			Ticket #{report.id}
		</h1>

		<!-- svelte-ignore event_directive_deprecated -->
		<form on:submit|preventDefault={updateReport} class="flex flex-col gap-4">
            <div>
                <label for="title" class="block text-sm font-medium text-gray-700 mb-1">Titre</label>
                <input
                    id="title"
                    type="text"
                    placeholder="Titre"
                    bind:value={report.title}
                    class="w-full border rounded px-4 py-3 text-sm focus:outline-none focus:ring-2 focus:ring-green-500"
                    required
                />
            </div>

            <div>
                <label for="content" class="block text-sm font-medium text-gray-700 mb-1">Contenu</label>
                <textarea
                    id="content"
                    rows="6"
                    placeholder="Contenu du rapport"
                    bind:value={report.content}
                    class="w-full border rounded px-4 py-3 text-sm focus:outline-none focus:ring-2 focus:ring-green-500 resize-y"
                    required
                ></textarea>
            </div>

            <div>
				<label for="type" class="block text-sm font-medium text-gray-700 mb-1">Type</label>
				<select
                    id="type"
                    bind:value={report.type}
                    class="w-full border rounded px-4 py-3 text-sm focus:outline-none focus:ring-2 focus:ring-green-500"
                >
					{#each types as t}
						<option value={t}>{t}</option>
					{/each}
				</select>
			</div>

            <div>
				<label for="status" class="block text-sm font-medium text-gray-700 mb-1">Statut</label>
				<select
					id="status"
					bind:value={report.status}
					class="w-full border rounded px-4 py-3 text-sm focus:outline-none focus:ring-2 focus:ring-green-500"
				>
					{#each statuses as s}
						<option value={s}>{s}</option>
					{/each}
				</select>
			</div>

			<button
				type="submit"
				class="bg-green-600 hover:bg-green-700 text-white rounded py-3 font-semibold transition"
			>
				Enregistrer
			</button>
		</form>
	</main>
{/if}
