<script lang='ts'>
    import { onMount } from "svelte";
    import { page } from "$app/stores";
    import { gameStore } from "$lib/stores/game";
    import { get, writable } from "svelte/store";
    import { api } from "$lib/api";
    import { goto } from "$app/navigation";
    import GameMenu from "$lib/components/GameMenu.svelte";
    import { extractDefinitions, getDefinition, type DefinitionGroup, type WiktionaryDefinition } from "$lib/utils/get_definition";

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

	function extractWordsFromBoard(board: string[][]): string[] {
		const H = board.length;
		const W = H > 0 ? board[0].length : 0;
		const isLetter = (c: string) => /^[A-Z]$/.test(c);
		const found = new Set<string>();

		// Horizontal scan
		for (let y = 0; y < H; y++) {
			let x = 0;
			while (x < W) {
				if (isLetter(board[y][x]) && (x === 0 || !isLetter(board[y][x - 1]))) {
					let w = '';
					let xi = x;
					while (xi < W && isLetter(board[y][xi])) {
						w += board[y][xi];
						xi++;
					}
					if (w.length >= 2) found.add(w);
					x = xi;
				} else {
					x++;
				}
			}
		}

		// Vertical scan
		for (let x = 0; x < W; x++) {
			let y = 0;
			while (y < H) {
				if (isLetter(board[y][x]) && (y === 0 || !isLetter(board[y - 1][x]))) {
					let w = '';
					let yi = y;
					while (yi < H && isLetter(board[yi][x])) {
						w += board[yi][x];
						yi++;
					}
					if (w.length >= 2) found.add(w);
					y = yi;
				} else {
					y++;
				}
			}
		}

		return Array.from(found).sort((a, b) => a.localeCompare(b));
	}

	onMount(async () => {
		const id = $page.params.id;
		const stored = get(gameStore);
		if (stored?.id === id) {
			game = stored;
			words = extractWordsFromBoard(stored.board);
			loading = false;
		} else {
            try {
                const res = await api.get<GameInfo>(`/game/${id}`);
                game = res.data;
                gameStore.set(res.data);
                words = extractWordsFromBoard(res.data.board);
            } catch (e: any) {
                error = e?.response?.data?.message || 'Impossible de charger les mots de la partie';
            } finally {
                loading = false;
            }
        }


        if (words.length) {
            words.forEach(async (word) => {
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

                definitions.push(newDef);
            });
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
    <div class="px-4 pt-4 flex justify-between items-center">
		<button
			class="text-sm text-blue-600 hover:underline flex items-center"
			onclick={backToGame}
		>
			← Retour à la partie
		</button>
        <GameMenu showScores={showScore} gameId={game.id} />
  	</div>

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
						<section class="rounded-lg border border-amber-300 bg-white shadow-sm overflow-hidden">
							<header class="flex items-center justify-between gap-3 px-4 py-3 bg-amber-50 border-b border-amber-200">
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
													<span class="inline-flex items-center rounded-full bg-amber-100 text-amber-800 text-xs font-medium px-2.5 py-1">
														{group.type}
													</span>
												</div>
												<ol class="list-decimal pl-6 p-3 space-y-1 text-gray-800">
													{#each group.definitions as d}
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
