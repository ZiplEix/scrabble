<div class="flex flex-col overflow-hidden bg-gradient-to-b from-emerald-50 to-white" style="height: 100dvh;" aria-busy="true" role="status">
    <!-- Header skeleton -->
    <div class="px-3 pt-3 pb-2">
        <div class="flex items-center justify-between">
            <div class="flex items-center gap-2">
                <div class="w-9 h-9 rounded-full bg-emerald-200/70 animate-pulse"></div>
                <div class="h-4 w-28 rounded bg-emerald-200/70 animate-pulse"></div>
            </div>
            <div class="flex items-center gap-2">
                <div class="h-8 w-16 rounded-full bg-emerald-200/70 animate-pulse hidden sm:block"></div>
                <div class="w-9 h-9 rounded-full bg-emerald-200/70 animate-pulse"></div>
            </div>
        </div>
        <div class="mt-2 flex items-center gap-2">
            <div class="h-6 w-24 rounded-full bg-gray-200 animate-pulse"></div>
            <div class="h-6 w-16 rounded-full bg-gray-200 animate-pulse"></div>
            <div class="h-6 w-20 rounded-full bg-gray-200 animate-pulse"></div>
        </div>
    </div>

    <!-- Main skeleton -->
    <main class="flex-1 flex flex-col min-h-0 overflow-hidden">
        <!-- Board placeholder: affichage d'une grille 15x15 vide (sans lettres ni labels) -->
        <div class="flex-1 grid place-items-center px-3 min-h-0">
            <div class="mx-auto w-full max-w-[min(95vw,640px)]">
                <div class="mx-auto rounded-sm ring-1 ring-black/5 bg-white p-2" style="width: min(95vw, 100%); height: min(95vw, 100%);">
                    <div class="grid grid-cols-15 gap-[1px] w-full h-full animate-pulse" aria-hidden="true">
                        {#each Array(225) as _, i}
                            <div class="relative aspect-square w-full bg-emerald-50 border border-emerald-100 rounded-sm"></div>
                        {/each}
                    </div>
                </div>
            </div>
        </div>

        <!-- Rack placeholder -->
        <div class="px-3 flex-none mb-3">
            <div class="mx-auto max-w-[640px] rounded-2xl bg-white/90 backdrop-blur-md ring-1 ring-black/5 shadow-lg">
                <div class="px-2 pt-2 pb-2">
                    <div class="flex justify-center gap-1">
                        {#each Array(7) as _, i}
                            <div class="m-1 w-11 h-11 rounded-xl bg-amber-100 ring-1 ring-amber-300/60 animate-pulse"></div>
                        {/each}
                    </div>
                </div>
            </div>
        </div>

        <!-- Actions placeholder -->
        <div class="px-3 flex-none" style="padding-bottom: calc(env(safe-area-inset-bottom) + 12px);">
            <div class="mx-auto max-w-[640px]">
                <div class="w-full rounded-2xl bg-white/90 backdrop-blur-md shadow-lg ring-1 ring-black/5">
                    <div class="grid grid-cols-4 divide-x">
                        {#each Array(4) as _, i}
                            <div class="h-12 px-2 flex flex-col items-center justify-center">
                                <div class="w-6 h-6 rounded bg-gray-200 animate-pulse"></div>
                                <div class="mt-1 h-3 w-12 rounded bg-gray-200 animate-pulse"></div>
                            </div>
                        {/each}
                    </div>
                </div>
            </div>
        </div>
    </main>
    <span class="sr-only">Chargement…</span>
</div>

<style>
    :global(.grid-cols-15) {
        grid-template-columns: repeat(15, minmax(0, 1fr));
    }
</style>
