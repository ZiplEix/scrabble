<script lang="ts">
	import { onMount } from 'svelte';
	import { api } from '$lib/api';
	import { page } from '$app/stores';
	import { get, derived, writable } from 'svelte/store';
	import Board from '$lib/components/Board.svelte';
	import { pendingMove, selectedLetter } from '$lib/stores/pendingMove';
	import { letterValues } from '$lib/lettres_value';
	import { dndzone } from 'svelte-dnd-action';

	let gameId = '';
	let game = $state<GameInfo | null>(null);
	let error = $state('');
	let loading = $state(true);
	type RackLetter = { id: string; char: string };
	let originalRack = writable<RackLetter[]>([]);
	let showScores = $state(false);

	let moveScore = derived(
		[pendingMove, page],
		([$moves, $page], set) => {
			if (!$moves.length || !game) return set(0);
			set(undefined);

			api.post(`/game/${$page.params.id}/simulate_score`, {
				letters: $moves.map(m => ({
					x: m.x,
					y: m.y,
					char: m.letter.toUpperCase()
				}))
			})
			.then(res => set(res.data.score))
			.catch(() => set(0));
		}
	)

	onMount(async () => {
		gameId = $page.params.id;
		if (!gameId) return;

		try {
			const res = await api.get(`/game/${gameId}`);
			game = res.data;
			originalRack.set(game!.your_rack.split('').map((char, i) => ({
				id: `${i}-${char}-${crypto.randomUUID()}`,
				char
			})));
		} catch (e: any) {
			error = e?.response?.data?.error || 'Erreur lors du chargement de la partie';
		} finally {
			loading = false;
		}
	});

	function onSelectLetter(letter: string) {
		const current = get(selectedLetter);
		selectedLetter.set(current === letter ? null : letter);
	}

	function onPlaceLetter(x: number, y: number, cell: string) {
		const currentMoves = get(pendingMove);
		const existing = currentMoves.find((m) => m.x === x && m.y === y);

		if (existing) {
			pendingMove.set(currentMoves.filter((m) => !(m.x === x && m.y === y)));
			return;
		}

		if (cell) return;
		const letter = get(selectedLetter);
		if (!letter) return;

		pendingMove.update((moves) => {
			const filtered = moves.filter((m) => !(m.x === x && m.y === y));
			return [...filtered, { x, y, letter }];
		});
		selectedLetter.set(null);
	}

	const placedLetters = derived(pendingMove, (moves) => moves.map((m) => m.letter));
	const visibleRack = derived(
		[originalRack, placedLetters],
		([$rack, $used]) => {
			const usedCopy = [...$used];
			return $rack.filter((l) => {
				const idx = usedCopy.indexOf(l.char);
				if (idx !== -1) {
					usedCopy.splice(idx, 1);
					return false;
				}
				return true;
			});
		}
	);

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

		const direction = sameRow ? "H" : "V";
		const startX = sorted[0].x;
		const startY = sorted[0].y;
		let word = sorted.map(l => l.letter.toUpperCase()).join("");

		const body = {
			word,
			x: startX,
			y: startY,
			dir: direction,
			letters: move.map((m) => ({ x: m.x, y: m.y, char: m.letter.toUpperCase() })),
			score: get(moveScore)
		};

		try {
			await api.post(`/game/${gameId}/play`, body);
			const res = await api.get(`/game/${gameId}`);
			game = res.data;
			originalRack.set(game!.your_rack.split('').map((char, i) => ({
				id: `${i}-${char}-${crypto.randomUUID()}`,
				char
			})));
			pendingMove.set([]);
			selectedLetter.set(null);
		} catch (e: any) {
			const msg = e?.response?.data?.message || 'Erreur lors de la validation du coup.';
			alert(msg);
		}
	}

	async function drawNewRack() {
		const ok = confirm('Êtes-vous sûr de vouloir changer toutes vos lettres ? Cela remplacera vos lettres actuelles et passera votre tour.');
		if (!ok) return;

		try {
			const res = await api.get(`/game/${gameId}/new_rack`);
			const newRack = res.data as string[];
			console.log('Nouveau rack:', newRack);
			if (newRack.length === 0) {
				alert('Plus de lettres disponibles dans le sac.');
				return;
			}
			originalRack.set(newRack.map((char, i) => ({
				id: `${i}-${char}-${crypto.randomUUID()}`,
				char
			})));
			pendingMove.set([]);
			selectedLetter.set(null);
		} catch (e: any) {
			const msg = e?.response?.data?.message || 'Erreur lors du tirage d\'un nouveau rack.';
			console.error(e);
			alert(msg);
		}
	}

	async function passTurn() {
		const ok = confirm('Êtes-vous sûr de vouloir passer votre tour ?');
		if (!ok) return;
		try {
			await api.post(`/game/${gameId}/pass`);
			const res = await api.get(`/game/${gameId}`);
			game = res.data;
		} catch (e: any) {
			const msg = e?.response?.data?.message || 'Erreur lors du passage du tour.';
			alert(msg);
		}
	}
