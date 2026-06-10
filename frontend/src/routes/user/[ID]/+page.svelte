<script lang="ts">
    import { onMount } from 'svelte';
    import { page } from '$app/stores';
    import { getUserProfile, getUserRatingHistory, addFriend, removeFriend } from '$lib/api';
    import { goto } from '$app/navigation';
    import { user } from '$lib/stores/user';
    import { defaultUserInfos, type UserInfos, type RatingHistoryResponse } from '$lib/types/user_infos';
    import RankBadge from '$lib/components/RankBadge.svelte';
    import UserStats from '$lib/components/UserStats.svelte';
    import HeadToHeadView from '$lib/components/HeadToHeadView.svelte';
    import AchievementsList from '$lib/components/AchievementsList.svelte';
    import ProgressionChart from '$lib/components/ProgressionChart.svelte';

    let loading = $state(true);
    let error = $state<string | null>(null);
    let userInfos = $state<UserInfos>(defaultUserInfos as any);
    let activeTab = $state<'stats' | 'progression' | 'h2h' | 'achievements'>('stats');
    let togglingFriend = $state(false);

    let limit = $state(25);
    let ratingHistory = $state<RatingHistoryResponse[]>([]);
    let isLoadingHistory = $state(false);

    let historyRatings = $derived(ratingHistory.map(h => h.rating));
    let peakRating = $derived(ratingHistory.length > 0 ? Math.max(...historyRatings, userInfos.rating) : userInfos.rating);
    let ratingTrend = $derived(ratingHistory.length > 1 ? ratingHistory[ratingHistory.length - 1].rating - ratingHistory[0].rating : 0);
    let gamesInPeriod = $derived(ratingHistory.filter(h => h.game_info));
    let winsInPeriod = $derived(gamesInPeriod.filter(h => h.game_info?.won).length);
    let winRatePeriod = $derived(gamesInPeriod.length > 0 ? Math.round((winsInPeriod / gamesInPeriod.length) * 100) : null);

    $effect(() => {
        if (activeTab === 'progression') {
            loadRatingHistory(limit);
        }
    });

    async function loadRatingHistory(lim: number) {
        if (!userInfos || userInfos.id === 0) return;
        isLoadingHistory = true;
        try {
            ratingHistory = await getUserRatingHistory(userInfos.id, lim);
        } catch (e) {
            console.error('Failed to load rating history:', e);
        } finally {
            isLoadingHistory = false;
        }
    }

    // Vérifier si on doit afficher le bouton d'ami
    let showFriendBtn = $derived(
        $user && userInfos && userInfos.id !== 0 && $user.id !== userInfos.id
    );

    async function toggleFriend() {
        if (togglingFriend) return;
        togglingFriend = true;
        try {
            if (userInfos.is_friend) {
                await removeFriend(userInfos.id);
                userInfos.is_friend = false;
            } else {
                await addFriend(userInfos.id);
                userInfos.is_friend = true;
            }
        } catch (e) {
            console.error('failed to toggle friend status:', e);
            alert("Erreur lors de la modification du statut d'ami");
        } finally {
            togglingFriend = false;
        }
    }

    // Vérifier si on doit afficher le face-à-face
    let showH2H = $derived(
        userInfos && userInfos.id !== 0 && $user && $user.id !== userInfos.id && userInfos.head_to_head
    );

    onMount(async () => {
        const id = $page.params.ID;
        try {
            userInfos = await getUserProfile(Number(id));
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
                            <div class="flex flex-wrap items-center gap-2 justify-center sm:justify-start">
                                <div class="inline-flex items-center gap-1.5 bg-stone-100/80 px-2.5 py-1 rounded-full border border-stone-200/30">
                                    <RankBadge rating={userInfos.rating} size="sm" />
                                    <span class="text-xs font-black text-stone-600">{userInfos.rating} IPS</span>
                                </div>

                                {#if showFriendBtn}
                                    <button
                                        onclick={toggleFriend}
                                        disabled={togglingFriend}
                                        class="inline-flex items-center gap-1 px-3 py-1 rounded-full text-xs font-extrabold shadow-sm transition active:scale-95 cursor-pointer disabled:opacity-50
                                        {userInfos.is_friend
                                            ? 'bg-stone-100 hover:bg-stone-200 text-stone-700 border border-stone-200'
                                            : 'bg-brand-emerald/10 text-brand-emerald border border-brand-emerald/20 hover:bg-brand-emerald hover:text-white'}"
                                    >
                                        {#if userInfos.is_friend}
                                            <span class="text-xs">👤✓</span> Ami(e)
                                        {:else}
                                            <span class="text-xs">👤+</span> Ajouter en ami
                                        {/if}
                                    </button>
                                {/if}
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
                        📊 Statistiques
                    </button>
                    <button
                        class="flex-1 text-center py-2.5 text-xs rounded-xl transition font-extrabold cursor-pointer
                        {activeTab === 'progression' 
                            ? 'bg-brand-emerald text-white shadow-sm shadow-brand-emerald/20' 
                            : 'text-stone-600 hover:bg-stone-50'}"
                        aria-pressed={activeTab === 'progression'}
                        onclick={() => (activeTab = 'progression')}
                    >
                        📈 Progression
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
                {:else if activeTab === 'progression'}
                    <div class="flex flex-col gap-6">
                        <!-- Header with Period Filters -->
                        <div class="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
                            <div>
                                <h2 class="text-lg font-black text-stone-850">Historique d'IPS & Progression</h2>
                                <p class="text-xs text-stone-500 font-medium">Suivez l'évolution des points IPS de {userInfos.username}.</p>
                            </div>
                            
                            <!-- Period filters -->
                            <div class="flex items-center gap-1 bg-stone-100/80 border border-stone-200/60 p-1 rounded-xl">
                                {#each [10, 25, 50] as pLimit}
                                    <button
                                        type="button"
                                        onclick={() => { limit = pLimit; loadRatingHistory(pLimit); }}
                                        class="px-3 py-1.5 text-[10px] font-black rounded-lg transition cursor-pointer
                                        {limit === pLimit 
                                            ? 'bg-white text-stone-800 shadow-sm' 
                                            : 'text-stone-500 hover:text-stone-850'}"
                                    >
                                        {pLimit} parties
                                    </button>
                                {/each}
                            </div>
                        </div>

                        <!-- KPI Cards Row -->
                        <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
                            <!-- KPI 1: IPS Actuel -->
                            <div class="rounded-2xl bg-white border border-stone-200/50 p-4 shadow-sm flex flex-col justify-between min-h-[90px]">
                                <span class="text-[10px] font-bold text-stone-400 uppercase tracking-wider">IPS Actuel</span>
                                <div class="flex items-baseline gap-1 mt-1">
                                    <span class="text-2xl font-black text-stone-850">{userInfos.rating}</span>
                                    <span class="text-[10px] font-bold text-stone-500">pts</span>
                                </div>
                                <p class="text-[10px] text-stone-500 mt-1 font-medium">Classement en direct</p>
                            </div>

                            <!-- KPI 2: Meilleur IPS -->
                            <div class="rounded-2xl bg-white border border-stone-200/50 p-4 shadow-sm flex flex-col justify-between min-h-[90px]">
                                <span class="text-[10px] font-bold text-stone-400 uppercase tracking-wider">Meilleur IPS</span>
                                <div class="flex items-baseline gap-1 mt-1">
                                    <span class="text-2xl font-black text-emerald-650">{peakRating}</span>
                                    <span class="text-[10px] font-bold text-emerald-500">pts</span>
                                </div>
                                <p class="text-[10px] text-stone-500 mt-1 font-medium">Pic sur la période</p>
                            </div>

                            <!-- KPI 3: Tendance -->
                            <div class="rounded-2xl bg-white border border-stone-200/50 p-4 shadow-sm flex flex-col justify-between min-h-[90px]">
                                <span class="text-[10px] font-bold text-stone-400 uppercase tracking-wider">Tendance</span>
                                <div class="flex items-baseline gap-1 mt-1">
                                    <span class="text-2xl font-black {ratingTrend > 0 ? 'text-emerald-650' : ratingTrend < 0 ? 'text-rose-650' : 'text-stone-700'}">
                                        {ratingTrend > 0 ? '+' : ''}{ratingTrend}
                                    </span>
                                    <span class="text-[10px] font-bold {ratingTrend > 0 ? 'text-emerald-500' : ratingTrend < 0 ? 'text-rose-500' : 'text-stone-400'}">IPS</span>
                                </div>
                                <p class="text-[10px] text-stone-500 mt-1 font-medium">Variation (période)</p>
                            </div>

                            <!-- KPI 4: Taux de Victoire -->
                            <div class="rounded-2xl bg-white border border-stone-200/50 p-4 shadow-sm flex flex-col justify-between min-h-[90px]">
                                <span class="text-[10px] font-bold text-stone-400 uppercase tracking-wider">Taux de victoire</span>
                                <div class="flex items-baseline gap-1 mt-1">
                                    <span class="text-2xl font-black text-brand-emerald">
                                        {winRatePeriod !== null ? `${winRatePeriod}%` : '—'}
                                    </span>
                                </div>
                                <p class="text-[10px] text-stone-500 mt-1 font-medium">Sur les parties jouées</p>
                            </div>
                        </div>

                        <!-- Chart Card -->
                        {#if isLoadingHistory}
                            <div class="rounded-3xl bg-white border border-stone-200/50 p-6 shadow-sm min-h-[280px] flex items-center justify-center">
                                <div class="flex flex-col items-center gap-3">
                                    <div class="w-8 h-8 rounded-full border-2 border-stone-200 border-t-brand-emerald animate-spin"></div>
                                    <span class="text-xs text-stone-400 font-extrabold">Chargement de la courbe...</span>
                                </div>
                            </div>
                        {:else}
                            <ProgressionChart history={ratingHistory} />
                        {/if}
                    </div>
                {:else if activeTab === 'h2h' && userInfos.head_to_head}
                    <HeadToHeadView headToHead={userInfos.head_to_head} opponentName={userInfos.username} />
                {:else if activeTab === 'achievements'}
                    <AchievementsList achievements={userInfos.achievements} />
                {/if}
            </section>
        </div>
    {/if}
</main>

