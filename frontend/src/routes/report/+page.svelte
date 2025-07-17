<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import { goto } from '$app/navigation';

	type Report = {
		id: number;
		title: string;
		type: string;
		status: string;
		created_at: string;
		username: string;
	};

	let reports: Report[] = [];
	let error = '';

	onMount(async () => {
		try {
			const res = await api.get('/report');
			reports = res.data;
		} catch (err: any) {
			error = 'Impossible de charger les reports';
			console.error(err);
		}
	});

	function statusColor(status: string) {
		switch (status) {
			case 'open':
				return 'bg-yellow-100 text-yellow-700';
			case 'in_progress':
				return 'bg-blue-300 text-blue-700';
			case 'resolved':
				return 'bg-green-100 text-green-700';
			case 'rejected':
				return 'bg-red-200 text-red-700';
			default:
				return 'bg-gray-200 text-gray-700';
		}
	}
</script>

<main class="max-w-3xl mx-auto px-4 py-8">
	<h1 class="text-2xl font-bold mb-6 text-center text-gray-800">Liste des tickets</h1>

	{#if error}
		<p class="text-red-600 text-center">{error}</p>
	{:else if reports.length === 0}
		<p class="text-center text-gray-600">Aucun ticket pour le moment.</p>
	{:else}
		<div class="flex flex-col gap-4">
			{#each reports as report}
				<!-- svelte-ignore a11y_click_events_have_key_events -->
				<!-- svelte-ignore a11y_no_static_element_interactions -->
				<div
					class="border rounded p-4 shadow hover:shadow-md transition cursor-pointer hover:bg-gray-50"
					onclick={() => goto(`/report/${report.id}`)}
				>
					<div class="flex justify-between items-start">
						<h2 class="text-lg font-semibold text-gray-800">{report.title}</h2>
						<span class="text-xs px-2 py-1 rounded font-medium {statusColor(report.status)}">
							{report.status}
						</span>
					</div>
					<p class="text-sm text-gray-600 mt-1">Type : <strong>{report.type}</strong></p>
					<p class="text-sm text-gray-500">Par <strong>{report.username}</strong> le {new Date(report.created_at).toLocaleDateString()}</p>
				</div>
			{/each}
		</div>
	{/if}

    <div class="mt-8 flex justify-center">
		<button
			class="bg-green-600 hover:bg-green-700 text-white rounded px-5 py-3 font-semibold transition"
			onclick={() => goto('/report/new')}
		>
			Créer un nouveau ticket
		</button>
	</div>
</main>
