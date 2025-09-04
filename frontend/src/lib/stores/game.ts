import type { GameInfo } from "$lib/types/game_infos";
import { writable } from "svelte/store";

export const gameStore = writable<GameInfo | null>(null);
