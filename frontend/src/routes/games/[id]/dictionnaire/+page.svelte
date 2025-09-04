<script lang='ts'>
    import { onMount } from "svelte";
    import { page } from "$app/stores";
    import { gameStore } from "$lib/stores/game";
    import { get, writable } from "svelte/store";
    import { api } from "$lib/api";
    import { goto } from "$app/navigation";
    import GameMenu from "$lib/components/GameMenu.svelte";
    import { extractDefinitions, getDefinition, type DefinitionGroup, type WiktionaryDefinition } from "$lib/utils/get_definition";
  	import type { GameInfo } from "$lib/types/game_infos";

    type Definition = {
        word: string,
        url: string
        wikidef: WiktionaryDefinition[],
        def: DefinitionGroup[]
    }

    let game: GameInfo | null = $state<GameInfo | null>(null);
	let error = $state('');
	let loading = $state(true);
	let words: string[] = $state<string[]>([]);
    let definitions: Definition[] = $state<Definition[]>([]);
    let showScore = writable<boolean>(false);

	// Construit une map (x,y) -> timestamp de pose (ms) pour chaque tuile posée
	function buildTilePlacementMap(moves: any[] | undefined): Map<string, number> {
		const map = new Map<string, number>();
		if (!moves) return map;
		// Itérer du plus ancien au plus récent et n'écrire qu'une fois par case
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

	// Extrait les mots (>=2) avec leurs coordonnées de cases
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
					y++;
				}
			}
		}
		return spans;
	}

	// Trie les mots selon le timestamp max de leurs cases (plus récent -> plus ancien)
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
			// Si on n'a aucune info (ex: cases sans move), on garde -Infinity pour qu'ils finissent en fin de liste
			const prev = wordMaxTime.get(s.word);
			if (prev === undefined || (maxT > prev)) wordMaxTime.set(s.word, maxT);
		}
		const uniqueWords = Array.from(wordMaxTime.keys());
		uniqueWords.sort((a, b) => (wordMaxTime.get(b)! - wordMaxTime.get(a)!));
		return uniqueWords;
	}

	onMount(async () => {
		const id = $page.params.id;
		const stored = get(gameStore);
		if (stored?.id === id) {
			game = stored;
			const ordered = getWordsOrderedByPlacement(stored.board, stored.moves);
			words = ordered;
			loading = false;
		} else {
            try {
                const res = await api.get<GameInfo>(`/game/${id}`);
                game = res.data;
                gameStore.set(res.data);
				const ordered = getWordsOrderedByPlacement(res.data.board, res.data.moves);
				words = ordered;
            } catch (e: any) {
                error = e?.response?.data?.message || 'Impossible de charger les mots de la partie';
            } finally {
                loading = false;
            }
        }


		if (words.length) {
			// Charger les définitions séquentiellement pour préserver l'ordre choisi
			definitions = [];
			for (const word of words) {
				let newDef: Definition = {
					word,
					url: `https://fr.wiktionary.org/wiki/${encodeURIComponent(word)}`,
					wikidef: [],
					def: []
				};
				const wikidef = await getDefinition(word);
				newDef.wikidef = wikidef ? [wikidef] : [];
				if (wikidef?.url) {
					newDef.url = wikidef.url;
				}
				if (wikidef && wikidef.extract) {
					newDef.def = extractDefinitions(wikidef.extract);
				}
				definitions = [...definitions, newDef];
			}
        }
	})

    function backToGame() {
		if (game) goto(`/games/${game.id}`);
		else goto('/');
	}
</script>

{#if loading}
  	<p class="mt-8 text-center text-gray-600">Chargement de l’historique…</p>
{:else if error}
  	<p class="mt-8 text-center text-red-600">{error}</p>
{:else if game}
    <header class="px-3 pt-2 pb-2">
        <div class="flex items-center w-full justify-between gap-2">
            <div class="flex items-center gap-2 min-w-0">
                <button
                    class="p-2 rounded-lg hover:bg-white/60 ring-1 ring-black/5 bg-white/40 backdrop-blur-sm"
                    aria-label="Retour à la partie"
                    onclick={backToGame}
                    title="Retour"
                >
                    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" aria-hidden="true">
                        <path d="M4 12h16M10 6l-6 6 6 6" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                    </svg>
                </button>
                <h2 class="text-base font-semibold text-gray-900 truncate" title={game.name}>{game.name}</h2>
            </div>

            <div class="flex items-center gap-2">
                <button
                    class="hidden sm:flex items-center gap-1 px-2.5 h-8 rounded-lg bg-emerald-600/90 hover:bg-emerald-600 text-white text-[12px] font-medium shadow-sm ring-1 ring-emerald-700/30"
                    onclick={() => showScore.set(true)}
                    title="Voir le classement"
                    aria-label="Voir le classement"
                >
                    <span>🏆</span>
                    <span>Scores</span>
                </button>
                <GameMenu showScores={showScore} gameId={game.id} />
            </div>
        </div>
    </header>

    <div class="p-4">
        <h1 class="text-xl font-bold text-gray-800 text-center mb-4">
			Dictionnaire des mots de la grille de "{game.name}"
    	</h1>

		{#if words.length === 0}
			<p class="text-gray-600">Aucune lettre n'a encore été posée</p>
		{:else}
			{#if definitions.length === 0}
				<p class="text-gray-600">Chargement des définitions…</p>
			{:else}
				<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
					{#each definitions as def (def.word)}
						<section class="rounded-lg border border-green-300 bg-white shadow-sm overflow-hidden">
							<header class="flex items-center justify-between gap-3 px-4 py-3 bg-green-50 border-b border-green-200">
								<h2 class="text-lg font-semibold text-gray-800 uppercase tracking-wide">{def.word}</h2>
								{#if def?.wikidef?.[0]?.url || def?.url}
									<a
										class="text-sm text-blue-600 hover:underline whitespace-nowrap"
										href={(def?.wikidef?.[0]?.url) ?? def.url}
										target="_blank"
										rel="noopener noreferrer"
										title="Voir sur le Larousse"
									>Larousse ↗</a>
								{/if}
							</header>

							<div class="p-4">
								{#if def && def.def && def.def.length > 0}
									<div class="space-y-4">
										{#each def.def as group}
											<div class="rounded-md border border-gray-200">
												<div class="px-3 py-2 bg-gray-50 border-b border-gray-200 flex items-center gap-2">
													<span class="inline-flex items-center rounded-full bg-green-100 text-green-800 text-xs font-medium px-2.5 py-1">
														{group.type}
													</span>
												</div>
												<ol class="list-decimal pl-6 p-3 space-y-1 text-gray-800">
													{#each group.definitions.slice(0, 3) as d}
														<li class="leading-relaxed">{d}</li>
													{/each}
												</ol>
											</div>
										{/each}
									</div>
								{:else}
									<p class="text-sm text-gray-500 italic">Aucune définition trouvée dans le Larousse pour l’instant.</p>
								{/if}
							</div>
						</section>
					{/each}
				</div>
			{/if}
		{/if}
    </div>
{/if}
