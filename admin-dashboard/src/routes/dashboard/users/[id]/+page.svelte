<script lang="ts">
    import { page } from "$app/stores";
    import { get } from "svelte/store";
    import { api } from "$lib/api";
    import { onMount } from "svelte";
    import type { User } from "../type";

    let user: User | null = $state(null);
    let loading = $state(true);

    let navUser = $derived(($page.state?.user as User | undefined) ?? null);
    $effect(() => {
        if (navUser) {
            user = navUser;
            loading = false;
        }
    });

    onMount(async () => {
        try {
            if (navUser) {
                return;
            }
            // fallback to fetch
            const id = get(page).params.id;
            const res = await api.get(`/admin/user/${id}`);
            if (res && res.data) {
                user = res.data.user as User;
            }
        } catch (err) {
            console.error("Error loading user", err);
        } finally {
            loading = false;
        }
    });
</script>

<div class="max-w-5xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <header class="mb-6">
        <h1 class="text-2xl font-bold">Détail utilisateur</h1>
        {#if user}
            <p class="text-sm text-white/70 mt-1">{user.username} · ID {user.id}</p>
        {/if}
    </header>

    {#if loading}
        <div class="space-y-3">
            <div class="h-6 w-64 bg-white/10 animate-pulse rounded"></div>
            <div class="h-4 w-80 bg-white/10 animate-pulse rounded"></div>
            <div class="h-72 w-full bg-white/10 animate-pulse rounded"></div>
        </div>
    {:else if user}
        <section class="mb-6 bg-white/4 rounded-lg p-4">
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
                <div><span class="text-white/60">Username:</span> <span class="text-white/90">{user.username}</span></div>
                <div><span class="text-white/60">Rôle:</span> <span class="text-white/90">{user.role}</span></div>
                <div><span class="text-white/60">Créé le:</span> <span class="text-white/90">{new Date(user.created_at).toLocaleString()}</span></div>
                <div><span class="text-white/60">Dernière activité:</span> <span class="text-white/90">{user.last_activity_at ? new Date(user.last_activity_at).toLocaleString() : '—'}</span></div>
                <div><span class="text-white/60">Messages:</span> <span class="text-white/90">{user.messages_count}</span></div>
                <div><span class="text-white/60">Parties:</span> <span class="text-white/90">{user.games_count} (en cours {user.ongoing_games}, finies {user.finished_games})</span></div>
            </div>
        </section>

        <section class="bg-white/4 rounded-lg p-4">
            <h2 class="text-lg font-semibold mb-3">Parties</h2>
            <div class="overflow-x-auto">
                <table class="min-w-full text-sm table-fixed w-full">
                    <colgroup>
                        <col style="width:28%" />
                        <col style="width:24%" />
                        <col style="width:24%" />
                        <col style="width:24%" />
                    </colgroup>
                    <thead>
                        <tr class="text-left text-xs text-white/60">
                            <th class="px-3 py-2">ID</th>
                            <th class="px-3 py-2">Nom</th>
                            <th class="px-3 py-2">Statut</th>
                            <th class="px-3 py-2">Créée le</th>
                        </tr>
                    </thead>
                    <tbody>
                        {#each user.games as g}
                            <tr class="border-t border-white/6">
                                <td class="px-3 py-2 text-xs text-white/70">{g.id}</td>
                                <td class="px-3 py-2 text-sm text-white/90">{g.name}</td>
                                <td class="px-3 py-2 text-xs text-white/70">{g.status}</td>
                                <td class="px-3 py-2 text-xs text-white/70">{new Date(g.created_at).toLocaleString()}</td>
                            </tr>
                        {/each}
                        {#if user.games.length === 0}
                            <tr><td colspan="4" class="px-3 py-4 text-center text-white/60">Aucune partie</td></tr>
                        {/if}
                    </tbody>
                </table>
            </div>
        </section>
    {:else}
        <div class="text-white/70">Utilisateur introuvable.</div>
    {/if}

    <div class="mt-6">
        <a href="/dashboard/users" class="text-sm text-white/70 hover:underline">← Retour aux utilisateurs</a>
    </div>
</div>
