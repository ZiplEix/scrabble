<script lang="ts">
    import { onMount } from 'svelte';
    import { page } from '$app/stores';
    import { api } from '$lib/api';
    import { goto } from '$app/navigation';
    import { defaultUserInfos, type UserInfos } from '$lib/types/user_infos';
    import UserStats from '$lib/components/UserStats.svelte';

    let loading = true;
    let error: string | null = null;
    let userInfos: UserInfos = defaultUserInfos as any;

    onMount(async () => {
        const id = $page.params.ID;
        try {
            const res = await api.get(`/user/${id}`);
            userInfos = res.data as UserInfos;
        } catch (e: any) {
            console.error('failed to fetch user public:', e);
            error = e?.response?.data?.error || 'Erreur lors du chargement';
        } finally {
            loading = false;
        }
    });

    function goBack() {
        // client-side only
        if (typeof window !== 'undefined') {
            try {
                if (window.history.length > 1) {
                    window.history.back();
                    return;
                }
            } catch (e) {
                // ignore and fallback
            }
        }
        goto('/');
    }
</script>

<main class="max-w-4xl mx-auto px-4 pt-8">
    <div class="mb-6 flex items-center gap-2">
        <button class="p-2 rounded hover:bg-gray-100" aria-label="Retour" title="Retour" onclick={goBack}>
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none"><path d="M4 12h16M10 6l-6 6 6 6" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>
        </button>
        <h1 class="text-2xl font-semibold">Profil public de {userInfos.username}</h1>
    </div>

    {#if loading}
        <p class="text-center text-gray-500">Chargement…</p>
    {:else if error}
        <p class="text-center text-red-600">{error}</p>
    {:else}
        <div class="mx-auto">
            <div class="rounded-sm ring-1 ring-black/5 bg-white shadow p-4">
                <div class="grid grid-cols-1 md:grid-cols-3 gap-6 items-start">
                    <!-- Left column: basic info -->
                    <div class="flex flex-col items-start">
                        <div class="w-full flex items-center gap-4">
                            <div class="w-16 h-16 rounded-full bg-emerald-100 flex items-center justify-center text-2xl font-bold text-emerald-700">{userInfos.username ? userInfos.username.charAt(0).toUpperCase() : 'U'}</div>
                            <div>
                                <div class="text-xl font-semibold">{userInfos.username}</div>
                                <div class="text-sm text-gray-500">Rôle : {userInfos.role}</div>
                                <div class="text-sm text-gray-500">Inscrit le : {new Date(userInfos.created_at).toLocaleDateString()}</div>
                            </div>
                        </div>
                    </div>

                    <!-- Right columns: stats (span 2 on md) -->
                    <div class="md:col-span-2">
                        <UserStats {userInfos} />
                    </div>
                </div>
            </div>
        </div>
    {/if}
</main>
