<script lang="ts">
    import Navbar from '$lib/components/Navbar.svelte';
    import { onMount } from 'svelte';
    import '../app.css';
    import { registerSW } from '$lib/serviceWorker';
    import { browser } from '$app/environment';

    let { children } = $props();

    onMount(() => {
        if (browser) {
            registerSW(() => {
                if (confirm("Une nouvelle version est disponible. Voulez-vous recharger ?")) {
                    navigator.serviceWorker.getRegistration().then(reg => {
                        if (reg?.waiting) {
                            reg.waiting.postMessage({ type: 'SKIP_WAITING' });
                        }
                    });
                }
            });
        }
    });
</script>

<svelte:head>
    <title>Scrabble Online - Joue avec tes amis</title>

    <!-- Open Graph -->
    <meta property="og:title" content="Scrabble Online" />
    <meta property="og:description" content="Un Scrabble en ligne simple et rapide à jouer entre amis." />
    <meta property="og:image" content="https://scrabble.baptiste.zip/og-image.png" />
    <meta property="og:url" content="https://scrabble.baptiste.zip/" />
    <meta property="og:type" content="website" />

    <!-- Twitter Cards -->
    <meta name="twitter:card" content="summary_large_image" />
    <meta name="twitter:title" content="Scrabble Online" />
    <meta name="twitter:description" content="Un Scrabble en ligne simple et rapide à jouer entre amis." />
    <meta name="twitter:image" content="https://scrabble.baptiste.zip/og-image.png" />
</svelte:head>

<!-- NAVBAR -->
<Navbar />

{@render children()}
