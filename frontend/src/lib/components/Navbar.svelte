<script lang='ts'>
    import { goto } from '$app/navigation';
    import { user } from '$lib/stores/user';
  import { onMount } from 'svelte';

    let navE1: HTMLElement;

    function setNavVar() {
        const h = navE1?.offsetHeight ?? 72;
        document.documentElement.style.setProperty('--nav-h', `${h}px`);
    }

    onMount(() => {
        setNavVar();
        const ro = new ResizeObserver(setNavVar);
        if (navE1) ro.observe(navE1);
        window.addEventListener('resize', setNavVar);
        return () => {
            ro.disconnect();
            window.removeEventListener('resize', setNavVar);
        };
    });
</script>

<nav bind:this={navE1} class="flex justify-between items-center px-6 py-4 bg-white shadow-md">
	<a href="/" class="text-2xl font-bold text-green-700 tracking-tight">🧩 Scrabble</a>
	<div class="flex space-x-6 items-center">
		{#if !$user}
			<a href="/login" class="bg-green-600 text-white px-4 py-2 rounded hover:bg-green-700">Connexion</a>
        {:else}
            <button
                class="bg-green-600 text-white px-4 py-2 rounded hover:bg-green-700"
                on:click={() => goto('/me')}
            >
                Mon profil
            </button>
		{/if}
	</div>
</nav>

<style>
    :root { --nav-h: 72px; } /* fallback */
</style>
