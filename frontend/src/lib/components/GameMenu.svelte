<script lang="ts">
    import { goto } from "$app/navigation";
    import type { Writable } from "svelte/store";
    import ChatRedirectionButton from "./ChatRedirectionButton.svelte";

    let { showScores, gameId }: {
        showScores: Writable<boolean>;
        gameId: string;
    } = $props()

    let menuOpen = $state(false);

    function closeMenu() {
        menuOpen = false;
    }
</script>

<div class="px-3 flex gap-2">
    <!-- chat redirection -->
    <ChatRedirectionButton gameId={gameId} />

    <!-- burger menu opener -->
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
            onfocusout={(e) => {
                if (!(e.currentTarget as Node).contains(e.relatedTarget as Node)) closeMenu();
            }}
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
                class="w-full text-left px-3 py-2 text-sm hover:bg-gray-50 flex items-center justify-between"
                onclick={() => { goto(`/games/${gameId}/dictionnaire`); closeMenu(); }}
            >
                <span>📖 Dictionnaire</span>
            </button>
        </div>
    {/if}
</div>
