<script lang="ts">
    import { onMount } from 'svelte';
    import { get } from 'svelte/store';
    import { page } from '$app/stores';
    import { api } from '$lib/api';

    let idParam: string = '';
    const params = get(page).params;
    idParam = params.id ?? '';

    type LogFull = {
        id: number;
        date: string;
        level: string;
        route?: string;
        message?: string;
        username?: string;
        method?: string;
        status?: number;
        reason?: string;
        request_id?: string;
        raw?: any;
    }

    let loading = true;
    let err: string | null = null;
    let log: LogFull | null = null;

    async function loadLog() {
        loading = true;
        err = null;
        log = null;
        try {
            const nid = Number(idParam);
            const res = await api.get(`/admin/log/${nid}`);
            const found = res.data;
            if (!found) {
                err = `Log ${idParam} introuvable`;
                return;
            }

            log = {
                id: found.id ?? found.ID,
                date: found.date ?? found.received_at ?? new Date().toISOString(),
                level: found.level ?? (found.raw?.level || 'info'),
                route: found.route || (found.raw && (found.raw.route || found.raw.path)) || '',
                message: found.message || (found.raw && (found.raw.msg || found.raw.message)),
                username: found.username ?? found.raw?.username ?? undefined,
                method: found.method ?? found.raw?.method ?? undefined,
                status: found.status ?? found.raw?.status ?? undefined,
                reason: found.reason ?? found.raw?.reason ?? undefined,
                request_id: found.request_id ?? found.req_id ?? found.requestId ?? undefined,
                raw: found.raw ?? found.raw
            };
        } catch (e: any) {
            console.error('failed to load log', e);
            err = e?.message || String(e);
        } finally {
            loading = false;
        }
    }

    function copyRaw() {
        if (!log) return;
        const txt = JSON.stringify(log.raw || {}, null, 2);
        navigator.clipboard?.writeText(txt).then(() => {
            // silent success
        }).catch(() => {});
    }

    onMount(() => {
        loadLog();
    });
</script>

<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <header class="mb-6">
        <h1 class="text-2xl font-bold">Détail du log {idParam}</h1>
        <p class="text-sm text-white/70 mt-1">Affiche toutes les informations collectées pour ce log.</p>
    </header>

    {#if loading}
        <div class="bg-white/4 rounded-lg p-6">Chargement...</div>
    {:else}
        {#if err}
            <div class="bg-red-700/20 border border-red-700/30 rounded p-4">{err}</div>
        {:else if log}
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div class="bg-white/4 rounded-lg p-4">
                    <h2 class="text-lg font-medium mb-3">Métadonnées</h2>
                    <dl class="grid grid-cols-1 gap-y-2 text-sm text-white/80">
                        <div class="flex justify-between"><dt class="font-medium">ID</dt><dd>{log.id}</dd></div>
                        <div class="flex justify-between"><dt class="font-medium">Date</dt><dd>{new Date(log.date).toLocaleString()}</dd></div>
                        <div class="flex justify-between"><dt class="font-medium">Niveau</dt><dd>{log.level}</dd></div>
                        <div class="flex justify-between"><dt class="font-medium">Route</dt><dd class="truncate max-w-xs">{log.route}</dd></div>
                        <div class="flex justify-between"><dt class="font-medium">Méthode</dt><dd>{log.method}</dd></div>
                        <div class="flex justify-between"><dt class="font-medium">Status</dt><dd>{log.status}</dd></div>
                        <div class="flex justify-between"><dt class="font-medium">Utilisateur</dt><dd>{log.username}</dd></div>
                        <div class="flex justify-between"><dt class="font-medium">Request ID</dt><dd class="truncate max-w-xs">{log.request_id}</dd></div>
                        <div class="flex justify-between"><dt class="font-medium">Reason</dt><dd class="truncate max-w-xs">{log.reason}</dd></div>
                    </dl>
                </div>

                <div class="bg-white/4 rounded-lg p-4">
                    <h2 class="text-lg font-medium mb-3">Message</h2>
                    <div class="text-sm text-white/90 mb-4">{log.message}</div>

                    <h3 class="text-md font-medium mb-2">Raw JSON</h3>
                    <div class="relative">
                        <pre class="rounded bg-white/6 p-3 overflow-auto text-xs" style="max-height:420px;"><code>{JSON.stringify(log.raw || {}, null, 2)}</code></pre>
                        <button class="absolute top-2 right-2 px-2 py-1 bg-white/6 rounded hover:cursor-pointer" on:click={copyRaw}>Copier</button>
                    </div>
                </div>
            </div>
        {/if}
    {/if}

    <div class="mt-6">
        <a href="/dashboard/logs" class="text-sm text-white/70 hover:underline">← Retour aux logs</a>
    </div>
</div>
