/**
 * useBoardGame — logique partagée entre la page /games/[id] et la page /puzzles.
 *
 * Gère :
 *  - le rack (lettres disponibles du joueur)
 *  - les pendingMove (lettres posées mais pas encore soumises)
 *  - le score simulé
 *  - le placement / annulation / mélange des lettres
 *  - la soumission (via un callback `onSubmit` injecté par la page)
 */
import { writable, derived, get, type Readable } from 'svelte/store';
import { api } from '$lib/api';
import { pendingMove, type PendingMove } from '$lib/stores/pendingMove';

export type RackLetter = { id: string; char: string };

export type SubmitPayload = {
	word: string;
	x: number;
	y: number;
	dir: 'H' | 'V';
	letters: { x: number; y: number; char: string; blank: boolean }[];
	score: number;
};

export type BoardGameOptions = {
	/** Endpoint pour simuler le score (ex. /game/:id/simulate_score). Peut être un getter pour supporter les valeurs réactives. */
	readonly simulateScoreEndpoint: string;
	/** Callback appelé à la soumission d'un coup.  Doit lancer l'appel API et retourner le rack mis à jour, ou null si pas de rack (puzzle). */
	onSubmit: (payload: SubmitPayload) => Promise<string[] | null>;
};

export function useBoardGame(options: BoardGameOptions) {
	const { onSubmit } = options;

	// ── Rack ─────────────────────────────────────────────────────────────────
	const originalRack = writable<RackLetter[]>([]);

	function setRackFromString(rack: string) {
		originalRack.set(
			rack.split('').map((char, i) => ({
				id: `${i}-${char}-${crypto.randomUUID()}`,
				char
			}))
		);
	}

	function setRackFromArray(chars: string[]) {
		originalRack.set(
			chars.map((char, i) => ({
				id: `${i}-${char}-${crypto.randomUUID()}`,
				char
			}))
		);
	}

	/** Lettres du rack visibles (les lettres posées sur le board sont masquées) */
	const visibleRack: Readable<RackLetter[]> = derived(
		[originalRack, pendingMove],
		([$rack, $moves]) => {
			const usedIds = new Set($moves.map((m) => m.rackId).filter(Boolean) as string[]);
			return $rack.filter((r) => !usedIds.has(r.id));
		}
	);

	// ── Score simulé ─────────────────────────────────────────────────────────
	const moveScore: Readable<number> = derived(
		pendingMove,
		($moves, set) => {
			if (!$moves.length) return set(0);
			api
				.post(options.simulateScoreEndpoint, {
					letters: $moves.map((m) => ({
						x: m.x,
						y: m.y,
						char: m.letter.toUpperCase(),
						blank: !!m.blank
					}))
				})
				.then((res) => set(res.data.score))
				.catch(() => set(0));
		},
		0
	);

	// ── Placement ────────────────────────────────────────────────────────────
	/** Dépose une lettre du rack sur le plateau (appelé depuis Board via onDropFromRack) */
	function dropFromRack(char: string, x: number, y: number, id?: string, boardCell?: string) {
		// case déjà occupée (board historique ou coup en cours)
		const currentMoves = get(pendingMove);
		const occupiedByPending = currentMoves.some((m) => m.x === x && m.y === y);
		if (boardCell || occupiedByPending) {
			// restituer la pièce au rack
			originalRack.update((r) => {
				if (id && r.some((it) => it.id === id)) return r;
				const newItem = id
					? { id, char }
					: { id: `${Date.now()}-${char}-${crypto.randomUUID()}`, char };
				return [...r, newItem];
			});
			return;
		}

		let chosen = char;
		let isBlank = false;
		if (char === '?') {
			const input = prompt('Lettre du joker (A-Z) :');
			if (!input) {
				// annulé — restituer
				originalRack.update((r) => [
					...r,
					{ id: id ?? `${Date.now()}-?-${crypto.randomUUID()}`, char: '?' }
				]);
				return;
			}
			const up = input.trim().toUpperCase();
			if (!/^[A-Z]$/.test(up)) {
				alert('Veuillez entrer une seule lettre A–Z.');
				originalRack.update((r) => [
					...r,
					{ id: id ?? `${Date.now()}-?-${crypto.randomUUID()}`, char: '?' }
				]);
				return;
			}
			chosen = up;
			isBlank = true;
		}

		pendingMove.update((moves) => {
			const filtered = moves.filter((m) => !(m.x === x && m.y === y));
			return [...filtered, { x, y, letter: chosen, rackId: id, blank: isBlank }];
		});
		originalRack.update((r) => r.filter((item) => item.id !== id));
	}

	/** Reprendre une lettre posée du plateau et la renvoyer dans le rack */
	function takeBackFromBoard(x: number, y: number) {
		const moves = get(pendingMove);
		const move = moves.find((m) => m.x === x && m.y === y);
		if (!move) return;

		originalRack.update((r) => {
			if (move.rackId && r.some((it) => it.id === move.rackId)) return r;
			const char = move.blank ? '?' : move.letter;
			const item = move.rackId
				? { id: move.rackId, char }
				: { id: `${Date.now()}-${char}-${crypto.randomUUID()}`, char };
			return [...r, item];
		});
		pendingMove.update((ms) => ms.filter((m) => !(m.x === x && m.y === y)));
	}

	/** Annule toutes les lettres posées et les restitue au rack */
	function cancelPendingMove() {
		const moves = get(pendingMove);
		if (!moves.length) {
			pendingMove.set([]);
			return;
		}
		originalRack.update((r) => {
			const existingIds = new Set(r.map((i) => i.id));
			const toAdd = moves
				.map((m) => {
					if (m.rackId) return { id: m.rackId, char: m.letter };
					return { id: `${Date.now()}-${m.letter}-${crypto.randomUUID()}`, char: m.letter };
				})
				.filter((item) => !existingIds.has(item.id));
			return [...r, ...toAdd];
		});
		pendingMove.set([]);
	}

	/** Mélange aléatoirement les lettres du rack */
	function shuffleRack() {
		const rack = get(originalRack);
		for (let i = rack.length - 1; i > 0; i--) {
			const j = Math.floor(Math.random() * (i + 1));
			[rack[i], rack[j]] = [rack[j], rack[i]];
		}
		originalRack.set(rack);
	}

	// ── Soumission ────────────────────────────────────────────────────────────
	let submitting = $state(false);

	async function playMove() {
		const move = get(pendingMove);
		if (!move.length) return;

		const sorted = [...move].sort((a, b) => a.x - b.x || a.y - b.y);
		const sameRow = sorted.every((l) => l.y === sorted[0].y);
		const sameCol = sorted.every((l) => l.x === sorted[0].x);

		if (!sameRow && !sameCol) {
			alert('Les lettres doivent être alignées horizontalement ou verticalement.');
			return;
		}

		const direction = sameRow ? 'H' : 'V';
		const payload: SubmitPayload = {
			word: sorted.map((l) => l.letter.toUpperCase()).join(''),
			x: sorted[0].x,
			y: sorted[0].y,
			dir: direction,
			letters: move.map((m) => ({
				x: m.x,
				y: m.y,
				char: m.letter.toUpperCase(),
				blank: !!m.blank
			})),
			score: get(moveScore)
		};

		submitting = true;
		try {
			const newRack = await onSubmit(payload);
			pendingMove.set([]);
			if (newRack !== null) {
				setRackFromArray(newRack);
			}
		} catch (e: any) {
			const msg = e?.response?.data?.message ?? e?.message ?? 'Erreur lors de la validation du coup.';
			alert(msg);
		} finally {
			submitting = false;
		}
	}

	return {
		// state
		originalRack,
		visibleRack,
		moveScore,
		submitting: () => submitting,
		// actions
		setRackFromString,
		setRackFromArray,
		dropFromRack,
		takeBackFromBoard,
		cancelPendingMove,
		shuffleRack,
		playMove
	};
}
