<script lang="ts">
    import type { Writable } from "svelte/store";
    import { user } from "$lib/stores/user";
    import GameMenu from "./GameMenu.svelte";

    let {
        game,
        showScores,
        gameId,
    }: {
        game: GameInfo;
        showScores: Writable<boolean>;
        gameId: string;
    } = $props();

    // Runes dérivées pour l’état d’UI
    let isMyTurn = $derived(game.current_turn_username === $user?.username);
    let players = $derived([...game.players].sort((a, b) => b.score - a.score));
    let topScore = $derived(players.length ? players[0].score : 0);

    function initials(name: string) {
        if (!name) return "?";
        const parts = name.trim().split(/\s+/);
        const first = parts[0]?.[0] ?? "";
        const last = parts[1]?.[0] ?? "";
        return (first + last).toUpperCase() || name.slice(0, 2).toUpperCase();
    }
</script>

<header class="px-3 pt-2 pb-2">
    <!-- Ligne 1: retour, titre, menu -->
    <div class="flex items-center w-full justify-between gap-2">
        <div class="flex items-center gap-2 min-w-0">
            <a
                href="/"
                class="p-2 rounded-lg hover:bg-white/60 ring-1 ring-black/5 bg-white/40 backdrop-blur-sm"
                aria-label="Accueil"
            >
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" aria-hidden="true">
                    <path d="M4 12h16M10 6l-6 6 6 6" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
                </svg>
            </a>
            <h2
                class="text-base font-semibold text-gray-900 truncate"
                title={game.name}
            >
                {game.name}
            </h2>
        </div>

        <div class="flex items-center gap-2">
            <GameMenu {showScores} gameId={game.id} />
        </div>
    </div>

    <!-- Ligne 2: bandeau d’infos -->
    <div class="px-1 flex flex-col items-center gap-2 overflow-x-auto no-scrollbar pr-1">
        <!-- Indicateur de tour -->
        <div class="flex gap-2 mt-1">
            {#if isMyTurn}
                <span class="inline-flex items-center gap-1 px-2.5 py-1 rounded-full bg-emerald-100 text-[11px] text-emerald-800 ring-1 ring-emerald-300/60 whitespace-nowrap">
                    <span class="h-1.5 w-1.5 rounded-full bg-emerald-500"></span>
                    À vous de jouer
                </span>
            {:else}
                <span class="inline-flex items-center gap-1 px-2.5 py-1 rounded-full bg-amber-100 text-[11px] text-amber-800 ring-1 ring-amber-300/60 whitespace-nowrap">
                    Tour:<strong class="ml-0.5">{game.current_turn_username}</strong>
                </span>
            {/if}

            <!-- Lettres restantes -->
            <span class="inline-flex items-center gap-1 px-2.5 py-1 rounded-full bg-gray-100 text-[11px] text-gray-700 ring-1 ring-black/5 whitespace-nowrap">
                Restantes: <strong class="ml-1">{game.remaining_letters}</strong>
            </span>

            <span class="inline-flex items-center gap-1 px-2.5 py-1 rounded-full bg-gray-100 text-[11px] text-gray-700 ring-1 ring-black/5 whitespace-nowrap">
                Passes: <strong class="ml-1">{game.pass_count}/{game.players.length * 2}</strong>
            </span>
        </div>

        <!-- Mini-scoreboard -->
        <div class="flex items-center gap-2 mb-1">
            {#each players as p, i}
                <div
                    class="inline-flex items-center gap-1.5 px-2 py-1 rounded-full bg-white text-[11px] text-gray-800 ring-1 ring-black/5 shadow-sm whitespace-nowrap"
                    title={`${p.username} · ${p.score} pts`}
                >
                    <span class="relative h-6 w-6 shrink-0 grid place-items-center rounded-full bg-emerald-600 text-white text-[11px] font-semibold">
                        {initials(p.username)}
                        {#if i === 0 && p.score === topScore}
                            <span class="absolute -top-1 -right-1 text-[10px]">👑</span>
                        {/if}
                    </span>
                    <span class="tabular-nums font-semibold">{p.score}</span>
                </div>
            {/each}
        </div>
    </div>

    <div
        class="mt-2 h-[3px] rounded-full bg-gradient-to-r from-emerald-200 via-emerald-500/40 to-emerald-200"
    ></div>
</header>

<style>
    .no-scrollbar::-webkit-scrollbar {
        display: none;
    }
    .no-scrollbar {
        -ms-overflow-style: none;
        scrollbar-width: none;
    }
</style>
