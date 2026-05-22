<script lang="ts">
    import { onMount } from 'svelte';
    import { page } from '$app/stores';
    import { api } from '$lib/api';
    import { goto } from '$app/navigation';
    import { user } from '$lib/stores/user';
    import { defaultUserInfos, type UserInfos } from '$lib/types/user_infos';
    import RankBadge from '$lib/components/RankBadge.svelte';
    import UserStats from '$lib/components/UserStats.svelte';
    import HeadToHeadView from '$lib/components/HeadToHeadView.svelte';
    import AchievementsList from '$lib/components/AchievementsList.svelte';

    let loading = $state(true);
    let error = $state<string | null>(null);
    let userInfos = $state<UserInfos>(defaultUserInfos as any);
    let activeTab = $state<'stats' | 'h2h' | 'achievements'>('stats');

    // Vérifier si on doit afficher le face-à-face
    let showH2H = $derived(
        userInfos && userInfos.id !== 0 && $user && $user.id !== userInfos.id && userInfos.head_to_head
    );

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

<main class="max-w-4xl mx-auto px-4 pt-6 pb-12">
    <!-- Header Navigation -->
    <div class="mb-6 flex items-center gap-3">
        <button 
            class="p-2.5 rounded-2xl bg-white border border-stone-200/60 hover:bg-stone-50 active:scale-95 transition-all shadow-sm cursor-pointer" 
            aria-label="Retour" 
            title="Retour" 
            onclick={goBack}
        >
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" class="text-stone-600">
                <path d="M19 12H5M12 19l-7-7 7-7" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
        </button>
        <div>
            <span class="text-[10px] uppercase font-black tracking-wider text-stone-400">Profil de Joueur</span>
            <h1 class="text-xl font-black text-stone-800">
                {#if loading}
                    Chargement...
                {:else}
                    {userInfos.username}
                {/if}
            </h1>
        </div>
    </div>

    {#if loading}
        <div class="flex flex-col items-center justify-center py-20 gap-3">
            <div class="w-10 h-10 border-4 border-brand-emerald border-t-transparent rounded-full animate-spin"></div>
            <p class="text-stone-400 text-sm font-bold animate-pulse">Chargement du profil…</p>
        </div>
    {:else if error}
        <div class="bg-rose-50 border border-rose-200 rounded-2xl p-6 text-center shadow-sm">
            <span class="text-3xl block mb-2">⚠️</span>
            <h3 class="font-extrabold text-rose-800 text-sm mb-1">Erreur de chargement</h3>
            <p class="text-xs text-rose-600">{error}</p>
        </div>
    {:else}
        <div class="flex flex-col gap-6">
            <!-- Profil Header Card -->
            <div class="rounded-3xl bg-gradient-to-br from-white to-stone-50/40 ring-1 ring-black/5 shadow-md p-6 border border-white">
                <div class="flex flex-col sm:flex-row items-center sm:items-start gap-6">
                    <!-- Avatar -->
                    <div class="w-20 h-20 rounded-2xl bg-gradient-to-br from-emerald-400 to-brand-emerald flex items-center justify-center text-3xl font-black text-white shadow-md shadow-brand-emerald/10 border border-emerald-300">
                        {userInfos.username ? userInfos.username.charAt(0).toUpperCase() : 'U'}
                    </div>

                    <!-- User Meta -->
                    <div class="flex-1 text-center sm:text-left flex flex-col justify-center h-full">
                        <div class="flex flex-col sm:flex-row sm:items-center gap-2 mb-2 justify-center sm:justify-start">
                            <h2 class="text-2xl font-black text-stone-800 leading-none">{userInfos.username}</h2>
                            <div class="inline-flex items-center gap-1.5 bg-stone-100/80 px-2.5 py-1 rounded-full border border-stone-200/30 self-center sm:self-auto">
                                <RankBadge rating={userInfos.rating} size="sm" />
                                <span class="text-xs font-black text-stone-600">{userInfos.rating} Elo</span>
                            </div>
                        </div>
                        <div class="flex flex-wrap items-center justify-center sm:justify-start gap-x-4 gap-y-1.5 text-xs text-stone-500 font-bold">
                            <span class="flex items-center gap-1"><span class="text-stone-300">⚡</span> Rôle : {userInfos.role === 'admin' ? 'Administrateur' : 'Joueur'}</span>
                            <span class="hidden sm:inline text-stone-300">•</span>
                            <span class="flex items-center gap-1"><span class="text-stone-300">📅</span> Membre depuis : {new Date(userInfos.created_at).toLocaleDateString('fr-FR', { dateStyle: 'long' })}</span>
                        </div>
                    </div>
                </div>
            </div>

            <!-- Sleek Tabs Navigation -->
            <div class="w-full max-w-lg mx-auto">
                <div class="rounded-2xl bg-white border border-stone-200/40 p-1.5 shadow-sm flex gap-1">
                    <button
                        class="flex-1 text-center py-2.5 text-xs rounded-xl transition font-extrabold cursor-pointer
                        {activeTab === 'stats' 
                            ? 'bg-brand-emerald text-white shadow-sm shadow-brand-emerald/20' 
                            : 'text-stone-600 hover:bg-stone-50'}"
                        aria-pressed={activeTab === 'stats'}
                        onclick={() => (activeTab = 'stats')}
                    >
                        📈 Statistiques
                    </button>
                    {#if showH2H}
                        <button
                            class="flex-1 text-center py-2.5 text-xs rounded-xl transition font-extrabold cursor-pointer
                            {activeTab === 'h2h' 
                                ? 'bg-brand-emerald text-white shadow-sm shadow-brand-emerald/20' 
                                : 'text-stone-600 hover:bg-stone-50'}"
                            aria-pressed={activeTab === 'h2h'}
                            onclick={() => (activeTab = 'h2h')}
                        >
                            ⚔️ Face-à-Face
                        </button>
                    {/if}
                    <button
                        class="flex-1 text-center py-2.5 text-xs rounded-xl transition font-extrabold cursor-pointer
                        {activeTab === 'achievements' 
                            ? 'bg-brand-emerald text-white shadow-sm shadow-brand-emerald/20' 
                            : 'text-stone-600 hover:bg-stone-50'}"
                        aria-pressed={activeTab === 'achievements'}
                        onclick={() => (activeTab = 'achievements')}
                    >
                        🏆 Succès
                    </button>
                </div>
            </div>

            <!-- Tab Content -->
            <section class="transition-all duration-300">
                {#if activeTab === 'stats'}
                    <UserStats {userInfos} />
                {:else if activeTab === 'h2h' && userInfos.head_to_head}
                    <HeadToHeadView headToHead={userInfos.head_to_head} opponentName={userInfos.username} />
                {:else if activeTab === 'achievements'}
                    <AchievementsList achievements={userInfos.achievements} />
                {/if}
            </section>
        </div>
    {/if}
</main>

