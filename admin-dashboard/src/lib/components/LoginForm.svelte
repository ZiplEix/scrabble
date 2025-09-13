<script lang="ts">
    import { api } from '$lib/api';
  import { user } from '$lib/stores/user';
    import { createEventDispatcher } from 'svelte';
    const dispatch = createEventDispatcher();

    let username = $state('');
    let password = $state('');
    let error: string | null = $state(null);

    async function submit(e: Event) {
        e.preventDefault();
        error = null;
        try {
            const userNameToStore = username.trim().toLowerCase();
            console.log('Attempting login for', userNameToStore);
            const res = await api.post('/auth/admin/login', { username: userNameToStore, password });
            user.set({ username: userNameToStore, token: res.data.token });
            dispatch('success');
        } catch (err) {
            error = 'Échec de la connexion';
        }
    }
</script>

<form class="space-y-3" onsubmit={submit}>
    {#if error}
        <div class="text-sm text-red-300">{error}</div>
    {/if}

    <div>
        <label for="login-username" class="block text-sm text-white/80">Nom d'utilisateur</label>
        <input id="login-username" bind:value={username} class="w-full mt-1 px-3 py-2 rounded bg-white/6 text-white placeholder-white/60 focus:outline-none focus:ring-2 focus:ring-indigo-400" required />
    </div>

    <div>
        <label for="login-password" class="block text-sm text-white/80">Mot de passe</label>
        <input id="login-password" type="password" bind:value={password} class="w-full mt-1 px-3 py-2 rounded bg-white/6 text-white placeholder-white/60 focus:outline-none focus:ring-2 focus:ring-indigo-400" required />
    </div>

    <div class="flex items-center justify-end">
        <button type="submit" class="bg-gradient-to-r from-blue-500 to-violet-500 px-3 py-2 rounded text-white">Se connecter</button>
    </div>
</form>
