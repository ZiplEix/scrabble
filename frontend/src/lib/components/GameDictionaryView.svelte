<script lang="ts">
    import { onMount } from "svelte";
    import { extractDefinitions, getDefinition, type DefinitionGroup, type WiktionaryDefinition } from "$lib/utils/get_definition";
    import type { GameInfo } from "$lib/types/game_infos";

    let { game }: { game: GameInfo } = $props();

    type Definition = {
        word: string;
        url: string;
        wikidef: WiktionaryDefinition[];
        def: DefinitionGroup[];
    };

    let words = $state<string[]>([]);
    let definitions = $state<Definition[]>([]);
    let loading = $state(true);
    let error = $state('');

    // Map (x,y) -> timestamp of placement (ms)
    function buildTilePlacementMap(moves: any[] | undefined): Map<string, number> {
        const map = new Map<string, number>();
        if (!moves) return map;
        const sorted = [...moves].sort(
            (a, b) => new Date(a.played_at).getTime() - new Date(b.played_at).getTime()
        );
        for (const m of sorted) {
            const t = new Date(m.played_at).getTime();
            for (const l of m?.move?.letters ?? []) {
                const key = `${l.x},${l.y}`;
                if (!map.has(key)) map.set(key, t);
            }
        }
        return map;
    }

    // Extract words (length >= 2) with coordinate cells
    function extractWordSpans(board: string[][]): { word: string; cells: { x: number; y: number }[] }[] {
        const H = board.length;
        const W = H > 0 ? board[0].length : 0;
        const isLetter = (c: string) => /^[A-Z]$/.test(c);
        const spans: { word: string; cells: { x: number; y: number }[] }[] = [];

        // Horizontal
        for (let y = 0; y < H; y++) {
            let x = 0;
            while (x < W) {
                if (isLetter(board[y][x]) && (x === 0 || !isLetter(board[y][x - 1]))) {
                    let w = '';
                    const cells: { x: number; y: number }[] = [];
                    let xi = x;
                    while (xi < W && isLetter(board[y][xi])) {
                        w += board[y][xi];
                        cells.push({ x: xi, y });
                        xi++;
                    }
                    if (w.length >= 2) spans.push({ word: w.toUpperCase(), cells });
                    x = xi;
                } else {
                    x++;
                }
            }
        }

        // Vertical
        for (let x = 0; x < W; x++) {
            let y = 0;
            while (y < H) {
                if (isLetter(board[y][x]) && (y === 0 || !isLetter(board[y - 1][x]))) {
                    let w = '';
                    const cells: { x: number; y: number }[] = [];
                    let yi = y;
                    while (yi < H && isLetter(board[yi][x])) {
                        w += board[yi][x];
                        cells.push({ x, y: yi });
                        yi++;
                    }
                    if (w.length >= 2) spans.push({ word: w.toUpperCase(), cells });
                    y = yi;
                } else {
                    y++; // Corrected: increment y instead of x!
                }
            }
        }
        return spans;
    }

    // Sort words by latest placement timestamp
    function getWordsOrderedByPlacement(board: string[][], moves: any[] | undefined): string[] {
        const spans = extractWordSpans(board);
        const tileTime = buildTilePlacementMap(moves);
        const wordMaxTime = new Map<string, number>();
        for (const s of spans) {
            let maxT = -Infinity;
            for (const c of s.cells) {
                const t = tileTime.get(`${c.x},${c.y}`);
                if (typeof t === 'number' && t > maxT) maxT = t;
            }
            const prev = wordMaxTime.get(s.word);
            if (prev === undefined || (maxT > prev)) wordMaxTime.set(s.word, maxT);
        }
        const uniqueWords = Array.from(wordMaxTime.keys());
        uniqueWords.sort((a, b) => (wordMaxTime.get(b)! - wordMaxTime.get(a)!));
        return uniqueWords;
    }

    onMount(async () => {
        try {
            const ordered = getWordsOrderedByPlacement(game.board, game.moves);
            words = ordered;
            loading = false;
        } catch (e: any) {
            error = 'Impossible de lire les mots de la grille';
            loading = false;
            return;
        }

        if (words.length) {
            // Load definitions sequentially to preserve active ordering
            const defs: Definition[] = [];
            for (const word of words) {
                let newDef: Definition = {
                    word,
                    url: `https://fr.wiktionary.org/wiki/${encodeURIComponent(word)}`,
                    wikidef: [],
                    def: []
                };
                
                try {
                    const wikidef = await getDefinition(word);
                    newDef.wikidef = wikidef ? [wikidef] : [];
                    if (wikidef?.url) {
                        newDef.url = wikidef.url;
                    }
                    if (wikidef && wikidef.extract) {
                        newDef.def = extractDefinitions(wikidef.extract);
                    }
                } catch (err) {
                    console.warn(`failed to fetch definition for ${word}`, err);
                }
                defs.push(newDef);
                definitions = [...defs]; // update state incrementally for faster initial visual loading!
            }
        }
    });
