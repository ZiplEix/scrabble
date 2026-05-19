<script lang="ts">
    import { onMount } from 'svelte';
    import HeaderBar from '$lib/components/HeaderBar.svelte';
    import { api } from '$lib/api';
    import RankBadge from '$lib/components/RankBadge.svelte';
    import { user } from '$lib/stores/user';

    type LeaderboardEntry = {
        rank: number;
        user_id: number;
        username: string;
        rating: number;
        games: number;
    };

    let loading = $state(true);
    let error = $state<string | null>(null);
    let entries = $state<LeaderboardEntry[]>([]);
    let total = $state(0);

    onMount(async () => {
        try {
            const res = await api.get('/leaderboard?limit=100&offset=0');
            entries = res.data?.entries ?? [];
            total = res.data?.total ?? entries.length;
        } catch (e: any) {
            error = e?.response?.data?.message || 'Erreur lors du chargement du classement';
        } finally {
            loading = false;
        }
    });

    function rankStyle(rank: number): string {
        if (rank === 1) return 'bg-amber-50 text-amber-700 ring-amber-200';
        if (rank === 2) return 'bg-gray-50 text-gray-700 ring-gray-200';
        if (rank === 3) return 'bg-orange-50 text-orange-700 ring-orange-200';
        return 'bg-emerald-50 text-emerald-700 ring-emerald-100';
    }
</script>

<HeaderBar title="Classement global" back={true} />

<main class="max-w-2xl mx-auto px-4 py-6">
    <section class="mb-4">
        <div class="rounded-2xl bg-emerald-50 ring-1 ring-black/5 p-4">
            <h2 class="text-lg font-bold text-gray-900">Classement Elo</h2>
            <p class="text-[12px] text-gray-700 mt-1">
                Top joueurs triés par score Elo. {total} joueur{total > 1 ? 's' : ''} au total.
            </p>
            <a href="/leaderboard/info" class="inline-flex mt-3 text-xs font-semibold text-emerald-700 hover:text-emerald-800 hover:underline">
                Comprendre le système Elo et les rangs
            </a>
        </div>
    </section>

    {#if loading}
        <div class="rounded-sm ring-1 ring-black/5 bg-white shadow p-5 text-center text-gray-500">
            Chargement du classement...
        </div>
    {:else if error}
        <div class="rounded-sm ring-1 ring-red-200 bg-red-50 p-4 text-sm text-red-700">
            {error}
        </div>
    {:else if entries.length === 0}
        <div class="rounded-sm ring-1 ring-black/5 bg-white shadow p-5 text-center text-gray-500">
            Aucun joueur à afficher.
        </div>
    {:else}
        <div class="rounded-sm ring-1 ring-black/5 bg-white shadow overflow-hidden">
            <ul class="divide-y divide-gray-100">
                {#each entries as entry}
                    <li class="p-3 sm:p-4 {entry.user_id === $user?.id ? 'bg-emerald-50/60' : ''}">
                        <div class="flex items-center justify-between gap-3">
                            <div class="flex items-center gap-3 min-w-0">
                                <div class="w-9 h-9 rounded-full ring-1 flex items-center justify-center text-sm font-bold shrink-0 {rankStyle(entry.rank)}">
                                    {entry.rank}
                                </div>

                                <div class="min-w-0">
                                    <div class="flex items-center gap-2">
                                        <a href={'/user/' + entry.user_id} class="text-sm sm:text-base font-semibold text-gray-900 truncate hover:underline">
                                            {entry.username}
                                        </a>
                                        {#if entry.user_id === $user?.id}
                                            <span class="text-[11px] px-2 py-0.5 rounded-full bg-emerald-600 text-white">Vous</span>
                                        {/if}
                                    </div>
                                    <p class="text-[12px] text-gray-500">{entry.games} partie{entry.games > 1 ? 's' : ''}</p>
                                </div>
                            </div>

                            <div class="text-right flex items-center gap-2">
                                <RankBadge rating={entry.rating} size="md" />
                                <div class="text-lg font-extrabold text-emerald-700">{entry.rating}</div>
                            </div>
                        </div>
                    </li>
                {/each}
            </ul>
        </div>
    {/if}
</main>