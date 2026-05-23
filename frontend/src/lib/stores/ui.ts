import { writable } from 'svelte/store';

// store to hide global bottom tab bar when playing puzzles or active gameplay
export const hideTabBar = writable<boolean>(false);
