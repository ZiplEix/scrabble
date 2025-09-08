<script lang="ts">
    import { user, logout } from '$lib/stores/user';
    import { onMount } from 'svelte';
    import StatCard from '$lib/components/StatCard.svelte';
    import MinimalLineChart from '$lib/components/MinimalLineChart.svelte';
    import LogsTable from '$lib/components/LogsTable.svelte';

    // placeholder values — replace with API data
    const stats = {
        activeUsersWeek: 842,
        gamesCreatedWeek: 123,
        gamesActive: 37,
        messagesWeek: 1450,
        ticketsWeek: 12
    };

    // mock 48 points (last 48 hours) - random-ish but smoothed
    const points: number[] = Array.from({ length: 48 }, (_, i) => {
        const base = 20 + Math.sin(i / 6) * 6 + Math.random() * 6;
        return Math.round(base + (i > 36 ? 8 : 0));
    });

    // generate hour labels for last 48 hours, e.g. '00h', '01h' ...
    const labels: string[] = Array.from({ length: 48 }, (_, i) => {
        const d = new Date(Date.now() - (47 - i) * 60 * 60 * 1000);
        return d.getHours().toString().padStart(2, '0') + 'h';
    });

    // mock last 10 logs (replace with API call later)
    type LogEntry = { level: 'info' | 'warn' | 'error'; date: string; route: string; message: string };
    const logs: LogEntry[] = Array.from({ length: 10 }, (_, i) => {
        const level = (i % 3 === 0) ? 'error' : (i % 3 === 1) ? 'warn' : 'info';
        return {
            level,
            date: new Date(Date.now() - i * 60 * 1000).toISOString(),
            route: `/api/v1/resource/${i % 5}`,
            message: `Example log message number ${i + 1} — some extra details that will be truncated in the table view.`
        };
    });

    onMount(() => {
        // future: fetch dashboard summary data
    });
</script>

<div class="min-h-screen bg-slate-900 text-white">
    <!-- Sidebar fixed left (desktop) -->
    <nav class="hidden md:flex fixed left-0 top-0 h-screen w-64 bg-white/5 p-4 flex-col justify-between">
        <div>
            <div class="text-sm font-medium">Navigation</div>
            <ul class="mt-2 space-y-1">
                <li><a href="/dashboard" class="block px-3 py-2 rounded hover:bg-white/5">Accueil</a></li>
                <li><a href="/dashboard/logs" class="block px-3 py-2 rounded hover:bg-white/5">Logs</a></li>
                <li><a href="/dashboard/users" class="block px-3 py-2 rounded hover:bg-white/5">Utilisateurs</a></li>
                <li><a href="/dashboard/games" class="block px-3 py-2 rounded hover:bg-white/5">Parties</a></li>
                <li><a href="/dashboard/tickets" class="block px-3 py-2 rounded hover:bg-white/5">Tickets</a></li>
            </ul>
        </div>

        <div class="pt-4">
            <button class="w-full text-left px-3 py-2 rounded bg-white/6 hover:bg-white/8 text-sm" on:click={() => logout()}>Se déconnecter</button>
        </div>
    </nav>

    <!-- Main content area (shifted right on desktop to make room for sidebar) -->
    <div class="pl-0 md:pl-64">
        <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">

            <!-- Header / title -->
            <header class="mb-6">
                <h1 class="text-3xl font-bold">Scrabble Admin Dashboard</h1>
                <p class="text-sm text-white/70 mt-1">Résumé rapide des indicateurs clés</p>
            </header>

            <!-- Top stats row -->
            <section class="mb-6">
                <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-4">
                    <StatCard title="Utilisateurs actifs (semaine)" value={stats.activeUsersWeek} delta="+5%" icon="<svg class='w-6 h-6 text-white' xmlns='http://www.w3.org/2000/svg' fill='none' viewBox='0 0 24 24' stroke='currentColor'><path stroke-linecap='round' stroke-linejoin='round' stroke-width='1.5' d='M16 11c1.657 0 3-1.343 3-3S17.657 5 16 5s-3 1.343-3 3 1.343 3 3 3zM8 11c1.657 0 3-1.343 3-3S9.657 5 8 5 5 6.343 5 8s1.343 3 3 3zM8 13c-2.667 0-8 1.333-8 4v2h16v-2c0-2.667-5.333-4-8-4zM16 13c-.29 0-.577.014-.86.041C15.284 14.16 17 15 17 15v2h3v-2c0-2.667-5.333-4-3-4z'/></svg>" />

                    <StatCard title="Parties créées (semaine)" value={stats.gamesCreatedWeek} delta="+8%" icon="<svg class='w-6 h-6 text-white' xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='none' stroke='currentColor' stroke-width='1.5'><path stroke-linecap='round' stroke-linejoin='round' d='M12 8v8M8 12h8'/></svg>" />

                    <StatCard title="Parties en cours" value={stats.gamesActive} delta="-2%" icon="<svg class='w-6 h-6 text-white' xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='none' stroke='currentColor' stroke-width='1.5'><path stroke-linecap='round' stroke-linejoin='round' d='M5 3v18l15-9L5 3z'/></svg>" />

                    <StatCard title="Messages envoyés (semaine)" value={stats.messagesWeek} delta="+12%" icon="<svg class='w-6 h-6 text-white' xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='none' stroke='currentColor' stroke-width='1.5'><path stroke-linecap='round' stroke-linejoin='round' d='M21 15a2 2 0 01-2 2H7l-4 4V5a2 2 0 012-2h14a2 2 0 012 2z'/></svg>" />

                    <StatCard title="Tickets créés (semaine)" value={stats.ticketsWeek} delta="+0%" icon="<svg class='w-6 h-6 text-white' xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='none' stroke='currentColor' stroke-width='1.5'><path stroke-linecap='round' stroke-linejoin='round' d='M3 8v8a2 2 0 002 2h14a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2z'/></svg>" />
                </div>
            </section>

            <!-- Minimal 48-hour requests chart -->
            <section class="mb-8">
                <div class="bg-white/4 rounded-lg p-4">
                    <div class="w-full">
                        <MinimalLineChart height={200} data={points} labels={labels} xTickStep={6} color="#60a5fa" strokeWidth={1.6} />
                    </div>
                </div>
            </section>

            <!-- Last 10 logs table -->
            <section class="mb-8">
                <LogsTable {logs} />
            </section>
        </div>
    </div>
</div>
