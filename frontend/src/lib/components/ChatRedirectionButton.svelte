<script lang='ts'>
    import { goto } from "$app/navigation";
    import { api } from "$lib/api";
    import { onMount } from "svelte";

    let { gameId }: { gameId: string } = $props();

    let unread = $state(0);

    async function refreshUnread() {
        try {
            const res = await api.get(`/game/${gameId}/unread_messages_count`);
            unread = Number(res.data?.unread_count || 0);
        } catch (err) {
            console.error('failed to fetch unread count', err);
        }
    }

    onMount(async () => {
        await refreshUnread();
    });
</script>

<button
    class="relative shrink-0 h-8 w-8 rounded-lg bg-amber-200 text-black flex items-center justify-center active:scale-95 transition"
    aria-label="Ouvrir le chat"
    onclick={() => goto(`/games/${gameId}/chat`)}
    title="Chat de la partie"
>
    <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 24 24" fill="none" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 10h.01M12 10h.01M16 10h.01M21 12c0 4.418-4.03 8-9 8a9.77 9.77 0 0 1-4-.88L3 20l1.12-3.11A7.97 7.97 0 0 1 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
    </svg>
    {#if unread > 0}
        <span class="absolute -top-1.5 -right-1.5 h-4 min-w-4 px-1.5 text-[10px] leading-none rounded-full bg-red-500 text-white flex items-center justify-center font-semibold">{unread > 99 ? '99+' : unread}</span>
    {/if}
</button>