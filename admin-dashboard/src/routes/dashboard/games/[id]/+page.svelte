<script lang="ts">
    import { onMount } from 'svelte';
    import { page } from '$app/stores';
    import { api } from '$lib/api';
    import { derived } from 'svelte/store';

    type BoardBlank = { x: number; y: number };
    type PlayerInfo = { id: number; username: string; score: number; position: number; rack?: string };
    type MoveInfo = { player_id: number; move: any; played_at: string };
    type GameInfo = {
        id: string;
        name: string;
        board: any;
        players: PlayerInfo[];
        moves: MoveInfo[];
        current_turn: number;
        current_turn_username: string;
        status: string;
        remaining_letters: number;
        winner_username?: string;
        ended_at?: string;
        is_your_game?: boolean;
        blank_tiles?: BoardBlank[];
        pass_count: number;
        available_letters?: string;
    };

    let game: GameInfo | null = null;
    let loading = true;
    let errorMsg = '';
    let cellSize = 24; // px, contrôlera la taille de case

    const idStore = derived(page, ($page) => $page.params.id);
    let id: string;
    const unsubscribe = idStore.subscribe((v) => (id = v ?? ""));

    async function fetchGame() {
        errorMsg = '';
        try {
            const res = await api.get(`/admin/game/${id}`);
            game = res.data?.game as GameInfo;
        } catch (err) {
            console.error('failed to fetch admin game', err);
            errorMsg = 'Erreur lors du chargement de la partie';
            game = null;
        }
    }

    onMount(async () => {
        try {
            await fetchGame();
        } finally {
            loading = false;
        }
    });

    function cellValue(x: number, y: number): string {
        if (!game?.board) return '';
        // board peut être un array 2D de strings
        const v = game.board?.[y]?.[x];
        return typeof v === 'string' ? v : '';
    }

    function isBlankTile(x: number, y: number): boolean {
        if (!game?.blank_tiles) return false;
        return game.blank_tiles.some((b) => b.x === x && b.y === y);
    }

  // Helpers pour le rendu des coups
    function isPassMove(m: any): boolean {
        return m && typeof m === 'object' && m.type === 'pass';
    }
    function isWordMove(m: any): boolean {
        return (
            m && typeof m === 'object' && Array.isArray(m.letters) && typeof m.word === 'string'
        );
    }
    function dirArrow(dir: string | undefined): string {
        if (!dir) return '';
        return dir === 'H' ? '→' : dir === 'V' ? '↓' : dir;
    }

    // Comptage des lettres restantes à partir de la chaîne available_letters
    function remainingLettersCount(): Array<{char: string; count: number}> {
        if (!game?.available_letters) return [];
        const counts = new Map<string, number>();
        for (const ch of game.available_letters) {
            const key = ch.toUpperCase();
            counts.set(key, (counts.get(key) || 0) + 1);
        }
        // Tri alphabétique avec '?' en fin
        return Array.from(counts.entries())
            .map(([char, count]) => ({ char, count }))
            .sort((a, b) => {
                if (a.char === '?' && b.char !== '?') return 1;
                if (b.char === '?' && a.char !== '?') return -1;
                return a.char.localeCompare(b.char);
            });
    }
</script>

