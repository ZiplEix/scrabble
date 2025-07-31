<script lang='ts'>
  import { goto } from '$app/navigation';
    import { user } from '$lib/stores/user';
</script>

<nav class="flex justify-between items-center px-6 py-4 bg-white shadow-md">
	<a href="/" class="text-2xl font-bold text-green-700 tracking-tight">🧩 Scrabble</a>
	<div class="flex space-x-6 items-center">
		{#if !$user}
			<a href="/login" class="bg-green-600 text-white px-4 py-2 rounded hover:bg-green-700">Connexion</a>
		{:else}
            <button
                class="bg-green-600 text-white px-4 py-2 rounded hover:bg-green-700"
                on:click={() => {
                    user.set(null);
                    localStorage.removeItem('user');
                    goto('/')
                }}
            >
                Déconnexion
            </button>
		{/if}
	</div>
</nav>
