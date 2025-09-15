<script lang='ts'>
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { get } from 'svelte/store';
	import { api } from '$lib/api';
    import { goto } from '$app/navigation';

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
	let saving = false;
	let err: string | null = null;
	let report: Report | null = null;
	let deleting = false;

	const statuses = ['open', 'in_progress', 'resolved', 'rejected'];
	const types = ['bug', 'suggestion', 'feedback'];

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

	async function updateReport() {
		if (!report) return;
		saving = true;
		try {
			await api.patch(`/report/${report.id}`, {
				title: report.title,
				content: report.content,
				status: report.status,
				type: report.type,
			});
			alert('Ticket mis à jour');
		} catch (e: any) {
			console.error('Erreur mise à jour ticket', e);
			alert(e?.response?.data?.error || 'Erreur lors de la mise à jour');
		} finally {
			saving = false;
		}
	}

	async function deleteReport() {
		if (!report) return;
		const ok = confirm(`Supprimer définitivement le ticket #${report.id} ?`);
		if (!ok) return;
		deleting = true;
		try {
			await api.delete(`/report/${report.id}`);
			alert('Ticket supprimé');
			goto('/dashboard/tickets');
		} catch (e: any) {
			console.error('Erreur suppression ticket', e);
			alert(e?.response?.data?.error || 'Erreur lors de la suppression');
		} finally {
			deleting = false;
		}
	}
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
				<div><span class="text-white/60">Priorité:</span> <span class="text-white/90">{report.priority}</span></div>
				<div><span class="text-white/60">Auteur:</span> <span class="text-white/90">{report.username}</span></div>
			</div>
		</section>

		<section class="bg-white/4 rounded-lg p-4">
			<h2 class="text-lg font-semibold mb-4">Édition</h2>
			<form class="space-y-4" on:submit|preventDefault={updateReport}>
				<div>
					<label class="block text-sm text-white/70 mb-1" for="title">Titre</label>
					<input id="title" class="w-full bg-white/10 border border-white/10 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500" type="text" bind:value={report.title} required />
				</div>

				<div>
					<label class="block text-sm text-white/70 mb-1" for="content">Contenu</label>
					<textarea id="content" rows="8" class="w-full bg-white/10 border border-white/10 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500" bind:value={report.content} required></textarea>
				</div>

				<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
					<div>
						<label class="block text-sm text-white/70 mb-1" for="type">Type</label>
						<select id="type" class="w-full bg-white text-gray-900 border border-white/10 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500" bind:value={report.type}>
							{#each types as t}
								<option value={t}>{t}</option>
							{/each}
						</select>
					</div>
					<div>
						<label class="block text-sm text-white/70 mb-1" for="status">Statut</label>
						<select id="status" class="w-full bg-white text-gray-900 border border-white/10 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500" bind:value={report.status}>
							{#each statuses as s}
								<option value={s}>{s}</option>
							{/each}
						</select>
					</div>
				</div>

				<div class="pt-2 flex items-center gap-3">
					<button type="submit" class="bg-emerald-600 hover:bg-emerald-700 disabled:opacity-60 text-white rounded px-4 py-2 text-sm font-semibold" disabled={saving || deleting}>
						{#if saving}Enregistrement...{/if}
						{#if !saving}Enregistrer{/if}
					</button>
					<button type="button" class="bg-red-600 hover:bg-red-700 disabled:opacity-60 text-white rounded px-4 py-2 text-sm font-semibold" on:click={deleteReport} disabled={saving || deleting}>
						{#if deleting}Suppression...{/if}
						{#if !deleting}Supprimer{/if}
					</button>
				</div>
			</form>
		</section>
	{:else}
		<div class="text-white/70">Ticket introuvable.</div>
	{/if}

	<div class="mt-6">
		<a href="/dashboard/tickets" class="text-sm text-white/70 hover:underline">← Retour aux tickets</a>
	</div>
</div>
