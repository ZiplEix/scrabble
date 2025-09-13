<script lang='ts'>
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { get } from 'svelte/store';
	import { api } from '$lib/api';

	let idParam = get(page).params.id;

	type Report = {
		id: number;
		title: string;
		content: string;
		status: 'open' | 'in_progress' | 'resolved' | 'rejected' | string;
		priority: string;
		type: string;
		username: string;
		created_at: string;
		updated_at: string;
	}

	let loading = true;
	let err: string | null = null;
	let report: Report | null = null;

	async function loadReport() {
		loading = true;
		err = null;
		report = null;
		try {
			const res = await api.get(`/report/${idParam}`);
			const r = res.data;
			if (!r) {
				err = `Ticket ${idParam} introuvable`;
				return;
			}
			report = {
				id: Number(r.id),
				title: r.title,
				content: r.content,
				status: r.status,
				priority: r.priority,
				type: r.type,
				username: r.username,
				created_at: r.created_at,
				updated_at: r.updated_at,
			};
		} catch (e: any) {
			console.error('Erreur chargement ticket', e);
			err = e?.message || String(e);
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		loadReport();
	});
</script>

<div class="max-w-5xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
	<header class="mb-6">
		<h1 class="text-2xl font-bold">Ticket {idParam}</h1>
		{#if report}
			<p class="text-sm text-white/70 mt-1">{report.title}</p>
		{/if}
	</header>

	{#if loading}
		<div class="space-y-3">
			<div class="h-6 w-64 bg-white/10 animate-pulse rounded"></div>
			<div class="h-4 w-80 bg-white/10 animate-pulse rounded"></div>
			<div class="h-72 w-full bg-white/10 animate-pulse rounded"></div>
		</div>
	{:else if err}
		<div class="bg-red-700/20 border border-red-700/30 rounded p-4">{err}</div>
	{:else if report}
		<section class="mb-6 bg-white/4 rounded-lg p-4">
			<h2 class="text-lg font-semibold mb-3">Métadonnées</h2>
			<div class="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
				<div><span class="text-white/60">ID:</span> <span class="text-white/90">{report.id}</span></div>
				<div><span class="text-white/60">Créé le:</span> <span class="text-white/90">{new Date(report.created_at).toLocaleString()}</span></div>
				<div><span class="text-white/60">MAJ le:</span> <span class="text-white/90">{new Date(report.updated_at).toLocaleString()}</span></div>
				<div><span class="text-white/60">Statut:</span> <span class="text-white/90">{report.status}</span></div>
				<div><span class="text-white/60">Priorité:</span> <span class="text-white/90">{report.priority}</span></div>
				<div><span class="text-white/60">Type:</span> <span class="text-white/90">{report.type}</span></div>
				<div><span class="text-white/60">Auteur:</span> <span class="text-white/90">{report.username}</span></div>
			</div>
		</section>

		<section class="bg-white/4 rounded-lg p-4">
			<h2 class="text-lg font-semibold mb-3">Contenu</h2>
			<div class="whitespace-pre-wrap text-white/90 text-sm">{report.content}</div>
		</section>
	{:else}
		<div class="text-white/70">Ticket introuvable.</div>
	{/if}

	<div class="mt-6">
		<a href="/dashboard/tickets" class="text-sm text-white/70 hover:underline">← Retour aux tickets</a>
	</div>
</div>