</script>

{#if loading}
	<p class="text-center mt-8">Chargement...</p>
{:else if error}
	<p class="text-center text-red-600 mt-8">{error}</p>
{:else if game}
	<!-- Header: nom + tour + bouton classement -->
	<div class="px-4 pt-3 pb-1 w-full flex justify-between items-center">
		<div>
			<h2 class="text-xl font-bold text-gray-800">{game.name}</h2>
			<p class="text-sm text-gray-600">Tour de : <strong>{game.current_turn_username}</strong></p>
			<p class="text-xs text-gray-500">Lettres restantes : <strong>{game.remaining_letters}</strong></p>
		</div>
		<!-- Actions: classement + report -->
		<div class="flex flex-col items-end gap-2">
			<button
				class="text-xs bg-gray-200 px-3 py-1 rounded shadow hover:bg-gray-300"
				onclick={() => showScores = true}
			>
				Classement 🏆
			</button>
			<a
				href="/report"
				class="text-xs bg-gray-200 px-3 py-1 rounded shadow hover:bg-gray-300 text-center"
			>
				🛠️ reporter un bug
			</a>
		</div>
	</div>

	<div class="text-center mt-2 mb-1">
		<span class="inline-block bg-yellow-100 text-yellow-800 font-semibold text-lg px-4 py-2 rounded shadow">
			Score du coup : <strong>{$moveScore}</strong>
		</span>
	</div>

	<!-- Plateau + rack -->
	<div class="flex flex-col items-center justify-center w-full gap-2 px-2"
    	style="min-height: calc(100vh - 320px);">
		<!-- Plateau -->
		<div class="max-w-[95vw] w-full aspect-square">
			<Board
				{game}
				{onPlaceLetter}
			/>
		</div>

		<!-- Rack -->
		<div
			class="flex justify-center gap-1 mt-2 flex-wrap max-w-[95vw]"
			use:dndzone={{
				items: $visibleRack,
				flipDurationMs: 150,
				dropFromOthersDisabled: true,
				dragDisabled: false,
			}}
			onconsider={({ detail }) => originalRack.set(detail.items)}
			onfinalize={({ detail }) => originalRack.set(detail.items)}
		>
			{#each $visibleRack as item (item.id)}
				<!-- svelte-ignore a11y_click_events_have_key_events -->
				<div
					role="button"
					tabindex="0"
					class="relative w-12 h-12 rounded shadow text-center text-lg font-bold flex items-center justify-center border cursor-pointer
						{ $selectedLetter === item.char ? 'bg-yellow-400 border-yellow-600' : 'bg-yellow-100 border-yellow-400' }"
					onclick={() => onSelectLetter(item.char)}
				>
					{item.char}
					<span class="absolute bottom-0.5 right-1 text-xs font-normal text-gray-600">{letterValues[item.char]}</span>
				</div>
			{/each}
		</div>
	</div>

	<!-- Actions sticky -->
	<div class="fixed bottom-0 left-0 w-full bg-white border-t shadow-inner px-4 py-3 flex justify-between items-center">
		<div class="flex gap-3">
			<button class="bg-gray-400 text-white px-4 py-2 rounded shadow" onclick={() => { pendingMove.set([]); selectedLetter.set(null); }}>Annuler</button>
			<button class="bg-red-500 text-white px-4 py-2 rounded shadow" onclick={passTurn}>Passer</button>
		</div>
		<div class="flex gap-3">
			<button class="bg-orange-500 text-white px-4 py-2 rounded shadow" onclick={drawNewRack}>Échanger</button>
			<button class="bg-green-600 text-white px-4 py-2 rounded shadow" onclick={playMove}>Valider</button>
		</div>
	</div>

	<!-- Modal classement -->
	{#if showScores}
		<div class="fixed inset-0 bg-black/30 backdrop-blur-sm flex items-center justify-center z-50">
			<div class="bg-white rounded-lg shadow-lg w-[90vw] max-w-sm p-6">
				<h3 class="text-lg font-semibold mb-4 text-center">Classement</h3>
				<div class="flex flex-col gap-2">
					{#each game.players as player}
						<div class="flex justify-between items-center p-2 rounded border
							{player.id === game.current_turn ? 'bg-green-100 border-green-400' : 'bg-gray-50'}">
							<span>{player.username}</span>
							<span class="font-bold">{player.score}</span>
						</div>
					{/each}
				</div>
				<button class="mt-6 w-full bg-gray-300 py-2 rounded hover:bg-gray-400" onclick={() => showScores = false}>
					Fermer
				</button>
			</div>
		</div>
	{/if}
{/if}
