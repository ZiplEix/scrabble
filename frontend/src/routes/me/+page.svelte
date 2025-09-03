<script lang="ts">
    import { onMount } from 'svelte';
    import { user } from '$lib/stores/user';
    import { api } from '$lib/api';
    import UserStats from '$lib/components/UserStats.svelte';
    import MeAccount from '$lib/components/MeAccount.svelte';
    import MeOptions from '$lib/components/MeOptions.svelte';
    import { defaultUserInfos, type UserInfos } from '$lib/types/user_infos';

    let active: 'stats' | 'account' | 'options' = $state('stats');

    let userInfos = $state<UserInfos>(defaultUserInfos);

    onMount(async () => {
        try {
            const resp = await api.get('/me');
            userInfos = resp.data as UserInfos;
            console.log('Fetched /me:', userInfos);

            user.update(u => ({ ...(u as any), id: userInfos!.id, role: userInfos!.role, created_at: userInfos!.created_at }));
        } catch (e) {
            // ignore - interceptor will redirect on 401
            console.error('failed to fetch /auth/me', e);
        }
    });
</script>

<main class="max-w-md mx-auto px-4 py-6">
    <div class="flex items-center gap-2 mb-3">
        <a href="/" class="p-2 rounded hover:bg-gray-100" aria-label="Accueil">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none"><path d="M4 12h16M10 6l-6 6 6 6" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>
        </a>
        <h1 class="text-2xl font-semibold">Mon profil</h1>
    </div>

    <!-- Tabs (segmented control like home) -->
    <div class="mb-5">
        <div class="w-full max-w-md mx-auto">
            <div class="w-full rounded-full bg-white ring-1 ring-black/5 p-1 shadow-sm">
                <div class="grid grid-cols-3 gap-1">
                    <button
                        class="w-full text-center px-3 py-1.5 text-sm rounded-full transition {active==='stats' ? 'bg-emerald-600 text-white shadow font-bold' : 'text-gray-700 hover:bg-gray-50'}"
                        aria-pressed={active==='stats'}
                        onclick={() => (active = 'stats')}
                    >
                        Statistiques
                    </button>
                    <button
                        class="w-full text-center px-3 py-1.5 text-sm rounded-full transition {active==='account' ? 'bg-emerald-600 text-white shadow font-bold' : 'text-gray-700 hover:bg-gray-50'}"
                        aria-pressed={active==='account'}
                        onclick={() => (active = 'account')}
                    >
                        Compte
                    </button>
                    <button
                        class="w-full text-center px-3 py-1.5 text-sm rounded-full transition {active==='options' ? 'bg-emerald-600 text-white shadow font-bold' : 'text-gray-700 hover:bg-gray-50'}"
                        aria-pressed={active==='options'}
                        onclick={() => (active = 'options')}
                    >
                        Options
                    </button>
                </div>
            </div>
        </div>
    </div>

    <!-- Content -->
    <section>
        {#if active === 'stats'}
            <UserStats userInfos={userInfos} />
        {:else if active === 'account'}
            <MeAccount user={$user} />
        {:else}
            <MeOptions bind:userInfos />
        {/if}
    </section>
</main>
