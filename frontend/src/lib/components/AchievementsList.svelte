<script lang="ts">
    import type { AchievementResponse } from "$lib/types/user_infos";
    
    let { achievements = [] }: { achievements?: AchievementResponse[] } = $props();
    
    // Trier : les succès débloqués en premier
    let sortedAchievements = $derived(
        [...achievements].sort((a, b) => (b.unlocked ? 1 : 0) - (a.unlocked ? 1 : 0))
    );
    
    let unlockedCount = $derived(
        achievements.filter(a => a.unlocked).length
    );
    
    let totalCount = $derived(achievements.length || 21);
    let progressPercent = $derived((unlockedCount / totalCount) * 100);
</script>

<div class="rounded-2xl bg-white/90 backdrop-blur-md ring-1 ring-black/5 shadow-lg p-5">
    <!-- En-tête de progression -->
    <div class="mb-6">
        <div class="flex justify-between items-center mb-2">
            <h2 class="text-lg font-bold text-stone-800">🏆 Succès & Badges</h2>
            <span class="text-sm font-black text-brand-emerald bg-brand-emerald-light/35 px-3 py-1 rounded-full">
                {unlockedCount} / {totalCount} débloqués
            </span>
        </div>
        <div class="w-full bg-stone-100 h-3 rounded-full overflow-hidden">
            <div 
                class="bg-gradient-to-r from-amber-400 to-brand-emerald h-full rounded-full transition-all duration-500 ease-out" 
                style="width: {progressPercent}%"
            ></div>
        </div>
    </div>

    <!-- Grille des Succès -->
    {#if achievements.length === 0}
        <p class="text-center text-stone-400 text-sm py-4">Aucun succès disponible pour le moment.</p>
    {:else}
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-3.5">
            {#each sortedAchievements as ach}
                <div 
                    class="relative flex items-center gap-4 p-4 rounded-2xl border transition-all duration-300 group
                    {ach.unlocked 
                        ? 'bg-gradient-to-br from-white to-stone-50/50 border-amber-200 shadow-sm hover:shadow-md hover:border-amber-300 hover:-translate-y-0.5' 
                        : 'bg-stone-50/50 border-stone-200/50 opacity-60 hover:opacity-85'
                    }"
                >
                    <!-- Badge Icon -->
                    <div 
                        class="w-14 h-14 rounded-2xl flex items-center justify-center text-3xl shrink-0 transition-transform group-hover:scale-105
                        {ach.unlocked 
                            ? 'bg-gradient-to-br from-amber-100 to-amber-200/50 border border-amber-300 shadow-sm text-amber-600' 
                            : 'bg-stone-200 border border-stone-300 text-stone-400'
                        }"
                    >
                        {#if ach.unlocked}
                            {ach.badge_icon || '🏆'}
                        {:else}
                            🔒
                        {/if}
                    </div>

                    <!-- Infos Succès -->
                    <div class="flex-1 min-w-0">
                        <div class="flex items-center gap-1.5 mb-0.5">
                            <h4 class="font-extrabold text-sm text-stone-800 truncate">
                                {ach.title}
                            </h4>
                            {#if ach.unlocked}
                                <span class="w-2 h-2 rounded-full bg-amber-400 animate-ping absolute top-4 right-4"></span>
                                <span class="w-2 h-2 rounded-full bg-amber-400 absolute top-4 right-4"></span>
                            {/if}
                        </div>
                        <p class="text-xs text-stone-500 leading-normal line-clamp-2">
                            {ach.description}
                        </p>
                        {#if ach.unlocked && ach.unlocked_at}
                            <span class="text-[10px] text-brand-emerald font-bold mt-1.5 block">
                                Déverrouillé le {new Date(ach.unlocked_at).toLocaleDateString('fr-FR', { dateStyle: 'short' })}
                            </span>
                        {/if}
                    </div>
                </div>
            {/each}
        </div>
    {/if}
</div>