<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <header class="mb-6 flex items-start justify-between gap-4">
        <div>
            <h1 class="text-2xl font-bold flex items-center gap-3">
                {#if game}
                    <span>{game.name}</span>
                    {#if game.status === 'ended'}
                        <span class="inline-block px-2 py-0.5 rounded-full text-xs font-medium bg-emerald-600 text-white">FINIE</span>
                    {:else if game.status === 'ongoing'}
                        <span class="inline-block px-2 py-0.5 rounded-full text-xs font-medium bg-sky-600 text-white">EN COURS</span>
                    {:else}
                        <span class="inline-block px-2 py-0.5 rounded-full text-xs font-medium bg-zinc-600 text-white">{game.status}</span>
                    {/if}
                {:else}
                    <span>Partie</span>
                {/if}
            </h1>
            <p class="text-sm text-white/70 mt-1">Détails complets de la partie</p>
        </div>
        <a href="/dashboard/games" class="px-3 py-1.5 bg-white/6 rounded hover:bg-white/8 text-sm text-white flex items-center gap-2" aria-label="Retour aux parties">
            <span>←</span>
            <span>Retour aux parties</span>
        </a>
    </header>

    {#if loading}
            <div class="text-center text-white/70">Chargement...</div>
    {:else if errorMsg}
            <div class="text-center text-red-400">{errorMsg}</div>
    {:else if !game}
            <div class="text-center text-white/70">Partie introuvable</div>
    {:else}

        <!-- Résumé -->
        <section class="mb-6 bg-white/4 rounded-lg p-4">
            <div class="grid grid-cols-1 md:grid-cols-3 gap-3 text-sm">
                <div><span class="text-white/60">ID:</span> <span class="text-white">{game.id}</span></div>
                <div><span class="text-white/60">Tour courant:</span> <span class="text-white">{game.current_turn_username || '—'}</span></div>
                <div><span class="text-white/60">Lettres restantes:</span> <span class="text-white">{game.remaining_letters}</span></div>
                <div><span class="text-white/60">Pass count:</span> <span class="text-white">{game.pass_count}</span></div>
                <div><span class="text-white/60">Gagnant:</span> <span class="text-white">{game.winner_username || '—'}</span></div>
                {#if game.ended_at}
                    <div><span class="text-white/60">Fin:</span> <span class="text-white">{new Date(game.ended_at).toLocaleString()}</span></div>
                {/if}
            </div>
            {#if game.available_letters}
            <div class="mt-3">
                <h3 class="text-sm font-semibold mb-2">Répartition des lettres restantes</h3>
                <div class="flex flex-wrap gap-1">
                    {#each remainingLettersCount() as it}
                        <span class="inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-xs bg-white/10 text-white">
                            <span class="font-semibold">{it.char}</span>
                            <span class="text-white/70">×{it.count}</span>
                        </span>
                    {/each}
                </div>
            </div>
            {/if}
        </section>

        <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <!-- Joueurs -->
            <section class="bg-white/4 rounded-lg p-4">
                <h2 class="font-semibold mb-3">Joueurs</h2>
                <table class="w-full text-sm">
                <thead>
                    <tr class="text-left text-xs text-white/60">
                    <th class="px-2 py-1">Ordre</th>
                    <th class="px-2 py-1">Username</th>
                    <th class="px-2 py-1">Score</th>
                    <th class="px-2 py-1">Rack</th>
                    </tr>
                </thead>
                <tbody>
                    {#each game.players as p}
                        <tr class="border-t border-white/6">
                            <td class="px-2 py-1">{p.position}</td>
                            <td class="px-2 py-1">{p.username}</td>
                            <td class="px-2 py-1">{p.score}</td>
                            <td class="px-2 py-1">
                                {#if p.rack}
                                    <div class="flex flex-wrap gap-1">
                                        {#each p.rack.split('') as ch, idx}
                                            <span class={`inline-flex items-center justify-center w-5 h-6 text-xs rounded ${ch==='?'?'bg-amber-600 text-black':'bg-white/10 text-white'}`} title={`Lettre ${ch}`}>
                                                {ch}
                                            </span>
                                        {/each}
                                    </div>
                                {:else}
                                    <span class="text-white/40">—</span>
                                {/if}
                            </td>
                        </tr>
                    {/each}
                </tbody>
                </table>
            </section>

            <!-- Plateau -->
            <section class="bg-white/4 rounded-lg p-4">
                <h2 class="font-semibold mb-3">Plateau</h2>
                <div class="board" style={`--cell: ${cellSize}px`}>
                    {#each Array(15) as _, y}
                        {#each Array(15) as __, x}
                            {#key `${x}-${y}`}
                                <div class={`cell ${isBlankTile(x,y)?'bg-amber-600 text-black':'bg-white/10 text-white'}`}
                                    title={`(${x},${y})`}
                                >
                                    {cellValue(x,y)}
                                </div>
                            {/key}
                        {/each}
                    {/each}
                </div>
            </section>
        </div>

        <!-- Historique des coups -->
        <section class="bg-white/4 rounded-lg p-4 mt-6">
            <h2 class="font-semibold mb-3">Historique des coups</h2>
            <table class="w-full text-sm">
                <thead>
                    <tr class="text-left text-xs text-white/60">
                        <th class="px-2 py-1">Joueur</th>
                        <th class="px-2 py-1">Date</th>
                        <th class="px-2 py-1">Coup</th>
                    </tr>
                </thead>
                <tbody>
                    {#each game.moves as mv}
                        <tr class="border-t border-white/6 align-top">
                            <td class="px-2 py-1">{(game.players.find(p=>p.id===mv.player_id)?.username) || mv.player_id}</td>
                            <td class="px-2 py-1">{new Date(mv.played_at).toLocaleString()}</td>
                            <td class="px-2 py-1">
                                {#if isPassMove(mv.move)}
                                    <span class="inline-block px-2 py-0.5 rounded-full text-xs font-medium bg-zinc-600 text-white">Passe</span>
                                {:else if isWordMove(mv.move)}
                                <div>
                                    <div class="flex items-center gap-3">
                                        <span class="font-semibold text-white">{mv.move.word}</span>
                                        <span class="text-white/70">{dirArrow(mv.move.dir)} ({mv.move.x},{mv.move.y})</span>
                                        {#if typeof mv.move.score === 'number'}
                                            <span class="text-emerald-400 font-medium">+{mv.move.score} pts</span>
                                        {/if}
                                    </div>
                                    <div class="mt-1 flex flex-wrap gap-1">
                                        {#each mv.move.letters as l}
                                            <span
                                                class={`inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-xs ${l?.blank ? 'bg-amber-600 text-black' : 'bg-white/10 text-white'}`}
                                                title={`(${l?.x},${l?.y})`}
                                            >
                                                <span class="font-semibold">{l?.char}</span>
                                                <span class="text-white/60 text-[10px]">({l?.x},{l?.y})</span>
                                                {#if l?.blank}
                                                    <span class="text-[10px]">joker</span>
                                                {/if}
                                            </span>
                                        {/each}
                                    </div>
                                    <details class="mt-2">
                                        <summary class="cursor-pointer text-xs text-white/60">Voir JSON</summary>
                                        <pre class="mt-1 font-mono text-xs whitespace-pre-wrap bg-white/5 p-2 rounded">{JSON.stringify(mv.move, null, 2)}</pre>
                                    </details>
                                </div>
                                {:else}
                                    <pre class="font-mono text-xs whitespace-pre-wrap bg-white/5 p-2 rounded">{JSON.stringify(mv.move, null, 2)}</pre>
                                {/if}
                            </td>
                        </tr>
                    {/each}
                    {#if game.moves.length === 0}
                        <tr><td colspan="3" class="px-2 py-2 text-white/60">Aucun coup joué</td></tr>
                    {/if}
                </tbody>
            </table>
        </section>

        <div class="mt-4 flex items-center gap-4">
            <button class="px-3 py-1.5 bg-white/6 rounded hover:bg-white/8" on:click={fetchGame}>Rafraîchir</button>
            <label class="text-sm text-white/70 flex items-center gap-2">
                Taille des cases
                <input type="range" min="16" max="40" step="1" bind:value={cellSize} />
                <span class="text-white/80 text-xs">{cellSize}px</span>
            </label>
        </div>
    {/if}
</div>

<style>
    /* Plateau fixe: colonnes & rangées de taille constante pour éviter l'étirement plein écran */
    .board {
        /* taille d'une case contrôlée via style inline (fallback ci-dessous) */
        --cell: 22px;
        --gap: 2px;
        display: grid;
        grid-template-columns: repeat(15, var(--cell));
        grid-auto-rows: var(--cell);
        gap: var(--gap);
        width: fit-content;
        /* Empêche la grille de s'étirer pour remplir le parent */
        justify-content: start;
        overflow: auto;
    }
    .cell {
        width: var(--cell);
        height: var(--cell);
        display: flex;
        align-items: center;
        justify-content: center;
        font-size: 0.75rem; /* équiv. text-xs */
        border-radius: 0.25rem; /* équiv. rounded */
        user-select: none;
    }
    /* Ancienne classe conservée si utilisée ailleurs */
    .grid-cols-15 { grid-template-columns: repeat(15, minmax(0, 1fr)); }
    @media (min-width: 1024px) {
        /* Légère augmentation de la taille des cases sur grands écrans sans casser la grille */
        .board { --cell: clamp(18px, 2.2vw, 34px); }
    }
</style>
