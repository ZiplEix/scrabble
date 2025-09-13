<script lang='ts'>
    import { api } from "$lib/api";
    import { onMount } from "svelte";
    import { goto } from "$app/navigation";
    import type { User } from "./type";

    let users: User[] = $state([]);
    let loading = $state(true);

    onMount(async () => {
        loading = true;
        try {
            const res = await api.get('/admin/users');
            if (res && res.data) {
                users = res.data.users as User[];
            }
        } catch (err) {
            console.error('Error fetching users:', err);
        } finally {
            loading = false;
        }
    })

    function openUser(u: User) {
        goto(`/dashboard/users/${u.id}`, { state: { user: JSON.parse(JSON.stringify(u)) } });
    }
</script>

<div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <header class="mb-6">
        <h1 class="text-2xl font-bold">Utilisateur</h1>
        <p class="text-sm text-white/70 mt-1">Visualisez, filtrez et sélectionnez les utilisateurs.</p>
    </header>

    <section class="bg-white/4 rounded-lg p-4">
        <div class="overflow-x-auto">
            <table class="min-w-full text-sm table-fixed w-full">
                <colgroup>
                    <col style="width:10%" />
                    <col style="width:30%" />
                    <col style="width:15%" />
                    <col style="width:25%" />
                    <col style="width:20%" />
                </colgroup>
                <thead>
                    <tr class="text-left text-xs text-white/60">
                        <th class="px-3 py-2">ID</th>
                        <th class="px-3 py-2">Username</th>
                        <th class="px-3 py-2">Rôle</th>
                        <th class="px-3 py-2">Dernière activité</th>
                        <th class="px-3 py-2">Nb parties</th>
                    </tr>
                </thead>
                <tbody>
                    {#if loading}
                        {#each Array(8) as _, i}
                            <tr class="border-t border-white/6">
                                <td class="px-3 py-3"><div class="h-8 w-12 bg-white/10 animate-pulse rounded"></div></td>
                                <td class="px-3 py-3"><div class="h-8 w-40 bg-white/10 animate-pulse rounded"></div></td>
                                <td class="px-3 py-3"><div class="h-8 w-20 bg-white/10 animate-pulse rounded"></div></td>
                                <td class="px-3 py-3"><div class="h-8 w-48 bg-white/10 animate-pulse rounded"></div></td>
                                <td class="px-3 py-3"><div class="h-8 w-16 bg-white/10 animate-pulse rounded"></div></td>
                            </tr>
                        {/each}
                    {:else}
                        {#each users as u}
                            <tr
                                class="border-t border-white/6 hover:bg-white/6 hover:cursor-pointer"
                                onclick={() => openUser(u)}
                                role="button"
                            >
                                <td class="px-3 py-2 text-xs text-white/70">{u.id}</td>
                                <td class="px-3 py-2 text-sm text-white/90">{u.username}</td>
                                <td class="px-3 py-2">
                                    {#if u.role === 'admin'}
                                        <span class="inline-block px-2 py-0.5 rounded-full text-xs font-medium bg-indigo-600 text-white">ADMIN</span>
                                    {:else}
                                        <span class="inline-block px-2 py-0.5 rounded-full text-xs font-medium bg-slate-600 text-white">USER</span>
                                    {/if}
                                </td>
                                <td class="px-3 py-2 text-xs text-white/70">
                                    {#if u.last_activity_at}
                                        {new Date(u.last_activity_at).toLocaleString()}
                                    {:else}
                                        —
                                    {/if}
                                </td>
                                <td class="px-3 py-2 text-xs text-white/90">{u.games_count}</td>
                            </tr>
                        {/each}
                        {#if users.length === 0}
                            <tr><td colspan="5" class="px-3 py-4 text-center text-white/60">Aucun utilisateur</td></tr>
                        {/if}
                    {/if}
                </tbody>
            </table>
        </div>
    </section>
</div>