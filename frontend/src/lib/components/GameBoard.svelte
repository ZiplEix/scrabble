<!--
  GameBoard.svelte — composant réutilisable qui encapsule :
    - le plateau (Board.svelte)
    - le rack du joueur
    - la barre d'actions (Annuler, Mélanger, Valider + slot pour actions extras)

  Utilisé par /games/[id] et /puzzles.
-->
<script lang="ts">
	import { type Readable, type Writable } from 'svelte/store';
	import { dndzone } from 'svelte-dnd-action';
	import { flip } from 'svelte/animate';
	import { cubicOut } from 'svelte/easing';
	import Board from '$lib/components/Board.svelte';
	import { pendingMove } from '$lib/stores/pendingMove';
	import { letterValues } from '$lib/lettres_value';
	import type { GameInfo } from '$lib/types/game_infos';
	import type { RackLetter } from '$lib/hooks/useBoardGame.svelte';

	interface Props {
		game: GameInfo;
		visibleRack: Readable<RackLetter[]>;
		originalRack: Writable<RackLetter[]>;
		moveScore: Readable<number>;
		submitting: boolean;
		onDropFromRack: (char: string, x: number, y: number, id?: string) => void;
		onTakeFromBoard: (x: number, y: number) => void;
		onCancelPendingMove: () => void;
		onShuffleRack: () => void;
		onPlayMove: () => void;
		/** Afficher le bouton Valider même quand aucune lettre n'est posée. */
		showValidateWhenIdle?: boolean;
		/** Autoriser Valider sans lettres (ex: puzzle = soumission à 0 point). */
		enableValidateWhenIdle?: boolean;
		/** Slot pour boutons spécifiques (Passer, Échanger...). Affiché quand aucun coup n'est en cours. */
		extraIdleActions?: import('svelte').Snippet;
		/** Slot pour boutons spécifiques toujours visibles. */
		extraActions?: import('svelte').Snippet;
		disabled?: boolean;
	}

	let {
		game,
		visibleRack,
		originalRack,
		moveScore,
		submitting,
		onDropFromRack,
		onTakeFromBoard,
		onCancelPendingMove,
		onShuffleRack,
		onPlayMove,
		showValidateWhenIdle = false,
		enableValidateWhenIdle = false,
		extraIdleActions,
		extraActions,
		disabled = false,
	}: Props = $props();

	let selectedTile = $state<{ id: string; char: string } | null>(null);
	let selectedBoardCell = $state<{ x: number; y: number } | null>(null);

	function onPlaceLetter(x: number, y: number, _cell: string) {
		if (selectedTile) {
			onDropFromRack(selectedTile.char, x, y, selectedTile.id);
			selectedTile = null; // deselect after placing
		}
	}

	function onSelectBoardCell(x: number | null, y?: number | null) {
		if (x === null) {
			selectedBoardCell = null;
		} else {
			selectedBoardCell = { x, y: y! };
		}
	}

	let hasPendingLetters = $derived($pendingMove.length > 0);
	let shouldShowValidate = $derived(hasPendingLetters || showValidateWhenIdle);
	let canValidate = $derived(
		hasPendingLetters ? $moveScore > 0 : enableValidateWhenIdle
	);
</script>

