<script lang="ts">
    import { onMount } from 'svelte';
    import { user } from '$lib/stores/user';
    import { api } from '$lib/api';
    import RankBadge from '$lib/components/RankBadge.svelte';
    import UserStats from '$lib/components/UserStats.svelte';
    import MeAccount from '$lib/components/MeAccount.svelte';
    import MeOptions from '$lib/components/MeOptions.svelte';
    import AchievementsList from '$lib/components/AchievementsList.svelte';
    import ProgressionChart from '$lib/components/ProgressionChart.svelte';
    import { defaultUserInfos, type UserInfos, type FriendInfo, type RatingHistoryResponse } from '$lib/types/user_infos';

    let active: 'stats' | 'progression' | 'achievements' | 'friends' | 'account' | 'options' = $state('stats');

    let userInfos = $state<UserInfos>(defaultUserInfos);
    let friends = $state<FriendInfo[]>([]);

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
        if (active === 'progression') {
            loadRatingHistory(limit);
        }
    });

    async function loadRatingHistory(lim: number) {
        if (!userInfos || userInfos.id === 0) return;
        isLoadingHistory = true;
        try {
            const resp = await api.get(`/user/${userInfos.id}/rating-history?limit=${lim}`);
            ratingHistory = resp.data as RatingHistoryResponse[];
        } catch (e) {
            console.error('Failed to load rating history:', e);
        } finally {
            isLoadingHistory = false;
        }
    }

    onMount(async () => {
        try {
            const [resp, friendsRes] = await Promise.all([
                api.get('/me'),
                api.get('/users/friends')
            ]);
            userInfos = resp.data as UserInfos;
            friends = friendsRes.data;
            console.log('Fetched /me:', userInfos);

            user.update(u => ({ ...(u as any), id: userInfos!.id, role: userInfos!.role, created_at: userInfos!.created_at }));
        } catch (e) {
            // ignore - interceptor will redirect on 401
            console.error('failed to fetch /auth/me or friends', e);
        }
    });

    async function removeFriend(friendId: number) {
        if (!confirm("Voulez-vous vraiment retirer cette personne de vos amis ?")) return;
        try {
            await api.delete(`/users/friends/${friendId}`);
            friends = friends.filter(f => f.id !== friendId);
        } catch (e) {
            console.error('failed to remove friend:', e);
            alert("Erreur lors de la suppression de l'ami");
        }
    }

    const bgGradients = [
        'from-emerald-400 to-brand-emerald',
        'from-blue-400 to-blue-600',
        'from-purple-400 to-purple-600',
        'from-amber-400 to-brand-gold',
        'from-rose-400 to-rose-600'
    ];

    function getAvatarColor(username: string): string {
        let hash = 0;
        for (let i = 0; i < username.length; i++) {
            hash = username.charCodeAt(i) + ((hash << 5) - hash);
        }
        const index = Math.abs(hash) % bgGradients.length;
        return bgGradients[index];
    }

    let unlockedAchievementsCount = $derived(
        userInfos.achievements ? userInfos.achievements.filter(a => a.unlocked).length : 0
    );
    let totalAchievementsCount = $derived(
        userInfos.achievements ? userInfos.achievements.length : 0
    );
</script>

