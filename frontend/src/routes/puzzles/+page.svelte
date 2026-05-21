<script lang="ts">
	import { onMount } from 'svelte';
	import { get } from 'svelte/store';
	import HeaderBar from '$lib/components/HeaderBar.svelte';
	import { api } from '$lib/api';
	import PuzzleTimer from '$lib/components/PuzzleTimer.svelte';
	import GameBoard from '$lib/components/GameBoard.svelte';
	import Board from '$lib/components/Board.svelte';
	import { useBoardGame } from '$lib/hooks/useBoardGame.svelte';
	import { pendingMove } from '$lib/stores/pendingMove';
	import type { PuzzleInfo, PuzzleAttempt } from '$lib/types/puzzle';
	import type { GameInfo } from '$lib/types/game_infos';

	let loading = $state(true);
	let error = $state<string | null>(null);
	let puzzle = $state<PuzzleInfo | null>(null);
	let submitted = $state<PuzzleAttempt | null>(null);
	let isSubmitting = $state(false);
	let attemptId = $state<string | null>(null);
	let startedAt = $state<Date | null>(null);
	let timeoutSeconds = $state(0);
	let timedOut = $state(false);
	let puzzleGame = $state<GameInfo | null>(null);
	let submittedWords = $derived((submitted?.words_played ?? []) as Array<{ word: string; position: string; direction: string; score: number }>);
	let submittedPuzzleGame = $derived(puzzleGame ? buildGameWithSubmittedWords(puzzleGame, submittedWords) : null);

	const boardGame = useBoardGame({
		get simulateScoreEndpoint() {
			return puzzle ? `/puzzles/${puzzle.id}/simulate_score` : '';
		},
		onSubmit: async (payload) => {
			if (!puzzle || timedOut) return null;

			const wordsPlayed = payload.word
				? [{
					word: payload.word,
					position: `${payload.x},${payload.y}`,
					direction: payload.dir === 'H' ? 'horizontal' : 'vertical'
				}]
				: [];

			await submitAttempt(wordsPlayed, payload.letters);
			return null;
		}
	});

	onMount(async () => {
		await loadCurrentPuzzle();
	});

	async function loadCurrentPuzzle() {
		try {
			loading = true;
			error = null;
			const res = await api.get('/puzzles/today');
			const loadedPuzzle = res.data as PuzzleInfo;
			puzzle = loadedPuzzle;
			puzzleGame = toGameInfo(loadedPuzzle);
			boardGame.setRackFromString(loadedPuzzle.available_letters);
			pendingMove.set([]);

			if (puzzle && !puzzle.has_player_attempted) {
				const startRes = await api.post(`/puzzles/${puzzle.id}/start`, {});
				attemptId = startRes.data.attempt_id;
				startedAt = new Date(startRes.data.started_at);
				timeoutSeconds = startRes.data.timeout_seconds;
			} else if (puzzle?.has_player_attempted) {
				await loadSubmittedAttempt(puzzle.id);
			}
		} catch (e: any) {
			error = e?.response?.data?.message || 'Erreur lors du chargement du puzzle';
		} finally {
			loading = false;
		}
	}

	async function loadSubmittedAttempt(puzzleId: string) {
		try {
			const historyRes = await api.get('/puzzles?limit=50&offset=0');
			const history = (historyRes.data || []) as Array<{
				id: string;
				has_attempted: boolean;
				player_attempt?: PuzzleAttempt;
			}>;

			const item = history.find((h) => h.id === puzzleId && h.has_attempted && h.player_attempt);
			if (item?.player_attempt) {
				submitted = item.player_attempt;
			}
		} catch (e) {
			console.warn('Impossible de charger la tentative existante du puzzle', e);
		}
	}

	async function submitAttempt(
		wordsPlayed: Array<{ word: string; position: string; direction: string }>,
		letters: Array<{ x: number; y: number; char: string; blank?: boolean }> = []
	) {
		if (!puzzle || timedOut) return;

		try {
			isSubmitting = true;
			const res = await api.post(`/puzzles/${puzzle.id}/attempts`, {
				puzzle_id: puzzle.id,
				words_played: wordsPlayed,
				letters
			});

			submitted = res.data;
		} catch (e: any) {
			const message =
				e?.response?.data?.error ||
				e?.response?.data?.message ||
				e?.message ||
				'Erreur lors de la soumission de la tentative';
			throw new Error(message);
		} finally {
			isSubmitting = false;
		}
	}

	async function handleTimeout() {
		timedOut = true;
		if (!puzzle || submitted || isSubmitting) return;
		const currentMoves = get(pendingMove);
		if (currentMoves.length === 0) {
			try {
				await submitAttempt([]);
			} catch (e: any) {
				alert(e?.message || 'Erreur lors de la soumission de la tentative');
			}
			return;
		}
		await boardGame.playMove();
	}

	async function handlePuzzleValidate() {
		const currentMoves = get(pendingMove);
		if (currentMoves.length === 0) {
			try {
				await submitAttempt([]);
			} catch (e: any) {
				alert(e?.message || 'Erreur lors de la soumission de la tentative');
			}
			return;
		}
		await boardGame.playMove();
	}

	function toGameInfo(p: PuzzleInfo): GameInfo {
		const board = normalizeBoard(p.board);
		return {
			id: p.id,
			name: `Puzzle du jour ${p.puzzle_date}`,
			board,
			your_rack: p.available_letters,
			players: [],
			moves: [],
			current_turn: 0,
			current_turn_username: '',
			status: 'active',
			remaining_letters: 0,
			is_your_game: false,
			pass_count: 0
		};
	}

	function normalizeBoard(raw: any): string[][] {
		if (!Array.isArray(raw)) {
			return Array.from({ length: 15 }, () => Array.from({ length: 15 }, () => ''));
		}
		return raw.map((row: any) => {
			if (!Array.isArray(row)) return Array.from({ length: 15 }, () => '');
			return row.map((cell: any) => (typeof cell === 'string' ? cell : ''));
		});
	}

	function getLevelLabel(level: number): string {
		const labels: Record<number, string> = {
			0: 'Infini',
			1: 'Facile',
			2: 'Moyen',
			3: 'Difficile'
		};
		return labels[level] || `Niveau ${level}`;
	}

	function buildGameWithSubmittedWords(base: GameInfo, words: Array<{ word: string; position: string; direction: string }>): GameInfo {
		const board = base.board.map((row) => [...row]);

		for (const w of words) {
			const [sx, sy] = w.position.split(',').map((n) => Number.parseInt(n, 10));
			if (!Number.isFinite(sx) || !Number.isFinite(sy)) continue;

			const horizontal = w.direction === 'horizontal';
			for (let i = 0; i < w.word.length; i++) {
				const x = horizontal ? sx + i : sx;
				const y = horizontal ? sy : sy + i;
				if (x < 0 || x >= 15 || y < 0 || y >= 15) break;
				const ch = w.word[i]?.toUpperCase?.() ?? '';
				if (!ch) continue;
				if (!board[y][x] || board[y][x] === ch) {
					board[y][x] = ch;
				}
			}
		}

		return {
			...base,
			board
		};
	}