<div class="flex flex-col min-h-0 h-full flex-1 overflow-hidden">
	<!-- Board -->
	<div class="flex-1 grid place-items-center px-1 sm:px-3 min-h-0">
		<div class="mx-auto w-full max-w-full sm:max-w-[640px]">
			<div
				class="mx-auto rounded-xl ring-1 ring-black/5 bg-amber-500/10 shadow p-1 sm:p-2 border border-amber-500/20"
				style="width: min(100vw - 8px, 100%); height: min(100vw - 8px, 100%);"
			>
				<Board
					{game}
					{onPlaceLetter}
					{selectedBoardCell}
					{selectedTile}
					{onSelectBoardCell}
					onTakeFromBoard={onTakeFromBoard}
					onDropFromRack={(char, x, y, id) => {
						onDropFromRack(char, x, y, id);
						selectedBoardCell = null;
						selectedTile = null;
					}}
				/>
			</div>
		</div>
	</div>

	<!-- Rack -->
	<div class="px-3 flex-none mb-3">
		<div class="mx-auto max-w-[640px] rounded-2xl bg-white/90 backdrop-blur-md ring-1 ring-black/5 shadow-lg">
			<div class="px-2 pt-2 pb-2">
				<div class="overflow-x-auto no-scrollbar">
					<div
						class="flex gap-1 whitespace-nowrap justify-center min-h-[3rem]"
						use:dndzone={{
							items: $visibleRack,
							flipDurationMs: 150,
							dropFromOthersDisabled: false,
							dragDisabled: disabled,
						}}
						onconsider={({ detail }) => originalRack.set(detail.items)}
						onfinalize={({ detail }) => originalRack.set(detail.items)}
					>
						{#each $visibleRack as item (item.id)}
							<!-- svelte-ignore a11y_click_events_have_key_events -->
							<div
								role="button"
								tabindex="0"
								draggable={!disabled}
								onclick={() => {
									if (!disabled) {
										if (selectedBoardCell) {
											// Board-First placement!
											onDropFromRack(item.char, selectedBoardCell.x, selectedBoardCell.y, item.id);
											selectedBoardCell = null;
											selectedTile = null;
										} else {
											// Rack-First selection
											if (selectedTile?.id === item.id) {
												selectedTile = null; // deselect
											} else {
												selectedTile = { id: item.id, char: item.char };
											}
										}
									}
								}}
								ondragstart={(e) => {
									e.dataTransfer?.setData('text/plain', JSON.stringify({ char: item.char }));
									try { e.dataTransfer!.effectAllowed = 'move'; e.dataTransfer!.dropEffect = 'move'; } catch {}
									(window as any).__draggedTile = { char: item.char, id: item.id };
									(window as any).__dndActive = true;
									try { e.stopPropagation(); } catch {}
								}}
								ondragend={() => {
									try { (window as any).__draggedTile = null; (window as any).__dndActive = false; } catch {}
								}}
								class="relative inline-flex m-1 w-11 h-11 rounded-xl text-center text-lg font-bold items-center justify-center select-none transition-all duration-150
								{selectedTile?.id === item.id ? 'bg-gradient-to-br from-amber-100 to-amber-200 text-stone-900 ring-2 ring-brand-gold gold-glow scale-105 border-brand-gold/40' : 'bg-gradient-to-br from-amber-50 to-brand-gold-light text-stone-850 ring-1 ring-amber-300/50 shadow-sm'}
								{disabled ? 'opacity-50 cursor-not-allowed' : 'cursor-grab active:cursor-grabbing hover:translate-y-[-1px] active:scale-95'}"
								animate:flip={{ duration: 200, easing: cubicOut }}
							>
								{item.char}
								<span class="absolute bottom-0.5 right-1 text-[10px] font-normal text-stone-500">
									{letterValues[item.char] ?? (item.char === '?' ? 0 : '')}
								</span>
							</div>
						{/each}
					</div>
				</div>
			</div>
		</div>
	</div>

	<!-- Barre d'actions -->
	<div class="px-3 flex-none" style="padding-bottom: calc(env(safe-area-inset-bottom) + 12px);">
		<div class="mx-auto max-w-[640px]">
			<div class="w-full rounded-2xl bg-white/90 backdrop-blur-md shadow-lg ring-1 ring-black/5">
				<div class="grid divide-x {shouldShowValidate || !extraIdleActions ? 'grid-cols-3' : 'grid-cols-4'}">

					<!-- Annuler -->
					<button
						class="h-12 px-2 flex flex-col items-center justify-center text-[12px] font-medium disabled:opacity-40 active:scale-[0.98] transition"
						onclick={onCancelPendingMove}
						disabled={!$pendingMove.length || disabled}
						aria-label="Annuler le coup en cours"
					>
						<svg width="18" height="18" viewBox="0 0 24 24" fill="none" aria-hidden="true">
							<path d="M9 16l-4-4 4-4" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
							<path d="M20 20a8 8 0 0 0-8-8H5" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
						</svg>
						<span>Annuler</span>
					</button>

					<!-- Mélanger -->
					<button
						class="h-12 px-2 flex flex-col items-center justify-center text-[12px] font-medium active:scale-[0.98] transition {disabled ? 'opacity-40' : ''}"
						onclick={onShuffleRack}
						{disabled}
						aria-label="Mélanger les lettres"
					>
						<svg width="18" height="18" viewBox="0 0 24 24" version="1.1" xml:space="preserve" xmlns="http://www.w3.org/2000/svg">
							<path d="M8.7,14.2C8,14.7,7.1,15,6.2,15H4c-0.6,0-1,0.4-1,1s0.4,1,1,1h2.2c1.3,0,2.6-0.4,3.7-1.2c0.4-0.3,0.5-1,0.2-1.4C9.7,13.9,9.1,13.8,8.7,14.2z"/>
							<path d="M13,10.7c0.3,0,0.6-0.1,0.8-0.3C14.5,9.5,15.6,9,16.8,9h0.8l-0.3,0.3c-0.4,0.4-0.4,1,0,1.4c0.2,0.2,0.5,0.3,0.7,0.3s0.5-0.1,0.7-0.3l2-2c0.1-0.1,0.2-0.2,0.2-0.3c0.1-0.2,0.1-0.5,0-0.8c-0.1-0.1-0.1-0.2-0.2-0.3l-2-2c-0.4-0.4-1-0.4-1.4,0s-0.4,1,0,1.4L17.6,7h-0.8c-1.8,0-3.4,0.8-4.6,2.1c-0.4,0.4-0.3,1,0.1,1.4C12.5,10.7,12.8,10.7,13,10.7z"/>
							<path d="M20.7,15.3l-2-2c-0.4-0.4-1-0.4-1.4,0s-0.4,1,0,1.4l0.3,0.3h-1.5c-1.6,0-2.9-0.9-3.6-2.3l-1.2-2.4C10.3,8.3,8.2,7,5.9,7H4C3.4,7,3,7.4,3,8s0.4,1,1,1h1.9c1.6,0,2.9,0.9,3.6,2.3l1.2,2.4c1,2.1,3.1,3.4,5.4,3.4h1.5l-0.3,0.3c-0.4,0.4-0.4,1,0,1.4c0.2,0.2,0.5,0.3,0.7,0.3s0.5-0.1,0.7-0.3l2-2C21.1,16.3,21.1,15.7,20.7,15.3z"/>
						</svg>
						<span>Mélanger</span>
					</button>

					<!-- Actions extra (slot) visibles quand idle ou toujours -->
					{#if !hasPendingLetters && extraIdleActions && !showValidateWhenIdle}
						{@render extraIdleActions()}
					{/if}

					{#if extraActions}
						{@render extraActions()}
					{/if}

					{#if shouldShowValidate}
						<!-- Valider -->
						<button
							class="relative h-12 px-2 flex flex-col items-center justify-center text-[12px] font-semibold text-white bg-green-600 rounded-r-2xl active:scale-[0.98] transition disabled:opacity-60"
							onclick={onPlayMove}
							disabled={submitting || disabled || !canValidate}
							aria-label="Valider le coup"
						>
							{#if submitting}
								<svg class="animate-spin" width="18" height="18" viewBox="0 0 24 24" fill="none" aria-hidden="true">
									<circle cx="12" cy="12" r="10" stroke="rgba(255,255,255,0.4)" stroke-width="4"></circle>
									<path d="M22 12a10 10 0 0 1-10 10" stroke="white" stroke-width="4" stroke-linecap="round"></path>
								</svg>
							{:else}
								<svg width="18" height="18" viewBox="0 0 24 24" fill="none" aria-hidden="true">
									<path d="M20 7l-9 9-4-4" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
								</svg>
							{/if}
							<span class="mt-1">{submitting ? 'Envoi...' : 'Valider'}</span>
							<span class="absolute -top-2 -right-2 text-[10px] px-2 py-0.5 rounded-full bg-white text-green-700 shadow ring-1 ring-black/5">
								{hasPendingLetters ? $moveScore : 0}
							</span>
						</button>
					{/if}
				</div>
			</div>
		</div>
	</div>
</div>
