<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import { goto } from '$app/navigation';
	import HeaderBar from '$lib/components/HeaderBar.svelte';
	import Card from '$lib/ui/Card.svelte';
	import Button from '$lib/ui/Button.svelte';

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
			const res = await api.get('/report/me');
			reports = res.data;
		} catch (err: any) {
			error = 'Impossible de charger les tickets.';
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

<HeaderBar title="Mes tickets" back={true} />
<main class="max-w-3xl mx-auto px-4 py-6">
	{#if error}
		<p class="text-red-600 text-center">{error}</p>
	{:else if reports.length === 0}
		<p class="text-center text-gray-600">Aucun ticket pour le moment.</p>
	{:else}
		<div class="flex flex-col gap-3">
			{#each reports as report}
				<Card>
					<button type="button" class="w-full text-left cursor-pointer" onclick={() => goto(`/report/${report.id}`)}>
						<div class="flex justify-between items-start">
							<h2 class="text-base font-semibold text-gray-900 truncate pr-2">{report.title}</h2>
							<span class="text-[11px] px-2 py-0.5 rounded-full font-medium {statusColor(report.status)}">
								{report.status}
							</span>
						</div>
						<p class="text-[12px] text-gray-700 mt-1">Type : <strong>{report.type}</strong></p>
						<p class="text-[12px] text-gray-500">Par <strong>{report.username}</strong> — {new Date(report.created_at).toLocaleDateString()}</p>
					</button>
				</Card>
			{/each}
		</div>
	{/if}

	<div class="mt-6 flex justify-center">
		<Button variant="primary" size="lg" onclick={() => goto('/report/new')}>
			Nouveau ticket
		</Button>
	</div>

	<!-- Floating action button (bottom-right) for quick create -->
	<div class="fixed right-4 bottom-[calc(env(safe-area-inset-bottom)+16px)] z-40">
		<button
			onclick={() => goto('/report/new')}
			class="inline-flex items-center justify-center w-14 h-14 rounded-full bg-green-600 text-white shadow-lg hover:bg-green-700 active:scale-95 transition"
			aria-label="Nouveau ticket"
			title="Nouveau ticket"
		>
			<svg width="22" height="22" viewBox="0 0 24 24" fill="none" aria-hidden="true"><path d="M12 5v14M5 12h14" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg>
		</button>
	</div>
</main>
