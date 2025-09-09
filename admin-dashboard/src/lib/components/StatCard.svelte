<script lang="ts">
    export let title: string;
    export let value: string | number| undefined;
    export let delta: number | null = null;
    export let icon: string | null = null;

    function deltaIsPositive(d: number | null) {
        if (d === null) return null;
        console.log('Delta value:', d);
        return d > 0;
    }
</script>

<div class="group relative bg-white/4 hover:bg-white/5 transition-colors duration-200 rounded-lg p-4 flex flex-col items-center text-center">
    {#if icon}
        <div class="mb-3 w-14 h-14 rounded-full bg-white/6 flex items-center justify-center transform transition-transform duration-200 group-hover:scale-105">
            {@html icon}
        </div>
    {/if}

    <div class="text-3xl font-extrabold">{value}</div>

    <div class="mt-2 h-6">
        {#if delta != null}
            {#if deltaIsPositive(delta)}
                <div class="inline-flex items-center gap-1 text-emerald-300 text-sm">
                    <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg"><path d="M12 5v14" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/><path d="M5 12l7-7 7 7" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>
                    <span>{delta}%</span>
                </div>
            {:else}
                <div class="inline-flex items-center gap-1 text-rose-300 text-sm">
                    <svg class="w-4 h-4" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg"><path d="M12 19V5" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/><path d="M19 12l-7 7-7-7" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>
                    <span>{delta}%</span>
                </div>
            {/if}
        {/if}
    </div>

    <!-- Tooltip bubble: hidden by default, appears on hover -->
    <div role="tooltip" class="absolute left-1/2 -translate-x-1/2 -top-14 opacity-0 pointer-events-none group-hover:opacity-100 group-hover:pointer-events-auto transition-opacity duration-150">
        <div class="bg-slate-800 text-sm text-white/90 px-3 py-2 rounded shadow-lg">{title}</div>
        <div class="absolute left-1/2 -translate-x-1/2 top-full w-3 h-3 bg-slate-800 rotate-45"></div>
    </div>
</div>
