import { writable } from "svelte/store";

export const gameStore = writable<GameInfo | null>(null);
