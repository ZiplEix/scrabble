<script lang="ts">
    import type { HeadToHeadInfo } from "$lib/types/user_infos";
    
    let { headToHead, opponentName }: { headToHead: HeadToHeadInfo; opponentName: string } = $props();
    
    // Pourcentages de victoires pour le graphique en barre VS
    let totalWins = $derived(headToHead.user_wins + headToHead.opponent_wins);
    let userWinPercent = $derived(
        totalWins > 0 ? (headToHead.user_wins / totalWins) * 100 : 50
    );
    let oppWinPercent = $derived(
        totalWins > 0 ? (headToHead.opponent_wins / totalWins) * 100 : 50
    );
</script>

<div class="rounded-2xl bg-white/90 backdrop-blur-md ring-1 ring-black/5 shadow-lg p-5 flex flex-col gap-6">
    <!-- En-tête VS -->
    <div class="text-center">
        <h2 class="text-lg font-bold text-stone-800 mb-1">⚔️ Face-à-Face direct</h2>
        <p class="text-xs text-stone-400">Historique complet de vos affrontements</p>
    </div>

    <!-- Panel Face-à-Face global -->
    <div class="grid grid-cols-3 items-center gap-4 bg-stone-50/50 p-4 border border-stone-200/40 rounded-2xl">
        <div class="text-center">
            <span class="text-[10px] uppercase tracking-wider text-stone-400 font-bold block mb-1">Moi</span>
            <span class="text-3xl font-black text-brand-emerald block">{headToHead.user_wins}</span>
            <span class="text-[10px] text-stone-400 font-bold block mt-1">victoire{headToHead.user_wins > 1 ? 's' : ''}</span>
        </div>
        <div class="text-center">
            <div class="inline-flex items-center justify-center w-11 h-11 rounded-full bg-stone-200 border-2 border-white shadow-sm font-black text-stone-600 text-xs">
                {headToHead.games_played}
            </div>
            <span class="text-[10px] tracking-wide text-stone-400 font-bold block mt-1.5">Matchs joués</span>
        </div>
        <div class="text-center">
            <span class="text-[10px] uppercase tracking-wider text-stone-400 font-bold block mb-1">{opponentName}</span>
            <span class="text-3xl font-black text-rose-500 block">{headToHead.opponent_wins}</span>
            <span class="text-[10px] text-stone-400 font-bold block mt-1">victoire{headToHead.opponent_wins > 1 ? 's' : ''}</span>
        </div>
    </div>

    <!-- Barre VS visuelle -->
    {#if totalWins > 0}
        <div class="flex flex-col gap-1.5">
            <div class="flex justify-between text-[11px] font-black text-stone-500 px-1">
                <span>{Math.round(userWinPercent)}%</span>
                <span>{Math.round(oppWinPercent)}%</span>
            </div>
            <div class="w-full h-3 rounded-full flex overflow-hidden border border-white shadow-inner">
                <div class="bg-brand-emerald h-full transition-all" style="width: {userWinPercent}%"></div>
                <div class="bg-rose-500 h-full transition-all" style="width: {oppWinPercent}%"></div>
            </div>
        </div>
    {/if}

    <!-- Statistiques moyennes -->
    <div class="grid grid-cols-2 gap-4">
        <!-- Moyenne Moi -->
        <div class="bg-gradient-to-br from-emerald-50/50 to-white p-3.5 border border-emerald-100 rounded-2xl flex flex-col gap-0.5">
            <span class="text-[10px] text-stone-400 font-extrabold uppercase">Moyenne (Moi)</span>
            <span class="text-xl font-black text-brand-emerald">{Math.round(headToHead.user_avg_score)} pts</span>
        </div>
        <!-- Moyenne Adversaire -->
        <div class="bg-gradient-to-br from-rose-50/50 to-white p-3.5 border border-rose-100 rounded-2xl flex flex-col gap-0.5">
            <span class="text-[10px] text-stone-400 font-extrabold uppercase">Moyenne ({opponentName})</span>
            <span class="text-xl font-black text-rose-500">{Math.round(headToHead.opp_avg_score)} pts</span>
        </div>
    </div>

    <!-- Historique des parties communes -->
    <div>
        <h3 class="text-xs font-black text-stone-400 uppercase tracking-wider mb-3">5 derniers affrontements</h3>
        {#if !headToHead.recent_games || headToHead.recent_games.length === 0}
            <p class="text-center text-stone-400 text-xs py-4">Aucune partie terminée trouvée ensemble.</p>
        {:else}
            <div class="flex flex-col gap-2.5">
                {#each headToHead.recent_games as game}
                    {@const didUserWin = game.status === 'ended' && game.winner === game.name} // wait winner will be username
                    <div class="flex items-center justify-between p-3 bg-stone-50/40 border border-stone-200/30 rounded-2xl">
                        <div class="flex flex-col gap-0.5">
                            <span class="text-xs font-bold text-stone-800 truncate max-w-[150px]">{game.name || 'Partie Amicale'}</span>
                            <span class="text-[10px] text-stone-400">
                                {new Date(game.created_at).toLocaleDateString('fr-FR', { dateStyle: 'short' })}
                            </span>
                        </div>

                        <!-- Scores -->
                        <div class="flex items-center gap-3">
                            <div class="flex flex-col items-end">
                                <span class="text-xs font-black text-brand-emerald">{game.user_score} <span class="text-[9px] font-normal text-stone-400">pts</span></span>
                                <span class="text-[9px] font-extrabold text-stone-400">Moi</span>
                            </div>
                            <span class="text-stone-300 font-bold text-sm">:</span>
                            <div class="flex flex-col items-start">
                                <span class="text-xs font-black text-rose-500">{game.opp_score} <span class="text-[9px] font-normal text-stone-400">pts</span></span>
                                <span class="text-[9px] font-extrabold text-stone-400">{opponentName}</span>
                            </div>
                        </div>

                        <!-- Statut -->
                        <div class="text-right">
                            {#if game.status === 'ended'}
                                <span class="text-[10px] font-black uppercase px-2 py-0.5 rounded-full bg-stone-100 text-stone-500">
                                    Terminé
                                </span>
                            {:else}
                                <span class="text-[10px] font-black uppercase px-2 py-0.5 rounded-full bg-amber-100 text-amber-600 animate-pulse">
                                    En cours
                                </span>
                            {/if}
                        </div>
                    </div>
                {/each}
            </div>
        {/if}
    </div>
</div>
