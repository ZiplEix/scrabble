<script lang="ts">
    import { logout } from '$lib/stores/user';
</script>

<nav class="hidden md:flex fixed left-0 top-0 h-screen w-64 bg-white/5 p-4 flex-col justify-between">
    <div>
        <div class="text-sm font-medium">Navigation</div>
        <ul class="mt-2 space-y-1">
            <li>
                <a href="/dashboard" class="px-3 py-2 rounded hover:bg-white/5 flex items-center gap-3">
                    <svg class="w-5 h-5 text-white/90" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg"><path d="M3 11.5L12 4l9 7.5V20a1 1 0 0 1-1 1h-5v-6H9v6H4a1 1 0 0 1-1-1v-8.5z" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/></svg>
                    <span>Accueil</span>
                </a>
            </li>
            <li>
                <a href="/dashboard/logs" class="px-3 py-2 rounded hover:bg-white/5 flex items-center gap-3">
                    <svg class="w-5 h-5 text-white/90" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg"><path d="M8 7h8M8 12h8M8 17h5" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>
                    <span>Logs</span>
                </a>
            </li>
            <li>
                <a href="/dashboard/users" class="px-3 py-2 rounded hover:bg-white/5 flex items-center gap-3">
                    <svg class="w-5 h-5 text-white/90" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg"><path d="M17 21v-2a4 4 0 0 0-4-4H9a4 4 0 0 0-4 4v2" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/><circle cx="12" cy="7" r="4" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/></svg>
                    <span>Utilisateurs</span>
                </a>
            </li>
            <li>
                <a href="/dashboard/games" class="px-3 py-2 rounded hover:bg-white/5 flex items-center gap-3">
                    <svg class="w-5 h-5 text-white/90" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg"><path d="M5 3v18l15-9L5 3z" stroke="currentColor" stroke-width="1.2" stroke-linecap="round" stroke-linejoin="round"/></svg>
                    <span>Parties</span>
                </a>
            </li>
            <li>
                <a href="/dashboard/tickets" class="px-3 py-2 rounded hover:bg-white/5 flex items-center gap-3">
                    <svg class="w-5 h-5 text-white/90" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg"><path d="M21 10v4a2 2 0 0 1-2 2h-2v-8h2a2 2 0 0 1 2 2zM7 6H5a2 2 0 0 0-2 2v4a2 2 0 0 0 2 2h2V6zM9 6h6v12H9V6z" stroke="currentColor" stroke-width="1" stroke-linecap="round" stroke-linejoin="round"/></svg>
                    <span>Tickets</span>
                </a>
            </li>
        </ul>
    </div>

    <div class="pt-4">
        <button class="w-full text-left px-3 py-2 rounded bg-white/6 hover:bg-white/8 text-sm" on:click={() => logout()}>Se déconnecter</button>
    </div>
</nav>