</script>

<div class="flex-1 overflow-y-auto px-4 py-4 space-y-4 no-scrollbar bg-stone-50/30">
    <div class="max-w-lg mx-auto">
        <h2 class="text-xs font-bold uppercase tracking-wider text-stone-400 text-center mb-4 select-none">
            📚 Dictionnaire de la Grille
        </h2>

        {#if loading}
            <div class="flex flex-col items-center justify-center py-12 gap-3">
                <svg class="animate-spin text-brand-emerald w-8 h-8" fill="none" stroke="currentColor" stroke-width="3" viewBox="0 0 24 24">
                    <circle cx="12" cy="12" r="10" stroke="rgba(12,106,77,0.15)" stroke-width="3"></circle>
                    <path d="M22 12a10 10 0 0 1-10 10" stroke="currentColor" stroke-width="3" stroke-linecap="round"></path>
                </svg>
                <p class="text-xs font-semibold text-stone-500">Extraction des mots de la grille...</p>
            </div>
        {:else if error}
            <p class="text-center text-xs font-bold text-red-600 py-6">⚠️ {error}</p>
        {:else if words.length === 0}
            <div class="flex flex-col items-center justify-center text-center py-12 select-none">
                <span class="text-3xl mb-2">🧩</span>
                <p class="text-xs font-bold text-stone-500">Aucun mot sur la grille</p>
                <p class="text-[10px] text-stone-400 mt-0.5">Posez vos premières lettres pour pouvoir consulter leurs définitions !</p>
            </div>
        {:else}
            <!-- List of words and definitions -->
            <div class="space-y-3.5">
                {#each definitions as def (def.word)}
                    <section class="rounded-3xl border border-stone-200/60 bg-white shadow-sm overflow-hidden relative group">
                        <!-- Word Header -->
                        <header class="flex items-center justify-between gap-3 px-4 py-3 bg-stone-50 border-b border-stone-100">
                            <h3 class="text-md font-extrabold text-stone-800 uppercase tracking-wide select-all">
                                {def.word}
                            </h3>
                            {#if def.url}
                                <a 
                                    href={def.url} 
                                    target="_blank" 
                                    rel="noopener noreferrer"
                                    class="text-[10px] font-bold text-brand-emerald bg-brand-emerald-light border border-brand-emerald/15 hover:bg-brand-emerald hover:text-white px-2.5 py-1 rounded-full transition-all"
                                >
                                    {def.url.includes('wiktionary.org') ? 'Wiktionnaire' : 'Larousse'} ↗
                                </a>
                            {/if}
                        </header>

                        <!-- Definition Body -->
                        <div class="p-4 text-xs text-stone-700 leading-relaxed space-y-3.5">
                            {#if def.def && def.def.length > 0}
                                {#each def.def as group}
                                    <div class="rounded-2xl border border-stone-200/50 p-3 bg-stone-50/40">
                                        <!-- Grammatical Class tag -->
                                        <div class="mb-2">
                                            <span class="inline-flex items-center rounded-full bg-brand-emerald-light text-brand-emerald text-[9px] font-bold px-2 py-0.5 border border-brand-emerald/10 select-none">
                                                {group.type}
                                            </span>
                                        </div>
                                        
                                        <ol class="list-decimal pl-4.5 space-y-1.5 text-[11px] text-stone-600 font-medium">
                                            {#each group.definitions.slice(0, 3) as d}
                                                <li class="pl-1">{d}</li>
                                            {/each}
                                        </ol>
                                    </div>
                                {/each}
                            {:else}
                                <p class="text-[11px] text-stone-400 italic text-center py-2">
                                    Aucune définition trouvée dans le Larousse pour l’instant.
                                </p>
                            {/if}
                        </div>
                    </section>
                {/each}

                {#if definitions.length < words.length}
                    <!-- Loading skeleton/placeholder for definitions still being loaded -->
                    <div class="flex items-center justify-center py-4 gap-2">
                        <svg class="animate-spin text-brand-emerald w-4 h-4" fill="none" stroke="currentColor" stroke-width="3" viewBox="0 0 24 24">
                            <circle cx="12" cy="12" r="10" stroke="rgba(12,106,77,0.15)" stroke-width="3"></circle>
                            <path d="M22 12a10 10 0 0 1-10 10" stroke="currentColor" stroke-width="3" stroke-linecap="round"></path>
                        </svg>
                        <span class="text-[10px] font-bold text-stone-400">Chargement des autres définitions...</span>
                    </div>
                {/if}
            </div>
        {/if}
    </div>
</div>
