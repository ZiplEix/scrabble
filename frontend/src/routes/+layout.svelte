<script lang="ts">
    import '../app.css';
    import { page } from '$app/stores';
    import { user } from '$lib/stores/user';
    import { goto } from '$app/navigation';
    import { hideTabBar } from '$lib/stores/ui';

    const { children } = $props();

    // Cacher la Tab Bar si l'utilisateur n'est pas connecté OU s'il est dans une partie (/games/[id]) OU si cache demandé par l'UI
    let showTabBar = $derived(
        $user &&
        !($page.url.pathname.startsWith('/games/') && $page.url.pathname !== '/games/new') &&
        !$page.url.pathname.startsWith('/admin') &&
        !$hideTabBar
    );

    // Fonction pour vérifier si l'onglet est actif
    function isActive(path: string) {
        if (path === '/') {
            return $page.url.pathname === '/';
        }
        return $page.url.pathname.startsWith(path);
    }
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

<!-- Conteneur principal de l'app -->
<div class="min-h-100dvh flex flex-col pb-[env(safe-area-inset-bottom)] select-none">
    
    <!-- Contenu de la page -->
    <div class="flex-1 w-full {showTabBar ? 'pb-24' : ''}">
        {@render children()}
    </div>

    <!-- Tab Bar Mobile Premium -->
    {#if showTabBar}
        <nav 
            class="fixed bottom-0 left-0 right-0 z-40 bg-white/90 backdrop-blur-lg border-t border-stone-200/60 shadow-[0_-4px_24px_-4px_rgba(120,110,90,0.06)] px-4 pt-2 pb-[calc(env(safe-area-inset-bottom)+8px)]"
            aria-label="Navigation principale"
        >
            <div class="max-w-md mx-auto grid grid-cols-4 gap-1">
                
                <!-- Parties (Home) -->
                <a 
                    href="/" 
                    class="flex flex-col items-center justify-center py-1.5 rounded-xl transition-all active:scale-95 {isActive('/') ? 'text-brand-emerald font-bold' : 'text-stone-500 hover:text-stone-700'}"
                >
                    <svg class="w-6 h-6 mb-0.5 transition-transform {isActive('/') ? 'scale-110' : ''}" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6"/>
                    </svg>
                    <span class="text-[10px] tracking-wide">Parties</span>
                </a>

                <!-- Classement -->
                <a 
                    href="/leaderboard" 
                    class="flex flex-col items-center justify-center py-1.5 rounded-xl transition-all active:scale-95 {isActive('/leaderboard') ? 'text-brand-emerald font-bold' : 'text-stone-500 hover:text-stone-700'}"
                >
                    <svg class="w-6 h-6 mb-0.5 transition-transform {isActive('/leaderboard') ? 'scale-110' : ''}" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M9 12l2 2 4-4M7.835 4.697a3.42 3.42 0 001.946-.806 3.42 3.42 0 014.438 0 3.42 3.42 0 001.946.806 3.42 3.42 0 013.138 3.138 3.42 3.42 0 00.806 1.946 3.42 3.42 0 010 4.438 3.42 3.42 0 00-.806 1.946 3.42 3.42 0 01-3.138 3.138 3.42 3.42 0 00-1.946.806 3.42 3.42 0 01-4.438 0 3.42 3.42 0 00-1.946-.806 3.42 3.42 0 01-3.138-3.138 3.42 3.42 0 00-.806-1.946 3.42 3.42 0 010-4.438 3.42 3.42 0 00.806-1.946 3.42 3.42 0 013.138-3.138z"/>
                    </svg>
                    <span class="text-[10px] tracking-wide">Classement</span>
                </a>

                <!-- Défis -->
                <a 
                    href="/puzzles" 
                    class="flex flex-col items-center justify-center py-1.5 rounded-xl transition-all active:scale-95 {isActive('/puzzles') ? 'text-brand-emerald font-bold' : 'text-stone-500 hover:text-stone-700'}"
                >
                    <svg class="w-6 h-6 mb-0.5 transition-transform {isActive('/puzzles') ? 'scale-110' : ''}" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M11 4a2 2 0 114 0v1a1 1 0 001 1h3a1 1 0 011 1v3a1 1 0 01-1 1h-1a2 2 0 100 4h1a1 1 0 011 1v3a1 1 0 01-1 1h-3a1 1 0 01-1-1v-1a2 2 0 10-4 0v1a1 1 0 01-1 1H7a1 1 0 01-1-1v-3a1 1 0 00-1-1H4a2 2 0 110-4h1a1 1 0 001-1V7a1 1 0 011-1h3a1 1 0 001-1V4z"/>
                    </svg>
                    <span class="text-[10px] tracking-wide">Défis</span>
                </a>

                <!-- Mon Compte -->
                <a 
                    href="/me" 
                    class="flex flex-col items-center justify-center py-1.5 rounded-xl transition-all active:scale-95 {isActive('/me') ? 'text-brand-emerald font-bold' : 'text-stone-500 hover:text-stone-700'}"
                >
                    <svg class="w-6 h-6 mb-0.5 transition-transform {isActive('/me') ? 'scale-110' : ''}" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"/>
                    </svg>
                    <span class="text-[10px] tracking-wide">Profil</span>
                </a>

            </div>
        </nav>
    {/if}
</div>
