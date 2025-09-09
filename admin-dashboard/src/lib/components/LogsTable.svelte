<script lang="ts">
    type LogEntry = { level: 'info' | 'warn' | 'error'; date: string; route: string; message: string };
    export let logs: LogEntry[] = [];
    export let title: string = 'Derniers logs';
    export let loading: boolean = false;
</script>

<div class="bg-white/4 rounded-lg p-4">
    <h2 class="text-lg font-medium mb-3">{title}</h2>
        <div>
            <table class="min-w-full text-sm table-fixed w-full">
                <colgroup>
                    <col style="width:8%" />
                    <col style="width:22%" />
                    <col style="width:22%" />
                    <col style="width:48%" />
                </colgroup>
            <thead>
                <tr class="text-left text-xs text-white/60">
                    <th class="px-3 py-2">Niveau</th>
                    <th class="px-3 py-2">Date</th>
                    <th class="px-3 py-2">Route</th>
                    <th class="px-3 py-2">Message</th>
                </tr>
            </thead>
            <tbody>
                {#if loading}
                    {#each Array(3) as _, i}
                        <tr class="border-t border-white/6">
                            <td class="px-3 py-2 align-top">
                                <div class="h-4 w-12 bg-white/6 rounded animate-pulse"></div>
                            </td>
                            <td class="px-3 py-2 align-top text-xs text-white/70">
                                <div class="h-3 w-24 bg-white/6 rounded animate-pulse"></div>
                            </td>
                            <td class="px-3 py-2 align-top text-xs text-white/70">
                                <div class="h-3 w-28 bg-white/6 rounded animate-pulse"></div>
                            </td>
                            <td class="px-3 py-2 align-top">
                                <div class="h-3 w-full bg-white/6 rounded animate-pulse"></div>
                            </td>
                        </tr>
                    {/each}
                {:else}
                    {#each logs as log}
                        <tr class="border-t border-white/6">
                            <td class="px-3 py-2 align-top">
                                {#if log.level === 'error'}
                                    <span class=" inline-block px-2 py-0.5 rounded-full text-xs font-medium bg-red-600 text-white">ERROR</span>
                                {:else if log.level === 'warn'}
                                    <span class="inline-block px-2 py-0.5 rounded-full text-xs font-medium bg-amber-600 text-black">WARN</span>
                                {:else}
                                    <span class="inline-block px-2 py-0.5 rounded-full text-xs font-medium bg-green-600 text-white">INFO</span>
                                {/if}
                            </td>
                            <td class="px-3 py-2 align-top text-xs text-white/70">{new Date(log.date).toLocaleString()}</td>
                            <td class="px-3 py-2 align-top text-xs text-white/70">{log.route}</td>
                            <td class="px-3 py-2 align-top">
                                <div class="truncate text-white/90">{log.message}</div>
                            </td>
                        </tr>
                    {/each}
                {/if}
            </tbody>
        </table>
    </div>
</div>
