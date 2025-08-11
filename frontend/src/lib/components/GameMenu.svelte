<script lang='ts'>
  import type { Writable } from "svelte/store";

    let { game, showScores, gameId }: {
        game: GameInfo;
        showScores: Writable<boolean>;
        gameId: string;
    } = $props()

    let menuOpen = $state(false);
  	function closeMenu() {
        menuOpen = false;
    }
</script>

<div class="flex items-center justify-between">
    <div>
        <div class="px-3 pt-2 pb-1 w-full relative">
            <div class="flex items-center gap-2">
                <h2 class="flex-1 text-base font-semibold text-gray-800 truncate">{game.name}</h2>
            </div>
        </div>
        <!-- Sous-ligne compacte (hauteur minimale) -->
        <p class="px-3 mt-0.5 text-[11px] leading-tight text-gray-600 flex items-center justify-between">
            <span>
                Lettres restantes : <strong class="font-semibold">{game.remaining_letters}</strong>
            </span>
            <span>
                Tour : <strong class="font-semibold">{game.current_turn_username}</strong>
            </span>
        </p>
    </div>
    <div class="px-3 relative">
        <button
            class="shrink-0 h-8 w-8 grid place-items-center rounded-lg bg-gray-100 hover:bg-gray-200 text-xl leading-none"
            aria-label="Ouvrir le menu"
            onclick={() => (menuOpen = !menuOpen)}
        >
            ⋯
        </button>

        {#if menuOpen}
            <div
                class="absolute right-3 mt-1 w-44 rounded-xl bg-white shadow-lg ring-1 ring-black/5 overflow-hidden z-20"
                onfocusout={(e) => { if (!e.currentTarget.contains(e.relatedTarget)) closeMenu(); }}
            >
                <button
                    class="w-full text-left px-3 py-2 text-sm hover:bg-gray-50"
                    onclick={() => { showScores.set(true); closeMenu(); }}
                >
                    🏆 Classement
                </button>
                <a
                    href="/report"
                    class="block px-3 py-2 text-sm hover:bg-gray-50"
                    onclick={closeMenu}
                >
                    🛠️ Reporter un bug
                </a>
                <a
                    href={`/games/${gameId}/history`}
                    class="block px-3 py-2 text-sm hover:bg-gray-50"
                    onclick={closeMenu}
                >
                    📜 Historique
                </a>
            </div>
        {/if}
    </div>
</div>
