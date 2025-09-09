<script lang="ts">
    import DashboardSidebar from '$lib/components/DashboardSidebar.svelte';

	let { children } = $props();
</script>

<div class="min-h-screen bg-slate-900 text-white">
    <DashboardSidebar />

    <div class="pl-0 md:pl-64">
        {@render children?.()}
    </div>
</div>

