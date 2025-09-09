<script lang="ts">
    import TopStats from '$lib/components/TopStats.svelte';
    import LogsGraph from '$lib/components/LogsGraph.svelte';
    import DashboardLogsTable from '$lib/components/DashboardLogsTable.svelte';
</script>

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
        <DashboardLogsTable />
    </section>
</div>

