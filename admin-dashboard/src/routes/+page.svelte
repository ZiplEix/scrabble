<script lang="ts">
    import { goto } from '$app/navigation';
    import LoginForm from '$lib/components/LoginForm.svelte';
    import { user } from '$lib/stores/user';
</script>

<div class="min-h-screen bg-gradient-to-b from-slate-900 via-slate-900/80 to-slate-900 text-white flex items-center">
    <div class="w-full max-w-3xl mx-auto px-4 sm:px-6 lg:px-8 py-24">
        <div class="bg-white/5 rounded-xl p-8 shadow-lg">
            <div class="text-center">
                <h1 class="text-3xl font-bold mb-2">Scrabble Admin</h1>
                <p class="text-sm text-white/70 mb-6">Connexion administrateur pour gérer utilisateurs, parties et rapports.</p>
            </div>

            {#if $user}
                <div class="flex flex-col items-center gap-4">
                    <div class="text-lg">Vous êtes connecté en tant que <strong>{$user.username}</strong></div>
                    <div class="flex gap-3">
                        <a href="/dashboard" class="px-4 py-2 bg-gradient-to-r from-blue-500 to-violet-500 rounded text-white">Aller au dashboard</a>
                        <a href="/" on:click|preventDefault={() => { user.set(null); }} class="px-4 py-2 bg-white/5 rounded">Se déconnecter</a>
                    </div>
                </div>
            {:else}
                <div class="grid grid-cols-1 md:grid-cols-2 gap-6">
                    <div>
                        <LoginForm on:success={() => { goto('/dashboard') }} />
                    </div>

                    <div class="space-y-4">
                        <div class="p-4 bg-white/3 rounded">
                            <h3 class="font-semibold">Ressources rapides</h3>
                            <ul class="mt-3 space-y-2 text-sm text-white/70">
                                <li><a href="/docs" class="text-blue-300 hover:underline">Documentation API</a></li>
                                <li><a href="/support" class="text-blue-300 hover:underline">Support</a></li>
                                <li><a href="/register" class="text-blue-300 hover:underline">Créer un compte</a></li>
                            </ul>
                        </div>

                        <div class="p-4 bg-white/3 rounded">
                            <h4 class="font-semibold">Besoin d'aide ?</h4>
                            <p class="text-sm text-white/70 mt-2">Contacte l'équipe technique ou consulte les logs du système si tu as un problème de connexion.</p>
                        </div>
                    </div>
                </div>
            {/if}
        </div>
    </div>
</div>
