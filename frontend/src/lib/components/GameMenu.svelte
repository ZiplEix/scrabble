<script lang="ts">
    import { goto } from "$app/navigation";
    import type { Writable } from "svelte/store";
    import ChatRedirectionButton from "./ChatRedirectionButton.svelte";

    let { showScores, gameId }: {
        showScores: Writable<boolean>;
        gameId: string;
    } = $props()

    let menuOpen = $state(false);
    let root: HTMLElement;

    function closeMenu() {
        menuOpen = false;
    }

    function handleWindowClick(e: MouseEvent) {
        if (!menuOpen) return;
        const target = e.target as Node | null;
        if (root && target && !root.contains(target)) {
            closeMenu();
        }
    }

    function handleWindowKey(e: KeyboardEvent) {
        if (e.key === 'Escape' && menuOpen) closeMenu();
    }
</script>

<svelte:window on:click={handleWindowClick} on:keydown={handleWindowKey} />

<div class="relative px-3 flex items-center gap-2" bind:this={root}>
    <!-- chat redirection -->
    <ChatRedirectionButton gameId={gameId} />

    <!-- burger/ellipsis menu opener -->
    <button
        class="shrink-0 h-9 w-9 grid place-items-center rounded-full bg-emerald-600/90 hover:bg-emerald-600 text-white shadow-sm ring-1 ring-emerald-700/30"
        aria-label="Ouvrir le menu"
        onclick={() => (menuOpen = !menuOpen)}
        title="Menu"
    >
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" aria-hidden="true">
            <circle cx="5" cy="12" r="1.6" fill="currentColor" />
            <circle cx="12" cy="12" r="1.6" fill="currentColor" />
            <circle cx="19" cy="12" r="1.6" fill="currentColor" />
        </svg>
    </button>

    {#if menuOpen}
        <div
            class="absolute right-0 top-full mt-2 w-44 rounded-xl bg-white shadow-lg ring-1 ring-black/5 overflow-hidden z-20"
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
