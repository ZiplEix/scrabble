<script lang="ts">
    import { logout } from '$lib/stores/user';
    import LogsTable from '$lib/components/LogsTable.svelte';
    import TopStats from '$lib/components/TopStats.svelte';
    import LogsGraph from '$lib/components/LogsGraph.svelte';

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
                <TopStats />
            </section>

            <!-- Minimal 48-hour requests chart -->
            <section class="mb-8">
                <LogsGraph />
            </section>

            <!-- Last 10 logs table -->
            <section class="mb-8">
                <LogsTable {logs} />
            </section>
        </div>
    </div>
</div>
