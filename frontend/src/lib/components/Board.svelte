<script lang="ts">
	import { derived, writable, get } from 'svelte/store';
	import { specialCells } from '$lib/cells';
	import { pendingMove } from '$lib/stores/pendingMove';
	import { letterValues } from '$lib/lettres_value';
	import { dndzone } from 'svelte-dnd-action';
	import { tick } from 'svelte';
  	import type { GameInfo } from '$lib/types/game_infos';

	let {
		game,
		onPlaceLetter,
		onDropFromRack,
		onTakeFromBoard,
		selectedBoardCell = null,
		selectedTile = null,
		onSelectBoardCell,
		highlightRedCoords = []
	}: {
		game: GameInfo | null;
		onPlaceLetter: (x: number, y: number, cell: string) => void;
		onDropFromRack?: (char: string, x: number, y: number, id?: string) => void;
		onTakeFromBoard?: (x: number, y: number) => void;
		selectedBoardCell?: { x: number; y: number } | null;
		selectedTile?: { id: string; char: string } | null;
		onSelectBoardCell?: (x: number | null, y?: number | null) => void;
		highlightRedCoords?: Array<{ x: number; y: number }>;
	} = $props();

	type DisplayCell = {
		x: number;
		y: number;
		char: string;
		points: number | null;
		className: string;
		isPlacedLetter?: boolean;
		isPending?: boolean;
		isBlank?: boolean;
	};

	let lastMoveCoords = $derived((() => {
		if (!game?.moves?.length) return [];
		let lastIndex = 1;
		let last = game.moves[game.moves.length - lastIndex];

		while (last && last.move.type && last.move.type === 'pass') {
			lastIndex++;
			if (game.moves.length - lastIndex < 0) return [];
			last = game.moves[game.moves.length - lastIndex];
		}

		return last
			? (last.move.letters).map(m => ({ x: m.x, y: m.y }))
			: [];
		})()
	)

	let computedBoard = $derived((() => {
		if (!game) return [];

		return game.board.map((row: string[], y: number) =>
			row.map((cell: string, x: number): DisplayCell => {
				const key = `${y},${x}`;
				const special = specialCells.get(key);
				const pending = $pendingMove.find((p) => p.x === x && p.y === y);
				const displayed = cell || pending?.letter || special || '';
				const isPlacedLetter = cell !== "" && !pending;
				const inLastTurn = lastMoveCoords.some(p => p.x === x && p.y === y);
				const isHistoricBlank = !!game?.blank_tiles?.some((b) => b.x === x && b.y === y);

				let base = "relative aspect-square w-full text-center flex items-center justify-center border border-black/5 cursor-pointer select-none overflow-hidden";
				let color = "";

				if (isPlacedLetter) {
					color = inLastTurn
						? "bg-orange-200 text-yellow-800 rounded font-bold"  // dernier coup
						: "bg-white text-yellow-800 rounded font-bold";     // anciens coups
				} else if (pending) {
					color = "bg-white text-red-700 font-extrabold rounded";
				} else {
					switch (special) {
						case "TW":
							color = "bg-rose-500 text-white text-[9px] sm:text-[11px] font-black tracking-tight";
							break;
						case "DW":
						case "★":
							color = "bg-pink-300 text-stone-850 text-[9px] sm:text-[11px] font-black tracking-tight";
							break;
						case "TL":
							color = "bg-blue-600 text-white text-[9px] sm:text-[11px] font-black tracking-tight";
							break;
						case "DL":
							color = "bg-sky-300 text-stone-850 text-[9px] sm:text-[11px] font-black tracking-tight";
							break;
						default:
							color = "bg-board-green";
							break;
					}
				}

				return {
					x, y,
					char: displayed,
					points: (pending?.blank || isHistoricBlank) ? 0 : (/^[A-Z]$/.test(displayed) ? letterValues[displayed] : null),
					className: `${base} ${color}`,
					isPlacedLetter,
					isPending: !!pending,
					isBlank: !!pending?.blank || isHistoricBlank
				};
			})
		);
	})());

	// boardItems is the dndzone items array representing each cell (15x15)
	const boardItems = writable<any[]>([]);

	// track last hovered cell during drag to find true drop target
	let lastHovered: { x: number; y: number } | null = null;

	function mapCellToItem(c: DisplayCell, isSelected: boolean) {
		let finalClassName = c.className;
		const isRedHighlighted = highlightRedCoords.some(coord => coord.x === c.x && coord.y === c.y);

		if (isRedHighlighted) {
			// Red highlighted tiles for final puzzle attempt: premium high-contrast red/rose style!
			// Deep rich crimson red background, solid white text, crisp borders & modern ring shadow
			finalClassName = "relative aspect-square w-full text-center flex items-center justify-center cursor-pointer select-none overflow-hidden rounded-sm bg-red-600 text-white font-black border border-red-700 ring-2 ring-red-400/30 shadow-md z-10";
		} else if (isSelected) {
			// Warm golden pulsing selection highlight using built-in Tailwind animate-pulse!
			finalClassName = "relative aspect-square w-full text-center flex items-center justify-center cursor-pointer select-none overflow-hidden rounded-sm bg-amber-100 text-stone-900 ring-2 ring-brand-gold/50 border border-brand-gold animate-pulse z-10";
		} else if (c.isPlacedLetter) {
			// Clean, flat tile styling (no 3D effect, no hover scaling, completely static)
			if (c.className.includes("bg-orange-200")) {
				// Letters played in the last move
				finalClassName = "relative aspect-square w-full text-center flex items-center justify-center select-none overflow-hidden rounded-sm bg-orange-100 text-orange-950 font-bold border border-orange-400";
			} else {
				// Regular played letters on the board
				finalClassName = "relative aspect-square w-full text-center flex items-center justify-center select-none overflow-hidden rounded-sm bg-stone-50 text-stone-800 font-bold border border-stone-200";
			}
		} else if (c.isPending) {
			// Pending letter placement: modern, extremely high-contrast, premium indigo style!
			// Deep rich indigo background, solid white text, crisp borders & modern ring shadow
			finalClassName = "relative aspect-square w-full text-center flex items-center justify-center cursor-pointer select-none overflow-hidden rounded-sm bg-indigo-600 text-white font-black border border-indigo-700 ring-2 ring-indigo-500/30 shadow-md z-10";
		}

		return {
			id: `cell-${c.y}-${c.x}`,
			slotId: `${c.y}-${c.x}`,
			x: c.x,
			y: c.y,
			char: c.char,
			points: c.points,
			className: finalClassName,
			disabled: !!c.isPlacedLetter,
			dragDisabled: true,
			type: 'board-cell',
			isSelected,
			isPending: c.isPending,
			isRedHighlighted,
			isBlank: c.isBlank
		};
	}

	// rebuild base items from computedBoard whenever it changes
	$effect(() => {
		try {
			if ((window as any).__dndActive) return;
		} catch (err) {}
		if (computedBoard) {
			const flat = computedBoard.flat();
			const base = flat.map(c => {
				const isSelected = selectedBoardCell && selectedBoardCell.x === c.x && selectedBoardCell.y === c.y;
				return mapCellToItem(c, !!isSelected);
			});
			boardItems.set(base);
		}
	});

	// function handleDragOver(e: DragEvent) {
	// 	e.preventDefault();
	// }

	// function handleDragEnter(e: DragEvent, el: HTMLElement) {
	// 	e.preventDefault();
	// 	el.classList.add('drop-target');
	// }

	// function handleDragLeave(e: DragEvent, el: HTMLElement) {
	// 	e.preventDefault();
	// 	el.classList.remove('drop-target');
	// }

	function handleCellClick(x: number, y: number) {
		const moves = get(pendingMove);
		const idx = moves.findIndex(m => m.x === x && m.y === y);
		if (idx === -1) {
			if (selectedTile) {
				if (typeof onPlaceLetter === 'function') {
					onPlaceLetter(x, y, '');
				}
			} else {
				if (typeof onSelectBoardCell === 'function') {
					if (selectedBoardCell && selectedBoardCell.x === x && selectedBoardCell.y === y) {
						onSelectBoardCell(null);
					} else {
						onSelectBoardCell(x, y);
					}
				}
			}
			return;
		}
		if (typeof onTakeFromBoard === 'function') {
			onTakeFromBoard(x, y);
		} else {
			// fallback: remove from pendingMove locally
			pendingMove.update(ms => ms.filter(m => !(m.x === x && m.y === y)));
		}
	}

	// function handleDrop(e: DragEvent, x: number, y: number) {
	// 	e.preventDefault();
	// 	// prevent other handlers (e.g. svelte-dnd-action) from intercepting this drop
	// 	try { e.stopImmediatePropagation?.(); } catch {};
	// 	e.stopPropagation();
	// 	try {
	// 		let char: string | undefined;
	// 		const raw = e.dataTransfer?.getData('text/plain');
	// 		if (raw) {
	// 			try {
	// 				const data = JSON.parse(raw);
	// 				char = data?.char || data;
	// 			} catch (err) {
	// 				// maybe raw is just a single char
	// 				char = raw;
	// 			}
	// 		}
	// 		// fallback to global (some DnD libs strip dataTransfer)
	// 		if (!char && (window as any).__draggedTile) {
	// 			char = (window as any).__draggedTile.char;
	// 		}
	// 		if (!char) return;
	// 		// placement handled directly (drag-drop); call onPlaceLetter to reuse existing placement logic
	// 		// prefer move effect
	// 		try { if (e.dataTransfer) { e.dataTransfer.dropEffect = 'move'; } } catch (err) {}
	// 		onPlaceLetter(x, y, game!.board[y][x]);
	// 		// clear global fallback
	// 		try { (window as any).__draggedTile = null; } catch (err) {}
	// 	} catch (err) {
	// 		console.error('drop parse error', err);
	// 	}
	// }

	function handleConsider(detail: any) {
		try { (window as any).__dndActive = true; } catch {}
		const items = detail.items as any[] | undefined;
		if (!items) return;
		// build base from computedBoard to keep array length/order stable
		const flat = computedBoard?.flat() || [];

 		// if the dnd is trying to reorder existing board cells, be lenient for a single insertion
 		const originalIds = flat.map((c: any) => `cell-${c.y}-${c.x}`);
 		const insertedIndex = items.findIndex(it => !String(it.id).startsWith('cell-'));

		if (insertedIndex === -1) {
			// no external insertion => any index change means a reorder; reject ALL reorders of board cells
			for (let i = 0; i < items.length; i++) {
				const it = items[i];
				if (String(it.id).startsWith('cell-')) {
					const origIndex = originalIds.indexOf(it.id);
					if (origIndex !== -1 && origIndex !== i) {
						const baseRestore = flat.map(c => {
							const isSelected = selectedBoardCell && selectedBoardCell.x === c.x && selectedBoardCell.y === c.y;
							return mapCellToItem(c, !!isSelected);
						});
						boardItems.set(baseRestore);
						return;
					}
				}
			}
		} else {
			// there is an inserted external item; allow only the simple shift caused by insertion
			// compute expected index for each original cell after insertion and require exact mapping
			for (let origIndex = 0; origIndex < originalIds.length; origIndex++) {
				const originalId = originalIds[origIndex];
				const expectedIndex = origIndex < insertedIndex ? origIndex : origIndex + 1;
				const itAtExpected = items[expectedIndex];
				if (!itAtExpected || String(itAtExpected.id) !== originalId) {
					// More than a simple insertion occurred -> reject and restore
					const baseRestore = flat.map(c => {
						const isSelected = selectedBoardCell && selectedBoardCell.x === c.x && selectedBoardCell.y === c.y;
						return mapCellToItem(c, !!isSelected);
					});
					boardItems.set(baseRestore);
					return;
				}
			}
		}
		const base = flat.map(c => {
			const isSelected = selectedBoardCell && selectedBoardCell.x === c.x && selectedBoardCell.y === c.y;
			const item = mapCellToItem(c, !!isSelected);
			return {
				...item,
				dragDisabled: true,
				type: 'board-cell',
				__isPreview: false,
				__origChar: c.char
			};
		});

		// If there's an external item, create a silent preview entry so the DnD library can operate,
		// but don't change the visible character in the UI (we'll render the original char while dragging).
		const inserted = items.find(it => !String(it.id).startsWith('cell-'));
		if (inserted) {
			let targetX: number, targetY: number;
			if (lastHovered) {
				targetX = lastHovered.x;
				targetY = lastHovered.y;
			} else {
				const idx = items.findIndex(it => !String(it.id).startsWith('cell-'));
				targetX = idx % 15;
				targetY = Math.floor(idx / 15);
			}

			const slotIndex = targetY * 15 + targetX;
			if (base[slotIndex]) {
				// keep original char in __origChar, set char to inserted.char so the dnd action can detect the change
				base[slotIndex] = { ...base[slotIndex], __origChar: base[slotIndex].char, char: inserted.char, __isPreview: true };
			}
		}

		boardItems.set(base);
	}

	async function handleFinalize(detail: any) {
		try {
			const items = (detail as any).items as any[] | undefined;
			console.log('[board] finalize start', { items, lastHovered });

			let char: string | undefined;
			let id: string | undefined;
			let x: number | undefined;
			let y: number | undefined;

			if (items) {
				// try to find an injected external item first
				const insertedIndex = items.findIndex(it => !String(it.id).startsWith('cell-'));
				if (insertedIndex !== -1) {
					const inserted = items[insertedIndex];
					char = inserted.char || (inserted.payload && inserted.payload.char) || inserted.id;
					id = inserted.id;
					if (lastHovered) {
						x = lastHovered.x; y = lastHovered.y;
					} else {
						x = insertedIndex % 15; y = Math.floor(insertedIndex / 15);
					}
					console.log('[board] finalize found external item', { insertedIndex, char, id, x, y });
				}

				// fallback 1: if no external item, check global window.__draggedTile set by rack dragstart
				if (char === undefined && (window as any).__draggedTile) {
					const gt = (window as any).__draggedTile;
					char = gt.char; id = gt.id;
					if (lastHovered) {
						x = lastHovered.x; y = lastHovered.y;
					}
					console.log('[board] finalize fallback to window.__draggedTile', { char, id, lastHovered });
				}

				// fallback 2: detect which cell changed by diffing items vs computedBoard chars
				if (char === undefined && computedBoard) {
					const flatComputed = computedBoard.flat();
					// items are the dndzone items; find first index where char differs from computedBoard
					const diffIndex = items.findIndex((it, idx) => {
						const computedChar = flatComputed[idx] ? (flatComputed[idx].char ?? '') : '';
						const itChar = it.char ?? (it.payload && it.payload.char) ?? '';
						return itChar !== computedChar;
					});
					if (diffIndex !== -1) {
						const it = items[diffIndex];
						char = it.char || (it.payload && it.payload.char) || it.id;
						x = diffIndex % 15; y = Math.floor(diffIndex / 15);
						console.log('[board] finalize fallback diff detected', { diffIndex, char, x, y });
					}
				}
			}

			if (char !== undefined && x !== undefined && y !== undefined) {
				// update UI immediately so the tile stays visible
				boardItems.update(arr => {
					const idx = y! * 15 + x!;
					if (arr[idx]) arr[idx] = { ...arr[idx], char };
					return arr;
				});

				// wait one tick so dndzone settles and its internal DOM adjustments finish
				await tick();

				// then notify parent to update pendingMove/originalRack
				if (typeof onDropFromRack === 'function') {
					console.log('[board] calling onDropFromRack', { char, x, y, id });
					onDropFromRack(char, x, y, id);
				}
			} else {
				console.log('[board] finalize: no target detected, aborting');
			}

		} finally {
			try { (window as any).__dndActive = false; } catch (err) {}
			if (computedBoard) {
				const flat = computedBoard.flat();
				const base = flat.map(c => {
					const isSelected = selectedBoardCell && selectedBoardCell.x === c.x && selectedBoardCell.y === c.y;
					return mapCellToItem(c, !!isSelected);
				});
				boardItems.set(base);
			}
			lastHovered = null;
			console.log('[board] finalize end');
		}
	}
