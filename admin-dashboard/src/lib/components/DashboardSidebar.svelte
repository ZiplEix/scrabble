<script lang="ts">
    import { logout } from '$lib/stores/user';
</script>

<nav class="hidden md:flex fixed left-0 top-0 h-screen w-64 bg-white/5 p-4 flex-col justify-between">
    <div>
        <div class="text-sm font-medium">Navigation</div>
        <ul class="mt-2 space-y-1">
            <li><a href="/dashboard" class="block px-3 py-2 rounded hover:bg-white/5">Accueil</a></li>
            <li><a href="/dashboard/logs" class="block px-3 py-2 rounded hover:bg-white/5">Logs</a></li>
            <li><a href="/dashboard/users" class="block px-3 py-2 rounded hover:bg-white/5">Utilisateurs</a></li>
            <li><a href="/dashboard/games" class="block px-3 py-2 rounded hover:bg-white/5">Parties</a></li>
            <li><a href="/dashboard/tickets" class="block px-3 py-2 rounded hover:bg-white/5">Tickets</a></li>
        </ul>
    </div>

    <div class="pt-4">
        <button class="w-full text-left px-3 py-2 rounded bg-white/6 hover:bg-white/8 text-sm" on:click={() => logout()}>Se déconnecter</button>
    </div>
</nav>
