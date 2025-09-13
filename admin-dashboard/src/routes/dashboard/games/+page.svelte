<script lang="ts">
    import { onMount } from 'svelte';
    import { api } from '$lib/api';
  import { goto } from '$app/navigation';

    type AdminGame = {
        id: string;
        name: string;
        status: string;
        created_at: string; // ISO
        current_turn_user_id?: number;
        current_turn_username?: string;
        winner_username?: string;
        ended_at?: string;
        pass_count: number;
        players_count: number;
        moves_count: number;
        last_play_time: string; // ISO
        created_by_username?: string;
    };

    // data
    let games: AdminGame[] = [];
    let filtered: AdminGame[] = [];
    let loading = true;

    // filters
    let qName = '';
    let qCreatedBy = '';
    let qCurrentTurn = '';
    let qWinner = '';
    let statusFilter: 'all' | 'ongoing' | 'ended' = 'all';

    // tri
    type Sortable = keyof Pick<AdminGame, 'last_play_time' | 'created_at' | 'name' | 'players_count' | 'moves_count' | 'pass_count' | 'status'>;
    let sortField: Sortable = 'last_play_time';
    let sortDir: 'asc' | 'desc' = 'desc';

    function applyFilters() {
        filtered = games.filter(g => {
            if (statusFilter !== 'all' && g.status !== statusFilter) return false;
            if (qName && !g.name.toLowerCase().includes(qName.toLowerCase())) return false;
            if (qCreatedBy && !(g.created_by_username || '').toLowerCase().includes(qCreatedBy.toLowerCase())) return false;
            if (qCurrentTurn && !(g.current_turn_username || '').toLowerCase().includes(qCurrentTurn.toLowerCase())) return false;
            if (qWinner && !(g.winner_username || '').toLowerCase().includes(qWinner.toLowerCase())) return false;
            return true;
        });
        sortNow();
    }

    function sortNow() {
        filtered = filtered.slice().sort((a, b) => {
            let va: any = a[sortField];
            let vb: any = b[sortField];
            if (sortField === 'created_at' || sortField === 'last_play_time') {
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
            sortDir = field === 'name' ? 'asc' : 'desc';
        }
        sortNow();
    }

    async function fetchGames() {
        try {
            const res = await api.get('/admin/games');
            const items = (res.data?.games || []) as any[];
            games = items.map((g: any) => ({
                id: g.id,
                name: g.name,
                status: g.status,
                created_at: g.created_at,
                current_turn_user_id: g.current_turn_user_id ?? undefined,
                current_turn_username: g.current_turn_username ?? '',
                winner_username: g.winner_username ?? '',
                ended_at: g.ended_at ?? undefined,
                pass_count: g.pass_count ?? 0,
                players_count: g.players_count ?? 0,
                moves_count: g.moves_count ?? 0,
                last_play_time: g.last_play_time ?? g.created_at,
                created_by_username: g.created_by_username ?? ''
            })) as AdminGame[];
            filtered = games.slice();
            applyFilters();
        } catch (err) {
            console.error('failed to fetch admin games', err);
            games = [];
            filtered = [];
        }
    }

    onMount(async () => {
        try {
            await fetchGames();
        } finally {
            loading = false;
        }
    });

    function openGame(g: AdminGame) {
        goto(`/dashboard/games/${g.id}`);
    }
</script>

<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <header class="mb-6">
        <h1 class="text-2xl font-bold">Parties</h1>
        <p class="text-sm text-white/70 mt-1">Liste des parties (toutes, en cours et terminées) avec filtres et tri.</p>
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
                    <option class="text-black" value="ongoing">En cours</option>
                    <option class="text-black" value="ended">Terminée</option>
                </select>
            </div>

            <div class="flex items-center gap-2">
                <label for="filter-name" class="text-sm text-white/70 w-28">Nom</label>
                <input id="filter-name" class="bg-white/5 px-2 py-1 rounded flex-1" placeholder="nom de partie" bind:value={qName} oninput={applyFilters} />
            </div>

            <div class="flex items-center gap-2">
                <label for="filter-created-by" class="text-sm text-white/70 w-28">Créateur</label>
                <input id="filter-created-by" class="bg-white/5 px-2 py-1 rounded flex-1" placeholder="username" bind:value={qCreatedBy} oninput={applyFilters} />
            </div>

            <div class="flex items-center gap-2">
                <label for="filter-current" class="text-sm text-white/70 w-28">Tour courant</label>
                <input id="filter-current" class="bg-white/5 px-2 py-1 rounded flex-1" placeholder="username" bind:value={qCurrentTurn} oninput={applyFilters} />
            </div>

            <div class="flex items-center gap-2">
                <label for="filter-winner" class="text-sm text-white/70 w-28">Gagnant</label>
                <input id="filter-winner" class="bg-white/5 px-2 py-1 rounded flex-1" placeholder="username" bind:value={qWinner} oninput={applyFilters} />
            </div>

            <div class="flex items-center gap-2 justify-end md:justify-start">
                <button class="px-3 py-1.5 bg-white/6 rounded hover:bg-white/8" onclick={fetchGames}>Rafraîchir</button>
            </div>
        </div>
    </section>

    <!-- Tableau -->
    <section class="bg-white/4 rounded-lg p-4">
        <div class="overflow-x-auto">
            <table class="min-w-full text-sm table-fixed w-full">
                <colgroup>
                    <col style="width:16%" />
                    <col style="width:18%" />
                    <col style="width:10%" />
                    <col style="width:8%" />
                    <col style="width:8%" />
                    <col style="width:8%" />
                    <col style="width:16%" />
                    <col style="width:16%" />
                </colgroup>
                <thead>
                    <tr class="text-left text-xs text-white/60">
                        <th class="px-3 py-2 cursor-pointer" onclick={() => toggleSort('last_play_time')}>Dernière activité {sortField==='last_play_time'? (sortDir==='asc'?'▲':'▼') : ''}</th>
                        <th class="px-3 py-2 cursor-pointer" onclick={() => toggleSort('name')}>Nom {sortField==='name'? (sortDir==='asc'?'▲':'▼') : ''}</th>
                        <th class="px-3 py-2 cursor-pointer" onclick={() => toggleSort('status')}>Statut {sortField==='status'? (sortDir==='asc'?'▲':'▼') : ''}</th>
                        <th class="px-3 py-2 cursor-pointer" onclick={() => toggleSort('players_count')}>Joueurs {sortField==='players_count'? (sortDir==='asc'?'▲':'▼') : ''}</th>
                        <th class="px-3 py-2 cursor-pointer" onclick={() => toggleSort('moves_count')}>Coups {sortField==='moves_count'? (sortDir==='asc'?'▲':'▼') : ''}</th>
                        <th class="px-3 py-2 cursor-pointer" onclick={() => toggleSort('pass_count')}>Pass {sortField==='pass_count'? (sortDir==='asc'?'▲':'▼') : ''}</th>
                        <th class="px-3 py-2">Créée par</th>
                        <th class="px-3 py-2">Tour / Gagnant</th>
                    </tr>
                </thead>
                <tbody>
                    {#if loading}
                        <tr><td colspan="8" class="px-3 py-4 text-center text-white/60">Chargement...</td></tr>
                    {:else}
                        {#each filtered as g}
                            <tr
                                class="border-t border-white/6 hover:bg-white/6 hover:cursor-pointer"
                                onclick="{() => openGame(g)}"
                            >
                                <td class="px-3 py-2 align-top text-xs text-white/70">{new Date(g.last_play_time).toLocaleString()}</td>
                                <td class="px-3 py-2 align-top text-sm text-white/90 truncate" title={g.name}>{g.name}</td>
                                <td class="px-3 py-2 align-top">
                                    {#if g.status === 'ended'}
                                        <span class="inline-block px-2 py-0.5 rounded-full text-xs font-medium bg-emerald-600 text-white">FINIE</span>
                                    {:else if g.status === 'ongoing'}
                                        <span class="inline-block px-2 py-0.5 rounded-full text-xs font-medium bg-sky-600 text-white">EN COURS</span>
                                    {:else}
                                        <span class="inline-block px-2 py-0.5 rounded-full text-xs font-medium bg-zinc-600 text-white">{g.status}</span>
                                    {/if}
                                </td>
                                <td class="px-3 py-2 align-top text-xs text-white/70">{g.players_count}</td>
                                <td class="px-3 py-2 align-top text-xs text-white/70">{g.moves_count}</td>
                                <td class="px-3 py-2 align-top text-xs text-white/70">{g.pass_count}</td>
                                <td class="px-3 py-2 align-top text-xs text-white/70">{g.created_by_username}</td>
                                <td class="px-3 py-2 align-top text-xs text-white/80">
                                    {#if g.status === 'ended'}
                                        <div>Gagnant: <span class="text-white">{g.winner_username || '—'}</span></div>
                                        <!-- {#if g.ended_at}<div class="text-white/60">Fin: {new Date(g.ended_at).toLocaleString()}</div>{/if} -->
                                    {:else}
                                        <div>Tour: <span class="text-white">{g.current_turn_username || '—'}</span></div>
                                    {/if}
                                </td>
                            </tr>
                        {/each}
                        {#if filtered.length === 0}
                            <tr><td colspan="8" class="px-3 py-4 text-center text-white/60">Aucune partie trouvée</td></tr>
                        {/if}
                    {/if}
                </tbody>
            </table>
        </div>
    </section>
</div>
