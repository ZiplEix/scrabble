<script lang="ts">
    import type User from '$lib/types/user';
    import { goto } from '$app/navigation';
    import { logout } from '$lib/stores/user';
    export let user: User | null = null;

    function goToLogin() {
        goto('/login');
    }

    function doLogout() {
        logout();
        goto('/');
    }
</script>

<div>
    <h2 class="text-lg font-medium mb-3">Compte</h2>

    <div class="flex items-center gap-4 mb-4">
        <div class="w-14 h-14 bg-gray-200 rounded-full flex items-center justify-center text-gray-500">👤</div>
        <div>
            <p class="text-md font-medium">{(user && user.username) ?? 'Nom d\'utilisateur'}</p>
            <p class="text-xs text-gray-500">ID: {(user && (user as any).id) ?? '—'}</p>
        </div>
    </div>

    <div class="space-y-2 mb-4">
        <div class="flex items-center justify-between p-3 bg-gray-50 rounded">
            <span class="text-sm text-gray-700">Rôle</span>
            <span class="text-sm font-medium text-gray-900">{(user && (user as any).role) ?? 'user'}</span>
        </div>

        <div class="flex items-center justify-between p-3 bg-gray-50 rounded">
            <span class="text-sm text-gray-700">Inscription</span>
            <span class="text-sm font-medium text-gray-900">{(user && (user as any).created_at) ? new Date((user as any).created_at).toLocaleDateString('fr-FR') : '—'}</span>
        </div>
    </div>

    <div class="space-y-3">
        <button on:click={goToLogin} class="w-full bg-gray-100 text-gray-800 py-3 rounded text-sm">Changer le mot de passe</button>
        <button on:click={doLogout} class="w-full bg-red-100 text-red-800 py-3 rounded text-sm">Se déconnecter</button>
    </div>
</div>
