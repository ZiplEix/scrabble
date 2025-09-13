<script lang='ts'>
    import { onMount } from 'svelte';
    import { api } from '$lib/api';
  import { goto } from '$app/navigation';

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

    // Data
    let reports: Report[] = [];
    let filtered: Report[] = [];
    let loading = true;
    // plus d'actions, pas de loadingAction nécessaire

    // Filters
    let q = '';
    let statusFilter: 'all' | 'open' | 'in_progress' | 'resolved' | 'rejected' = 'all';
    let priorityFilter: 'all' | string = 'all';
    let typeFilter: 'all' | string = 'all';
    // option "onlyMine" retirée

    // Derived filter options from data
    $: priorities = Array.from(new Set(reports.map(r => r.priority).filter(Boolean)));
    $: types = Array.from(new Set(reports.map(r => r.type).filter(Boolean)));

    // Sorting
    type Sortable = keyof Pick<Report, 'created_at' | 'updated_at' | 'title' | 'status' | 'priority' | 'type'>;
    let sortField: Sortable = 'created_at';
    let sortDir: 'asc' | 'desc' = 'desc';

    function applyFilters() {
        filtered = reports.filter(r => {
            if (statusFilter !== 'all' && r.status !== statusFilter) return false;
            if (priorityFilter !== 'all' && r.priority !== priorityFilter) return false;
            if (typeFilter !== 'all' && r.type !== typeFilter) return false;
            if (q) {
                const Q = q.toLowerCase();
                const hay = `${r.title} ${r.content} ${r.username} ${r.type} ${r.priority}`.toLowerCase();
                if (!hay.includes(Q)) return false;
            }
            return true;
        });
        sortNow();
    }

    function sortNow() {
        filtered = filtered.slice().sort((a, b) => {
            let va: any = a[sortField];
            let vb: any = b[sortField];
            if (sortField === 'created_at' || sortField === 'updated_at') {
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

    function toggleSort(field: Sortable) {
        if (sortField === field) {
            sortDir = sortDir === 'asc' ? 'desc' : 'asc';
        } else {
            sortField = field;
            sortDir = 'asc';
        }
        sortNow();
    }

    async function loadReports() {
        loading = true;
        try {
            const res = await api.get('/report');
            const items: any[] = res.data || [];
            reports = items.map(r => ({
                id: Number(r.id),
                title: r.title,
                content: r.content,
                status: r.status,
                priority: r.priority,
                type: r.type,
                username: r.username,
                created_at: r.created_at,
                updated_at: r.updated_at,
            }));
            filtered = reports.slice();
            applyFilters();
        } catch (e) {
            console.error('Erreur chargement tickets', e);
        } finally {
            loading = false;
        }
    }


    onMount(() => {
        loadReports();
    });
</script>

<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <header class="mb-6">
        <h1 class="text-2xl font-bold">Tickets</h1>
        <p class="text-sm text-white/70 mt-1">Visualisez, filtrez et sélectionnez les tickets.</p>
    </header>

    <!-- Filtres -->
    <section class="mb-4 bg-white/4 rounded-lg p-4">
        <div class="grid grid-cols-1 md:grid-cols-3 gap-3">
            <div class="flex items-center gap-2">
                <label for="filter-status" class="text-sm text-white/70 w-28">Statut</label>
                <select
                    id="filter-status"
                    bind:value={statusFilter}
                    onchange={applyFilters}
                    class="px-2 py-1 rounded bg-white/5 text-white border border-white/10 focus:outline-none focus:ring-2 focus:ring-white/20"
                >
                    <option class="text-black" value="all">Tous</option>
                    <option class="text-black" value="open">Ouvert</option>
                    <option class="text-black" value="in_progress">En cours</option>
                    <option class="text-black" value="resolved">Résolu</option>
                    <option class="text-black" value="rejected">Rejeté</option>
                </select>
            </div>

            <div class="flex items-center gap-2">
                <label for="filter-priority" class="text-sm text-white/70 w-28">Priorité</label>
                <select id="filter-priority" bind:value={priorityFilter} onchange={applyFilters}
                    class="px-2 py-1 rounded bg-white/5 text-white border border-white/10 focus:outline-none focus:ring-2 focus:ring-white/20">
                    <option class="text-black" value="all">Toutes</option>
                    {#each priorities as p}
                        <option class="text-black" value={p}>{p}</option>
                    {/each}
                </select>
            </div>

            <div class="flex items-center gap-2">
                <label for="filter-type" class="text-sm text-white/70 w-28">Type</label>
                <select id="filter-type" bind:value={typeFilter} onchange={applyFilters}
                    class="px-2 py-1 rounded bg-white/5 text-white border border-white/10 focus:outline-none focus:ring-2 focus:ring-white/20">
                    <option class="text-black" value="all">Tous</option>
                    {#each types as t}
                        <option class="text-black" value={t}>{t}</option>
                    {/each}
                </select>
            </div>

            <div class="flex items-center gap-2 md:col-span-2">
                <label for="q" class="text-sm text-white/70 w-28">Recherche</label>
                <input id="q" class="bg-white/5 px-2 py-1 rounded flex-1" placeholder="titre, contenu, utilisateur" bind:value={q} oninput={applyFilters} />
            </div>
        </div>
    </section>

    <!-- Actions de masse retirées -->

    <!-- Table -->
    <section class="bg-white/4 rounded-lg p-4">
        <div class="overflow-x-auto">
            <table class="min-w-full text-sm table-fixed w-full">
                <colgroup>
                    <col style="width:20%" />
                    <col style="width:18%" />
                    <col style="width:18%" />
                    <col style="width:14%" />
                    <col style="width:30%" />
                </colgroup>
                <thead>
                    <tr class="text-left text-xs text-white/60">
                        <th class="px-3 py-2 cursor-pointer" onclick={() => toggleSort('created_at')}>Créé {sortField==='created_at'? (sortDir==='asc'?'▲':'▼') : ''}</th>
                        <th class="px-3 py-2 cursor-pointer" onclick={() => toggleSort('status')}>Statut {sortField==='status'? (sortDir==='asc'?'▲':'▼') : ''}</th>
                        <th class="px-3 py-2 cursor-pointer" onclick={() => toggleSort('priority')}>Priorité {sortField==='priority'? (sortDir==='asc'?'▲':'▼') : ''}</th>
                        <th class="px-3 py-2 cursor-pointer" onclick={() => toggleSort('type')}>Type {sortField==='type'? (sortDir==='asc'?'▲':'▼') : ''}</th>
                        <th class="px-3 py-2 cursor-pointer" onclick={() => toggleSort('title')}>Titre {sortField==='title'? (sortDir==='asc'?'▲':'▼') : ''}</th>
                    </tr>
                </thead>
                <tbody>
                    {#if loading}
                        {#each Array(8) as _, i}
                            <tr class="border-t border-white/6">
                                <td class="px-3 py-3"><div class="h-8 w-32 bg-white/10 animate-pulse rounded"></div></td>
                                <td class="px-3 py-3"><div class="h-8 w-20 bg-white/10 animate-pulse rounded"></div></td>
                                <td class="px-3 py-3"><div class="h-8 w-16 bg-white/10 animate-pulse rounded"></div></td>
                                <td class="px-3 py-3"><div class="h-8 w-16 bg-white/10 animate-pulse rounded"></div></td>
                                <td class="px-3 py-3"><div class="h-8 w-40 bg-white/10 animate-pulse rounded"></div></td>
                            </tr>
                        {/each}
                    {:else}
                        {#each filtered as r}
                            <tr
                                class="border-t border-white/6 hover:bg-white/6 hover:cursor-pointer"
                                onclick="{() => goto(`/dashboard/tickets/${r.id}`)}"
                            >
                                <td class="px-3 py-2 align-top text-xs text-white/70">{new Date(r.created_at).toLocaleString()}</td>
                                <td class="px-3 py-2 align-top">
                                    {#if r.status === 'resolved'}
                                        <span class="inline-block px-2 py-0.5 rounded-full text-xs font-medium bg-green-600 text-white">RESOLU</span>
                                    {:else if r.status === 'in_progress'}
                                        <span class="inline-block px-2 py-0.5 rounded-full text-xs font-medium bg-indigo-600 text-white">EN COURS</span>
                                    {:else if r.status === 'rejected'}
                                        <span class="inline-block px-2 py-0.5 rounded-full text-xs font-medium bg-amber-600 text-black">REJETÉ</span>
                                    {:else}
                                        <span class="inline-block px-2 py-0.5 rounded-full text-xs font-medium bg-slate-600 text-white">OUVERT</span>
                                    {/if}
                                </td>
                                <td class="px-3 py-2 align-top text-xs text-white/80">{r.priority}</td>
                                <td class="px-3 py-2 align-top text-xs text-white/80">{r.type}</td>
                                <td class="px-3 py-2 align-top text-sm text-white/90 truncate">{r.title}</td>
                            </tr>
                        {/each}
                        {#if filtered.length === 0}
                            <tr><td colspan="5" class="px-3 py-4 text-center text-white/60">Aucun ticket</td></tr>
                        {/if}
                    {/if}
                </tbody>
            </table>
        </div>
    </section>
</div>
