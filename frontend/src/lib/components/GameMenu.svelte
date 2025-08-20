<script lang='ts'>
  import { goto } from "$app/navigation";
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

<div class="flex items-center w-full justify-between">
    <!-- Title & infos -->
    <div class="w-full">
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

    <!-- Menu bouton -->
    <div class="px-3 relative flex gap-2">
        <button
				class="shrink-0 h-8 w-8 rounded-lg bg-amber-200 text-black flex items-center justify-center active:scale-95 transition"
				aria-label="Ouvrir le chat"
				onclick={() => goto(`/games/${gameId}/chat`)}
				title="Chat de la partie"
			>
				<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 24 24" fill="none" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 10h.01M12 10h.01M16 10h.01M21 12c0 4.418-4.03 8-9 8a9.77 9.77 0 0 1-4-.88L3 20l1.12-3.11A7.97 7.97 0 0 1 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
				</svg>
			</button>
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
                <button
                    class="block w-full text-left px-3 py-2 text-sm hover:bg-gray-50"
                    onclick={() => { goto(`/games/${gameId}/chat`); closeMenu(); }}
                >
                    💬 Chat
                </button>
            </div>
        {/if}
    </div>
</div>
