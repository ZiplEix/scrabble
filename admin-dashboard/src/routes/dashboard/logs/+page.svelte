<script lang="ts">
    import { onMount } from 'svelte';

    type Log = {
        id: number;
        date: string; // ISO
        level: 'info' | 'warn' | 'error';
        route: string;
        message: string;
        username?: string;
        method?: string;
        status?: number;
        reason?: string;
        raw?: any;
    }

    // UI state
    let logs: Log[] = [];
    let filtered: Log[] = [];
    let loading = true;

    // filters
    let qRoute = '';
    let qUser = '';
    let qMethod = '';
    let qStatus = '';
    let qReason = '';
    let levelFilter: string = 'all';
    // include admin logs from server
    let includeAdmin = false;

    // selection
    let selected = new Set<number>();
    let selectAll = false;

    // sorting
    let sortField: keyof Log = 'date';
    let sortDir: 'asc' | 'desc' = 'desc';

    function toggleSelect(id: number) {
        if (selected.has(id)) selected.delete(id);
        else selected.add(id);
        selectAll = filtered.length > 0 && selected.size === filtered.length;
    }

    function toggleSelectAll() {
        if (selectAll) {
            selected = new Set();
            selectAll = false;
        } else {
            selected = new Set(filtered.map(l => l.id));
            selectAll = true;
        }
    }

    function applyFilters() {
        filtered = logs.filter(l => {
            if (levelFilter !== 'all' && l.level !== levelFilter) return false;
            if (qRoute && !l.route.toLowerCase().includes(qRoute.toLowerCase())) return false;
            if (qUser && !(l.username || '').toLowerCase().includes(qUser.toLowerCase())) return false;
            if (qMethod && !(l.method || '').toLowerCase().includes(qMethod.toLowerCase())) return false;
            if (qStatus && String(l.status || '').indexOf(qStatus) === -1) return false;
            if (qReason && !(l.reason || '').toLowerCase().includes(qReason.toLowerCase())) return false;
            return true;
        });
        sortAndPaginate();
    }

    function sortAndPaginate() {
        filtered = filtered.sort((a, b) => {
            let va: any = a[sortField];
            let vb: any = b[sortField];
            if (sortField === 'date') {
                va = new Date(va).getTime();
                vb = new Date(vb).getTime();
            }
            if (va == null) return 1;
            if (vb == null) return -1;
            if (va > vb) return sortDir === 'asc' ? 1 : -1;
            if (va < vb) return sortDir === 'asc' ? -1 : 1;
            return 0;
        });
    }

    function toggleSort(field: keyof Log) {
        if (sortField === field) {
            sortDir = sortDir === 'asc' ? 'desc' : 'asc';
        } else {
            sortField = field;
            sortDir = 'asc';
        }
        sortAndPaginate();
    }

    import { api } from '$lib/api';

    const PAGE_SIZE = 50;
    let page = 1;
    let hasMore = false;
    let loadingMore = false;

    async function fetchPage(p: number) {
        try {
            const adminParam = includeAdmin ? '&admin=1' : '';
            const res = await api.get(`/admin/logs?page=${p}${adminParam}`);
            const items = res.data?.logs || [];
            const normalized = items.map((it: any, idx: number) => ({
                id: it.id ?? (p - 1) * PAGE_SIZE + idx + 1,
                date: it.date ?? it.received_at ?? new Date().toISOString(),
                level: it.level ?? (it.raw?.level || 'info'),
                route: (it.route || (it.raw && (it.raw.route || it.raw.path)) || ''),
                message: ((it.message || (it.raw && (it.raw.msg || it.raw.message))) as string).replace('http_request', '').trim(),
                username: it.username ?? it.raw?.username ?? undefined,
                method: it.method ?? it.raw?.method ?? undefined,
                status: it.status ?? it.raw?.status ?? undefined,
                reason: it.reason ?? it.raw?.reason ?? undefined,
                raw: it.raw ?? it.raw
            }));

            if (p === 1) {
                logs = normalized;
            } else {
                logs = logs.concat(normalized);
            }

            // determine if there are more pages
            hasMore = items.length === PAGE_SIZE;

            filtered = logs.slice();
            sortAndPaginate();
            return normalized.length;
        } catch (err) {
            console.error('failed to fetch logs page', p, err);
            return 0;
        }
    }

    // initial load: page 1
    onMount(async () => {
        try {
            const n = await fetchPage(1);
            // if less than page size, no more pages
            hasMore = n === PAGE_SIZE;
            page = 1;
        } finally {
            loading = false;
        }
    });

    async function reloadWithAdmin() {
        loading = true;
        page = 1;
        logs = [];
        filtered = [];
        try {
            const n = await fetchPage(1);
            hasMore = n === PAGE_SIZE;
            page = 1;
        } finally {
            loading = false;
        }
    }

    async function loadMore() {
        if (loadingMore || !hasMore) return;
        loadingMore = true;
        try {
            const next = page + 1;
            const n = await fetchPage(next);
            if (n > 0) page = next;
            if (n < PAGE_SIZE) hasMore = false;
        } finally {
            loadingMore = false;
        }
    }
