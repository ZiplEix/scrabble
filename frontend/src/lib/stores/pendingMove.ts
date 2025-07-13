import { writable } from 'svelte/store';

export type PendingMove = { x: number; y: number; letter: string };

export const pendingMove = writable<PendingMove[]>([]);
export const selectedLetter = writable<string | null>(null);
