import type User from "$lib/types/user";
import { get, writable } from "svelte/store";

const STORAGE_KEY = "user";

function createUserStore() {
    let initial: User | null = null;
    if (typeof localStorage !== "undefined") {
        const saved = localStorage.getItem(STORAGE_KEY);
		if (saved) {
			initial = JSON.parse(saved);
		}
    }

    const store = writable<User | null>(initial);

    store.subscribe((value) => {
        if (typeof localStorage === 'undefined') return;
		if (value) {
			localStorage.setItem(STORAGE_KEY, JSON.stringify(value));
		} else {
			localStorage.removeItem(STORAGE_KEY);
		}
    });

    return store;
}

import { supabase } from "$lib/supabase";

export async function logout() {
    await supabase.auth.signOut();
    user.set(null);
}

export const user = createUserStore();