</script>

{#snippet historyButton()}
	<a
		href="/puzzles/history"
		class="inline-flex items-center gap-2 rounded-full bg-white px-3 py-1.5 text-sm font-medium text-gray-700 ring-1 ring-black/5 shadow-sm hover:bg-gray-50"
		aria-label="Voir l'historique des puzzles"
	>
		<svg width="16" height="16" viewBox="0 0 24 24" fill="none" aria-hidden="true">
			<path d="M12 8v5l3 2" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
			<path d="M3.05 11A9 9 0 1 1 6 17.3" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
			<path d="M3 4v7h7" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
		</svg>
		<span>Historique</span>
	</a>
{/snippet}

<div class="h-[100dvh] flex flex-col overflow-hidden bg-gradient-to-b from-emerald-50 to-white">
	<HeaderBar title="Puzzle du jour" back={true} right={historyButton} />

	<main class="flex-1 min-h-0 overflow-hidden max-w-4xl w-full mx-auto">
		{#if loading}
			<div class="h-full grid place-items-center px-4">
				<p class="text-gray-600">Chargement du puzzle...</p>
			</div>
		{:else if error}
			<div class="h-full px-4 py-6">
				<div class="rounded-lg bg-red-50 p-4 ring-1 ring-red-200 mb-4">
					<p class="text-red-700 font-medium">{error}</p>
				</div>
			</div>
		{:else if submitted}
			<div class="h-full px-4 py-6">
				<div class="rounded-lg bg-emerald-50 p-6 ring-1 ring-emerald-200 mb-6">
					<h2 class="text-2xl font-bold text-emerald-700 mb-2">Puzzle complété!</h2>
					<div class="grid grid-cols-3 gap-4 mt-4">
						<div class="bg-white rounded p-3">
							<p class="text-gray-600 text-sm">Votre score</p>
							<p class="text-3xl font-bold text-emerald-700">{submitted.score}</p>
						</div>
						<div class="bg-white rounded p-3">
							<p class="text-gray-600 text-sm">Votre rang</p>
							<p class="text-3xl font-bold text-emerald-700">#{submitted.rank_today}</p>
						</div>
						<div class="bg-white rounded p-3">
							<p class="text-gray-600 text-sm">Temps utilisé</p>
							<p class="text-xl font-bold text-gray-700">
								{Math.floor(submitted.time_used_secs / 60)}m {submitted.time_used_secs % 60}s
							</p>
						</div>
					</div>
					{#if submittedWords.length > 0}
						<div class="mt-4 rounded-lg bg-white p-3 ring-1 ring-emerald-100">
							<p class="text-sm font-semibold text-gray-900 mb-2">Grille de votre tentative</p>
							<div class="mx-auto w-full max-w-[min(95vw,520px)]">
								<div class="mx-auto rounded-sm ring-1 ring-black/5 bg-white shadow p-2" style="width: min(95vw, 100%); height: min(95vw, 100%);">
									<Board
										game={submittedPuzzleGame}
										onPlaceLetter={() => {}}
										onTakeFromBoard={() => {}}
									/>
								</div>
							</div>
						</div>

						<div class="mt-4 rounded-lg bg-white p-4 ring-1 ring-emerald-100">
							<p class="text-sm font-semibold text-gray-900 mb-2">Vos mots joués</p>
							<ul class="space-y-1">
								{#each (submittedWords as Array<{ word: string; position: string; direction: string; score: number }>) as w}
									<li class="flex items-center justify-between text-sm text-gray-700">
										<span>{w.word} ({w.position}, {w.direction === 'horizontal' ? 'horizontale' : 'verticale'})</span>
										<span class="font-semibold text-emerald-700">{w.score} pts</span>
									</li>
								{/each}
							</ul>
						</div>
					{/if}
					<a href={puzzle ? `/puzzles/${puzzle.id}` : '/puzzles'} class="inline-flex mt-4 px-4 py-2 bg-emerald-600 text-white rounded-lg font-medium hover:bg-emerald-700">
						Voir le classement
					</a>
				</div>
			</div>
		{:else if puzzle}
			<div class="h-full flex flex-col min-h-0 overflow-hidden">
				<div class="flex-none px-4 pt-4 pb-2">
					<div class="flex justify-between items-center mb-4">
						<div>
							<h2 class="text-xl font-bold text-gray-900">Puzzle {getLevelLabel(puzzle.level)}</h2>
							<p class="text-sm text-gray-600">Date: {puzzle.puzzle_date}</p>
						</div>
						{#if startedAt && timeoutSeconds > 0}
							<PuzzleTimer {startedAt} {timeoutSeconds} on:timeout={handleTimeout} />
						{/if}
					</div>

					{#if puzzle.has_player_attempted}
						<div class="rounded-lg bg-amber-50 p-3 ring-1 ring-amber-200 mb-2">
							<p class="text-amber-700 text-sm font-medium">📝 Vous avez déjà tenté ce puzzle.</p>
						</div>
					{/if}
				</div>

				<div class="flex-1 min-h-0 overflow-hidden">
					{#if puzzleGame}
						<GameBoard
							game={puzzleGame}
							visibleRack={boardGame.visibleRack}
							originalRack={boardGame.originalRack}
							moveScore={boardGame.moveScore}
							submitting={isSubmitting || boardGame.submitting()}
							showValidateWhenIdle={true}
							enableValidateWhenIdle={true}
							onDropFromRack={(char: string, x: number, y: number, id?: string) =>
								boardGame.dropFromRack(char, x, y, id, puzzleGame!.board[y]?.[x] ?? '')}
							onTakeFromBoard={boardGame.takeBackFromBoard}
							onCancelPendingMove={boardGame.cancelPendingMove}
							onShuffleRack={boardGame.shuffleRack}
							onPlayMove={handlePuzzleValidate}
							disabled={isSubmitting || submitted !== null || timedOut || puzzle.has_player_attempted}
						/>
					{/if}
				</div>
			</div>
		{/if}
	</main>
</div>