<main class="max-w-5xl mx-auto px-4 py-6">
    <div class="flex items-center gap-2 mb-4">
        <a href="/" class="p-2 rounded hover:bg-gray-100" aria-label="Accueil">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none"><path d="M4 12h16M10 6l-6 6 6 6" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>
        </a>
        <h1 class="text-2xl font-black text-stone-850">Mon profil</h1>
    </div>

    <!-- Profil Header Card -->
    <div class="rounded-3xl bg-gradient-to-br from-white to-stone-50/40 ring-1 ring-black/5 shadow-md p-6 border border-white mb-6">
        <div class="flex flex-col sm:flex-row items-center sm:items-start gap-6">
            <!-- Avatar -->
            <div class="w-20 h-20 rounded-2xl bg-gradient-to-br from-emerald-400 to-brand-emerald flex items-center justify-center text-3xl font-black text-white shadow-md shadow-brand-emerald/10 border border-emerald-300">
                {userInfos.username ? userInfos.username.charAt(0).toUpperCase() : 'U'}
            </div>

            <!-- User Meta -->
            <div class="flex-1 text-center sm:text-left flex flex-col justify-center h-full">
                <div class="flex flex-wrap items-center justify-center sm:justify-start gap-x-4 gap-y-1.5 text-xs text-stone-500 font-bold mb-4">
                    <span class="flex items-center gap-1"><span class="text-stone-300">⚡</span> Rôle : {userInfos.role === 'admin' ? 'Administrateur' : 'Joueur'}</span>
                    <span class="hidden sm:inline text-stone-300">•</span>
                    <span class="flex items-center gap-1"><span class="text-stone-300">📅</span> Membre depuis : {new Date(userInfos.created_at).toLocaleDateString('fr-FR', { dateStyle: 'long' })}</span>
                </div>
                
                <!-- Quick stats summary row inside header -->
                <div class="flex flex-wrap gap-2 justify-center sm:justify-start">
                    <span class="text-xs bg-stone-100/60 px-2.5 py-1.5 rounded-xl border border-stone-200/20 text-stone-600 font-extrabold">
                        🎮 <strong class="text-stone-800">{userInfos.games_count}</strong> {userInfos.games_count > 1 ? 'parties' : 'partie'}
                    </span>
                    <span class="text-xs bg-emerald-50 text-emerald-800 px-2.5 py-1.5 rounded-xl border border-emerald-100/50 font-extrabold">
                        🏆 <strong class="text-emerald-950">{userInfos.victories}</strong> {userInfos.victories > 1 ? 'victoires' : 'victoire'}
                    </span>
                    <span class="text-xs bg-amber-50 text-amber-800 px-2.5 py-1.5 rounded-xl border border-amber-100/50 font-extrabold">
                        🎖️ <strong class="text-amber-950">{unlockedAchievementsCount}/{totalAchievementsCount || 21}</strong> {unlockedAchievementsCount > 1 ? 'succès' : 'succès'}
                    </span>
                    <span class="text-xs bg-purple-50 text-purple-800 px-2.5 py-1.5 rounded-xl border border-purple-100/50 font-extrabold">
                        👥 <strong class="text-purple-950">{friends.length}</strong> {friends.length > 1 ? 'amis' : 'ami'}
                    </span>
                </div>
            </div>
        </div>
    </div>

    <!-- Sleek Tabs Navigation -->
    <div class="mb-6">
        <div class="w-full">
            <div class="w-full rounded-2xl bg-white border border-stone-200/40 p-1.5 shadow-sm overflow-x-auto whitespace-nowrap scrollbar-none">
                <div class="flex gap-1 min-w-max">
                    <button
                        class="px-4 py-2.5 text-xs rounded-xl transition font-extrabold cursor-pointer
                        {active === 'stats' 
                            ? 'bg-brand-emerald text-white shadow-sm shadow-brand-emerald/20' 
                            : 'text-stone-600 hover:bg-stone-50'}"
                        aria-pressed={active === 'stats'}
                        onclick={() => (active = 'stats')}
                    >
                        📊 Tableau de bord
                    </button>
                    <button
                        class="px-4 py-2.5 text-xs rounded-xl transition font-extrabold cursor-pointer
                        {active === 'progression' 
                            ? 'bg-brand-emerald text-white shadow-sm shadow-brand-emerald/20' 
                            : 'text-stone-600 hover:bg-stone-50'}"
                        aria-pressed={active === 'progression'}
                        onclick={() => (active = 'progression')}
                    >
                        📈 Progression
                    </button>
                    <button
                        class="px-4 py-2.5 text-xs rounded-xl transition font-extrabold cursor-pointer
                        {active === 'achievements' 
                            ? 'bg-brand-emerald text-white shadow-sm shadow-brand-emerald/20' 
                            : 'text-stone-600 hover:bg-stone-50'}"
                        aria-pressed={active === 'achievements'}
                        onclick={() => (active = 'achievements')}
                    >
                        🏆 Succès
                    </button>
                    <button
                        class="px-4 py-2.5 text-xs rounded-xl transition font-extrabold cursor-pointer
                        {active === 'friends' 
                            ? 'bg-brand-emerald text-white shadow-sm shadow-brand-emerald/20' 
                            : 'text-stone-600 hover:bg-stone-50'}"
                        aria-pressed={active === 'friends'}
                        onclick={() => (active = 'friends')}
                    >
                        👥 Amis ({friends.length})
                    </button>
                    <button
                        class="px-4 py-2.5 text-xs rounded-xl transition font-extrabold cursor-pointer
                        {active === 'account' 
                            ? 'bg-brand-emerald text-white shadow-sm shadow-brand-emerald/20' 
                            : 'text-stone-600 hover:bg-stone-50'}"
                        aria-pressed={active === 'account'}
                        onclick={() => (active = 'account')}
                    >
                        👤 Compte
                    </button>
                    <button
                        class="px-4 py-2.5 text-xs rounded-xl transition font-extrabold cursor-pointer
                        {active === 'options' 
                            ? 'bg-brand-emerald text-white shadow-sm shadow-brand-emerald/20' 
                            : 'text-stone-600 hover:bg-stone-50'}"
                        aria-pressed={active === 'options'}
                        onclick={() => (active = 'options')}
                    >
                        ⚙️ Options
                    </button>
                </div>
            </div>
        </div>
    </div>

    <!-- Content -->
    <section>
        {#if active === 'stats'}
            <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
                <!-- Statistiques détaillées -->
                <div class="lg:col-span-2 flex flex-col gap-4">
                    <UserStats userInfos={userInfos} />
                </div>

                <!-- Section Amis latérale -->
                <div class="flex flex-col gap-4">
                    <div class="rounded-3xl bg-white border border-stone-200/50 p-5 shadow-sm flex flex-col gap-4">
                        <div class="flex items-center justify-between">
                            <h2 class="text-xs font-black text-stone-855 uppercase tracking-wider">👥 Mes Amis ({friends.length})</h2>
                            <a href="/leaderboard" class="text-xs font-bold text-brand-emerald hover:underline">Classement →</a>
                        </div>

                        {#if friends.length === 0}
                            <div class="py-8 text-center border border-dashed border-stone-200 rounded-2xl bg-stone-50/30 px-4">
                                <span class="text-2xl block mb-1">👥</span>
                                <p class="text-xs text-stone-700 font-extrabold">Aucun ami</p>
                                <p class="text-[10px] text-stone-400 mt-0.5 leading-normal">
                                    Ajoutez-en depuis le classement ou le profil public d'autres joueurs !
                                </p>
                            </div>
                        {:else}
                            <div class="flex flex-col gap-2.5 max-h-[350px] overflow-y-auto pr-1">
                                {#each friends.slice(0, 5) as friend}
                                    <div class="rounded-2xl border border-stone-200/60 p-2.5 bg-stone-50/20 flex items-center justify-between gap-3 hover:shadow-sm transition">
                                        <a
                                            href={`/user/${friend.id}`}
                                            class="flex items-center gap-2.5 min-w-0 flex-grow hover:opacity-80 transition"
                                        >
                                            <div class={`w-8 h-8 rounded-lg bg-gradient-to-br ${getAvatarColor(friend.username)} text-white flex items-center justify-center font-black shadow-sm text-xs shrink-0`}>
                                                {friend.username.charAt(0).toUpperCase()}
                                            </div>
                                            
                                            <div class="min-w-0">
                                                <p class="text-xs font-black text-stone-800 truncate leading-tight">{friend.username}</p>
                                                <div class="flex items-center gap-1 mt-0.5">
                                                    <RankBadge rating={friend.rating} size="sm" />
                                                    <span class="text-[10px] font-bold text-stone-500">{friend.rating} IPS</span>
                                                </div>
                                            </div>
                                        </a>

                                        <button
                                            type="button"
                                            onclick={() => removeFriend(friend.id)}
                                            class="shrink-0 w-6 h-6 rounded-full border border-stone-200 hover:border-red-200 bg-white hover:bg-rose-50 text-stone-400 hover:text-rose-600 flex items-center justify-center text-[10px] transition active:scale-90 cursor-pointer"
                                            title="Retirer de mes amis"
                                        >
                                            ✕
                                        </button>
                                    </div>
                                {/each}
                                {#if friends.length > 5}
                                    <button 
                                        onclick={() => (active = 'friends')}
                                        class="text-center py-2 text-xs font-bold text-brand-emerald hover:text-emerald-700 transition cursor-pointer"
                                    >
                                        Voir les {friends.length - 5} autres amis →
                                    </button>
                                {/if}
                            </div>
                        {/if}
                    </div>
                </div>
            </div>
        {:else if active === 'progression'}
            <div class="flex flex-col gap-6">
                <!-- Header with Period Filters -->
                <div class="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
                    <div>
                        <h2 class="text-lg font-black text-stone-850">Historique d'IPS & Progression</h2>
                        <p class="text-xs text-stone-500 font-medium">Suivez l'évolution de vos points IPS partie après partie.</p>
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
                            <span class="text-xs text-stone-400 font-extrabold">Chargement de votre courbe...</span>
                        </div>
                    </div>
                {:else}
                    <ProgressionChart history={ratingHistory} />
                {/if}
            </div>
        {:else if active === 'achievements'}
            <AchievementsList achievements={userInfos.achievements} />
        {:else if active === 'friends'}
            <div class="rounded-3xl bg-white border border-stone-200/50 p-6 shadow-sm flex flex-col gap-4">
                <div class="flex items-center justify-between">
                    <h2 class="text-lg font-black text-stone-800">Mes Amis ({friends.length})</h2>
                </div>

                {#if friends.length === 0}
                    <div class="py-12 text-center border border-dashed border-stone-200 rounded-2xl bg-stone-50/30">
                        <span class="text-3xl block mb-2">👥</span>
                        <p class="text-sm text-stone-700 font-extrabold">Vous n'avez pas encore d'amis</p>
                        <p class="text-xs text-stone-400 mt-1 max-w-xs mx-auto leading-normal">
                            Visitez le profil public des autres joueurs ou consultez le classement général pour les ajouter en ami en un clic !
                        </p>
                    </div>
                {:else}
                    <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
                        {#each friends as friend}
                            <div class="rounded-2xl border border-stone-200/60 p-3 bg-stone-50/30 flex items-center justify-between gap-3 shadow-sm hover:shadow transition">
                                <a
                                    href={`/user/${friend.id}`}
                                    class="flex items-center gap-3 min-w-0 flex-grow hover:opacity-80 transition"
                                >
                                    <!-- Avatar -->
                                    <div class={`w-10 h-10 rounded-xl bg-gradient-to-br ${getAvatarColor(friend.username)} text-white flex items-center justify-center font-black shadow-sm shrink-0`}>
                                        {friend.username.charAt(0).toUpperCase()}
                                    </div>
                                    
                                    <div class="min-w-0">
                                        <p class="text-sm font-black text-stone-800 truncate">{friend.username}</p>
                                        <div class="flex items-center gap-1 mt-0.5">
                                            <RankBadge rating={friend.rating} size="sm" />
                                            <span class="text-[10px] font-bold text-stone-500">{friend.rating} IPS</span>
                                        </div>
                                    </div>
                                </a>

                                <button
                                    type="button"
                                    onclick={() => removeFriend(friend.id)}
                                    class="shrink-0 w-8 h-8 rounded-full border border-stone-200 hover:border-red-200 bg-white hover:bg-rose-50 text-stone-400 hover:text-rose-600 flex items-center justify-center transition active:scale-90 cursor-pointer"
                                    title="Retirer de mes amis"
                                    aria-label="Retirer de mes amis"
                                >
                                    ✕
                                </button>
                            </div>
                        {/each}
                    </div>
                {/if}
            </div>
        {:else if active === 'account'}
            <MeAccount user={$user} />
        {:else}
            <MeOptions bind:userInfos />
        {/if}
    </section>
</main>

