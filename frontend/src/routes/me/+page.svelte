<script lang="ts">
    import { onMount } from 'svelte';
    import { user } from '$lib/stores/user';
    import { api } from '$lib/api';
    import MeStats from '$lib/components/MeStats.svelte';
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

    <!-- Tabs -->
    <div class="flex items-center gap-3 mb-5">
        <button
            onclick={() => (active = 'stats')}
            class="flex-1 text-sm py-3 rounded-lg text-center {active === 'stats' ? 'bg-green-600 text-white' : 'bg-gray-100 text-gray-700'}"
        >
            Statistiques
        </button>
        <button
            onclick={() => (active = 'account')}
            class="flex-1 text-sm py-3 rounded-lg text-center {active === 'account' ? 'bg-green-600 text-white' : 'bg-gray-100 text-gray-700'}"
        >
            Compte
        </button>
        <button
            onclick={() => (active = 'options')}
            class="flex-1 text-sm py-3 rounded-lg text-center {active === 'options' ? 'bg-green-600 text-white' : 'bg-gray-100 text-gray-700'}"
        >
            Options
        </button>
    </div>

    <!-- Content -->
    <section class="bg-white shadow-sm rounded-lg p-6">
        {#if active === 'stats'}
            <MeStats userInfos={userInfos} />
        {:else if active === 'account'}
            <MeAccount user={$user} />
        {:else}
            <MeOptions bind:userInfos />
        {/if}
    </section>
</main>
