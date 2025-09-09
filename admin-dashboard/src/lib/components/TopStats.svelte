<script lang="ts">
    import { onMount } from "svelte";
    import StatCard from "./StatCard.svelte";
    import { api } from '$lib/api';
    import type { AdminStats } from "$lib/types/stats";

    let stats: AdminStats | null = $state<AdminStats | null>(null);
    let statsError: string | null = $state<string | null>(null);
    let statsLoading: boolean = $state(true);

    async function fetchStats(): Promise<AdminStats | string> {
        try {
            const res = await api.get<AdminStats>('/admin/stats');
            return res.data;
        } catch (err: any) {
            return err?.response?.data?.message || 'Échec de la connexion';
        } finally {
            statsLoading = false;
        }
    }

    onMount(async () => {
        const statsRes = await fetchStats();
        if (typeof statsRes === 'string') {
            statsError = statsRes;
        } else {
            stats = statsRes;
        }
    });
</script>

<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-4">
    {#if statsLoading}
        <!-- skeleton placeholders when loading -->
        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-4 col-span-full">
            {#each Array(5) as _, i}
                <div class="bg-white/6 rounded-lg p-4 animate-pulse flex flex-col items-center gap-3">
                    <div class="h-10 w-10 rounded-full bg-white/10"></div>
                    <div class="h-8 w-28 rounded bg-white/10"></div>
                    <div class="h-4 w-16 rounded bg-white/8"></div>
                    <div class="mt-auto h-3 w-20 rounded bg-white/8"></div>
                </div>
            {/each}
        </div>
    {:else if statsError}
        <div class="col-span-full text-center text-rose-300">Erreur: {statsError}</div>
    {:else if !stats}
        <div class="col-span-full text-center text-white/70">Aucune statistique disponible.</div>
    {:else}
        <StatCard title="Utilisateurs actifs (semaine)" value={stats.active_users_count} delta={stats.active_users_pct_change} icon="<svg class='w-6 h-6 text-white' xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 24 24' stroke='currentColor'><path stroke-linecap='round' stroke-linejoin='round' stroke-width='1.5' d='M16 11c1.657 0 3-1.343 3-3S17.657 5 16 5s-3 1.343-3 3 1.343 3 3 3zM8 11c1.657 0 3-1.343 3-3S9.657 5 8 5 5 6.343 5 8s1.343 3 3 3zM8 13c-2.667 0-8 1.333-8 4v2h16v-2c0-2.667-5.333-4-8-4zM16 13c-.29 0-.577.014-.86.041C15.284 14.16 17 15 17 15v2h3v-2c0-2.667-5.333-4-3-4z'/></svg>" />
        <StatCard title="Parties créées (semaine)" value={stats.created_games_count} delta={stats.created_games_pct_change} icon="<svg class='w-6 h-6 text-white' xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='none' stroke='currentColor' stroke-width='1.5'><path stroke-linecap='round' stroke-linejoin='round' d='M12 8v8M8 12h8'/></svg>" />
        <StatCard title="Parties en cours" value={stats.active_games_count} delta={stats.active_games_pct_change} icon="<svg class='w-6 h-6 text-white' xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='none' stroke='currentColor' stroke-width='1.5'><path stroke-linecap='round' stroke-linejoin='round' d='M5 3v18l15-9L5 3z'/></svg>" />
        <StatCard title="Messages envoyés (semaine)" value={stats.sent_messages_count} delta={stats.sent_messages_pct_change} icon="<svg class='w-6 h-6 text-white' xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='none' stroke='currentColor' stroke-width='1.5'><path stroke-linecap='round' stroke-linejoin='round' d='M21 15a2 2 0 01-2 2H7l-4 4V5a2 2 0 012-2h14a2 2 0 012 2z'/></svg>" />
        <StatCard title="Tickets créés (semaine)" value={stats.tickets_created_count} delta={stats.tickets_created_pct_change} icon="<svg class='w-6 h-6 text-white' xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='none' stroke='currentColor' stroke-width='1.5'><path stroke-linecap='round' stroke-linejoin='round' d='M3 8v8a2 2 0 002 2h14a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2z'/></svg>" />
    {/if}
</div>