</script>

<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <header class="mb-6">
        <h1 class="text-2xl font-bold">Logs</h1>
        <p class="text-sm text-white/70 mt-1">Visualisez, filtrez et sélectionnez les logs.</p>
    </header>

    <!-- Filters -->
    <section class="mb-4 bg-white/4 rounded-lg p-4">
    <div class="grid grid-cols-1 md:grid-cols-3 gap-3">
            <div class="flex items-center gap-2">
                <label for="filter-level" class="text-sm text-white/70 w-24">Niveau</label>
                <select
                    id="filter-level"
                    bind:value={levelFilter}
                    on:change={applyFilters}
                    class="px-2 py-1 rounded bg-white/5 text-white border border-white/10 focus:outline-none focus:ring-2 focus:ring-white/20"
                >
                    <option class="text-black" value="all">Tous</option>
                    <option class="text-black" value="info">Info</option>
                    <option class="text-black" value="warn">Warn</option>
                    <option class="text-black" value="error">Error</option>
                </select>
            </div>

            <div class="flex items-center gap-2">
                <label for="filter-route" class="text-sm text-white/70 w-24">Route</label>
                <input id="filter-route" class="bg-white/5 px-2 py-1 rounded flex-1" placeholder="/api/..." bind:value={qRoute} on:input={applyFilters} />
            </div>

            <div class="flex items-center gap-2">
                <label for="filter-user" class="text-sm text-white/70 w-24">Utilisateur</label>
                <input id="filter-user" class="bg-white/5 px-2 py-1 rounded flex-1" placeholder="username" bind:value={qUser} on:input={applyFilters} />
            </div>

            <div class="flex items-center gap-2">
                <label for="filter-method" class="text-sm text-white/70 w-24">Méthode</label>
                <input id="filter-method" class="bg-white/5 px-2 py-1 rounded flex-1" placeholder="GET/POST" bind:value={qMethod} on:input={applyFilters} />
            </div>

            <div class="flex items-center gap-2">
                <label for="filter-status" class="text-sm text-white/70 w-24">Status</label>
                <input id="filter-status" class="bg-white/5 px-2 py-1 rounded flex-1" placeholder="200/401" bind:value={qStatus} on:input={applyFilters} />
            </div>

            <div class="flex items-center gap-2">
                <label for="filter-reason" class="text-sm text-white/70 w-24">Reason</label>
                <input id="filter-reason" class="bg-white/5 px-2 py-1 rounded flex-1" placeholder="invalid_credentials" bind:value={qReason} on:input={applyFilters} />
            </div>
            <div class="flex items-center gap-2 justify-end md:justify-start">
                <label class="text-sm text-white/70 w-24" for="include-admin">Admin</label>
                <div class="flex items-center gap-2">
                    <input id="include-admin" type="checkbox" bind:checked={includeAdmin} on:change={reloadWithAdmin} />
                    <label for="include-admin" class="text-sm text-white/80">Inclure logs admin</label>
                </div>
            </div>
        </div>
    </section>

    <!-- Table -->
    <section class="bg-white/4 rounded-lg p-4">
        <div class="overflow-x-auto">
            <table class="min-w-full text-sm table-fixed w-full">
                <colgroup>
                    <col style="width:4%" />
                    <col style="width:16%" />
                    <col style="width:8%" />
                    <col style="width:8%" />
                    <col style="width:18%" />
                    <col style="width:8%" />
                    <col style="width:12%" />
                    <col style="width:26%" />
                </colgroup>
                <thead>
                    <tr class="text-left text-xs text-white/60">
                        <th class="px-3 py-2"><input type="checkbox" checked={selectAll} on:change={toggleSelectAll} /></th>
                        <th class="px-3 py-2 cursor-pointer" on:click={() => toggleSort('date')}>Date {sortField==='date'? (sortDir==='asc'?'▲':'▼') : ''}</th>
                        <th class="px-3 py-2 cursor-pointer" on:click={() => toggleSort('level')}>Niveau {sortField==='level'? (sortDir==='asc'?'▲':'▼') : ''}</th>
                        <th class="px-3 py-2">Méthode</th>
                        <th class="px-3 py-2 cursor-pointer" on:click={() => toggleSort('route')}>Route {sortField==='route'? (sortDir==='asc'?'▲':'▼') : ''}</th>
                        <th class="px-3 py-2">Status</th>
                        <th class="px-3 py-2">Utilisateur</th>
                        <th class="px-3 py-2">Message / Reason</th>
                    </tr>
                </thead>
                <tbody>
                    {#if loading}
                        <tr><td colspan="8" class="px-3 py-4 text-center text-white/60">Chargement...</td></tr>
                    {:else}
                        {#each filtered as log}
                            <tr
                                class="border-t border-white/6 hover:bg-white/6 hover:cursor-pointer"
                                on:click={() => window.location.href=`/dashboard/logs/${log.id}`}
                            >
                                <td class="px-3 py-2 align-top">
                                    <input type="checkbox" checked={selected.has(log.id)} on:change={() => toggleSelect(log.id)} />
                                </td>
                                <td class="px-3 py-2 align-top text-xs text-white/70">{new Date(log.date).toLocaleString()}</td>
                                <td class="px-3 py-2 align-top">
                                    {#if log.level === 'error'}
                                        <span class=" inline-block px-2 py-0.5 rounded-full text-xs font-medium bg-red-600 text-white">ERROR</span>
                                    {:else if log.level === 'warn'}
                                        <span class="inline-block px-2 py-0.5 rounded-full text-xs font-medium bg-amber-600 text-black">WARN</span>
                                    {:else}
                                        <span class="inline-block px-2 py-0.5 rounded-full text-xs font-medium bg-green-600 text-white">INFO</span>
                                    {/if}
                                </td>
                                <td class="px-3 py-2 align-top text-xs text-white/70">{log.method}</td>
                                <td class="px-3 py-2 align-top text-xs text-white/70">{log.route}</td>
                                <td class="px-3 py-2 align-top text-xs text-white/70">{log.status}</td>
                                <td class="px-3 py-2 align-top text-xs text-white/70">{log.username}</td>
                                <td class="px-3 py-2 align-top">
                                    <div class="truncate text-white/90">{log.reason}</div>
                                </td>
                            </tr>
                        {/each}
                        {#if filtered.length === 0}
                            <tr><td colspan="8" class="px-3 py-4 text-center text-white/60">Aucun log trouvé</td></tr>
                        {/if}
                    {/if}
                </tbody>
            </table>
        </div>
    </section>
    <!-- Load more -->
    <div class="mt-4 flex justify-center">
        {#if hasMore}
            <button class="px-4 py-2 bg-white/6 rounded hover:bg-white/8" on:click={loadMore} disabled={loadingMore}>
                {#if loadingMore}Chargement...{:else}Charger plus{/if}
            </button>
        {/if}
    </div>
</div>
