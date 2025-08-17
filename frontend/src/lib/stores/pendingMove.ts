import { writable } from 'svelte/store';

export type PendingMove = { x: number; y: number; letter: string; rackId?: string };

export const pendingMove = writable<PendingMove[]>([]);
// selectedLetter removed: drag-and-drop is the single placement method now
