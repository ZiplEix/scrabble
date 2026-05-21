<script lang="ts">
    import UserLink from "$lib/components/UserLink.svelte";
    import type { GameInfo } from "$lib/types/game_infos";

    let { game }: { game: GameInfo } = $props();

    function getUsername(pid: number) {
        return game.players.find(p => p.id === pid)?.username ?? '–';
    }

    function formatDate(d: string) {
        const date = new Date(d);
        return date.toLocaleTimeString('fr-FR', { hour: '2-digit', minute: '2-digit' });
    }

    function formatFullDate(d: string) {
        const date = new Date(d);
        return date.toLocaleDateString('fr-FR');
    }
</script>

<div class="flex-1 overflow-y-auto px-4 py-4 space-y-3 no-scrollbar bg-stone-50/30">
    <div class="max-w-lg mx-auto">
        {#each [...game.moves].reverse() as move, idx}
            <!-- Reverse list so the latest move is at the very top! (Extremely practical on mobile) -->
            {@const originalIdx = game.moves.length - 1 - idx}
            <article class="rounded-2xl border border-stone-200/60 bg-white shadow-sm p-4 space-y-3 relative overflow-hidden group">
                
                <!-- Header of item: index + action badge + points -->
                <div class="flex items-center justify-between gap-2">
                    <div class="flex items-center gap-2 min-w-0">
                        <span class="inline-flex items-center px-2 py-0.5 rounded-full bg-stone-100 text-[10px] font-bold text-stone-600 border border-stone-200/20">
                            #{originalIdx + 1}
                        </span>
                        {#if !move.move.type}
                            <span class="font-extrabold text-stone-800 truncate text-sm" title={move.move.word}>
                                {move.move.word}
                            </span>
                            <span class="inline-flex items-center px-1.5 py-0.5 rounded-md bg-brand-emerald-light text-brand-emerald text-[9px] font-bold uppercase tracking-wider">
                                {move.move.dir}
                            </span>
                        {/if}
                    </div>
                    
                    <div>
                        {#if move.move.score}
                            <span class="inline-flex items-center gap-0.5 px-2.5 py-1 rounded-full bg-brand-emerald-light text-brand-emerald text-[10px] font-bold tabular-nums border border-brand-emerald/10 shadow-sm">
                                +{move.move.score} pts
                            </span>
                        {:else}
                            <span class="inline-flex items-center px-2.5 py-1 rounded-full bg-amber-50 text-brand-gold-hover text-[10px] font-semibold border border-brand-gold/10">
                                Passé
                            </span>
                        {/if}
                    </div>
                </div>

                <!-- Meta: Player info + Timestamp -->
                <div class="flex flex-wrap items-center gap-2 text-[10px] text-stone-500 font-medium border-t border-stone-100 pt-2.5">
                    <span class="inline-flex items-center gap-1 bg-stone-50 rounded-full px-2.5 py-0.5 border border-stone-200/40 text-stone-600">
                        Joué par <strong class="text-stone-700"><UserLink id={move.player_id} username={getUsername(move.player_id)} /></strong>
                    </span>
                    <span class="inline-flex items-center gap-1 bg-stone-50 rounded-full px-2.5 py-0.5 border border-stone-200/40">
                        ⏱️ {formatDate(move.played_at)} ({formatFullDate(move.played_at)})
                    </span>
                </div>

                <!-- Tile letters representation -->
                {#if move.move.letters?.length}
                    <div class="flex flex-wrap gap-1 mt-2.5 pl-0.5">
                        {#each move.move.letters as l}
                            <div 
                                class="grid place-items-center w-7 h-7 rounded-lg bg-gradient-to-br from-amber-50 to-brand-gold-light text-stone-800 font-extrabold text-[12px] border border-brand-gold/30 shadow-sm relative group select-none"
                                title={`Lettre posée en (${l.x}, ${l.y})`}
                            >
                                <span class="leading-none">{l.char}</span>
                            </div>
                        {/each}
                    </div>
                {/if}
                
            </article>
        {/each}

        {#if game.moves.length === 0}
            <div class="flex flex-col items-center justify-center text-center py-12 select-none">
                <span class="text-3xl mb-2">📜</span>
                <p class="text-xs font-bold text-stone-500">Aucun coup joué pour l'instant</p>
                <p class="text-[10px] text-stone-400 mt-0.5">Les coups de la partie s'afficheront ici au fur et à mesure.</p>
            </div>
        {/if}
    </div>
</div>