</script>

<div
	class="grid grid-cols-15 gap-[1px] border border-amber-500 w-full max-w-full mx-auto bg-amber-500"
	use:dndzone={{ items: $boardItems, dropFromOthersDisabled: false, dragDisabled: true }}
	onconsider={({ detail }) => { handleConsider(detail); }}
	onfinalize={({ detail }) => { handleFinalize(detail); }}
>
	{#each $boardItems as item (item.id)}
		<!-- svelte-ignore a11y_click_events_have_key_events -->
		<!-- svelte-ignore a11y_no_static_element_interactions -->
		<div
			class={item.className + (item.disabled ? ' pointer-events-none' : '')}
			draggable="false"
			onclick={() => { if (!item.disabled) handleCellClick(item.x, item.y); }}
			ondragover={(e) => { e.preventDefault(); }}
			ondragenter={(e) => { e.preventDefault(); lastHovered = { x: item.x, y: item.y }; (e.currentTarget as HTMLElement).classList.add('drop-target'); }}
			ondragleave={(e) => { e.preventDefault(); lastHovered = null; (e.currentTarget as HTMLElement).classList.remove('drop-target'); }}
			ondrop={(e) => { (e.currentTarget as HTMLElement).classList.remove('drop-target'); }}
		>
			<span class="text-[13px] sm:text-[18px] font-extrabold select-none transition-all duration-150">
				{item.__isPreview ? item.__origChar : item.char}
			</span>

			{#if item.points !== null}
				<span class="absolute bottom-[-1.5px] right-[1px] text-[7px] sm:text-[9px] font-bold tabular-nums {item.isPending ? 'text-indigo-200' : (item.isRedHighlighted ? 'text-red-200' : 'text-stone-600/90')}">
					{item.points}
				</span>
			{/if}
			{#if item.isBlank}
				<span title="Joker" class="absolute top-0.5 left-0.5 w-1.5 h-1.5 rounded-full {item.isPending ? 'bg-indigo-200' : (item.isRedHighlighted ? 'bg-red-200' : 'bg-gray-500/70')}"></span>
			{/if}
		</div>
	{/each}
</div>


<style>
	:global(.grid-cols-15) {
		grid-template-columns: repeat(15, minmax(0, 1fr));
	}

	:global(.drop-target) {
		outline: 3px solid rgba(255,255,255,0.6);
		transform: scale(1.02);
		transition: transform 0.08s ease;
	}

	@keyframes pulse-slow {
		0%, 100% {
			transform: scale(1.03);
			box-shadow: 0 0 12px rgba(245, 158, 11, 0.5), inset 0 0 4px rgba(245, 158, 11, 0.2);
			border-color: rgba(245, 158, 11, 0.8);
		}
		50% {
			transform: scale(1.0);
			box-shadow: 0 0 6px rgba(245, 158, 11, 0.2);
			border-color: rgba(245, 158, 11, 0.4);
		}
	}

	:global(.animate-pulse-slow) {
		animation: pulse-slow 2s cubic-bezier(0.4, 0, 0.2, 1) infinite;
	}
</style>
