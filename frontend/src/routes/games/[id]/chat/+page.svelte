<script lang="ts">
    import { onMount } from 'svelte';
    import { page } from '$app/stores';
    import { api } from '$lib/api';
    import { user } from '$lib/stores/user';
    import { gameStore } from '$lib/stores/game';
    import { get, writable } from 'svelte/store';
    import { goto } from '$app/navigation';
    import { tick } from 'svelte';
    import ChatBubble from '$lib/components/ChatBubble.svelte';

    const gameId = get(page).params.id;

    const messages = writable<any[]>([]);
    let content = '';
    let sending = false;
    let loading = writable(true);
    let error = '';
    let usernameMap: Record<string, string> = {};

    let currentGame: GameInfo | null = null;

    function backToGame() {
        goto(`/games/${gameId}`);
    }

    async function loadMessages() {
        try {
            const res = await api.get(`/game/${gameId}/messages`);
            messages.set(res.data || []);
            await tick();
            scrollToBottom();
        } catch (e: any) {
            console.error('failed to load messages', e);
            error = 'Impossible de charger les messages';
        } finally {
            loading.set(false);
        }
    }

    async function send() {
        const trimmed = content.trim();
        if (!trimmed) return;
        sending = true;
        try {
            await api.post(`/game/${gameId}/message`, { content: trimmed, meta: {} });
            content = '';
            await loadMessages();
            await tick();
            scrollToBottom();
            // scroll to bottom handled by DOM after update
        } catch (e: any) {
            error = e?.response?.data?.message || 'Erreur lors de l\'envoi';
        } finally {
            sending = false;
        }
    }

    async function removeMessage(msgId: number) {
        try {
            await api.delete(`/game/${gameId}/messages/${msgId}`);
            await loadMessages();
            await tick();
            scrollToBottom();
        } catch (e) {
            console.error('delete failed', e);
        }
    }

    gameStore.subscribe(g => {
        currentGame = g;
        if (g?.players) {
            usernameMap = {};
            g.players.forEach((p: any) => { usernameMap[String(p.id)] = p.username; });
        }
    });

    let currentUser: any = null;
    user.subscribe(u => { currentUser = u; });

    function scrollToBottom() {
        const el = document.getElementById('messages');
        if (!el) return;
        el.scrollTop = el.scrollHeight;
    }

    onMount(async () => {
        if (!currentGame || currentGame.id !== gameId) {
            try {
                const res = await api.get(`/game/${gameId}`);
                gameStore.set(res.data);
                currentGame = res.data;
                usernameMap = {};
                currentGame?.players?.forEach((p: any) => { usernameMap[String(p.id)] = p.username; });
            } catch (e) {
                console.warn('failed to load game info', e);
            }
        }

        await loadMessages();
        scrollToBottom();
    });

    function extractDayFromDate(date: Date): string {
        return date.toLocaleDateString(undefined, { weekday: 'long', year: 'numeric', month: 'long', day: 'numeric' });
    }

    // build a grouped view of messages with a flag when the day header should be shown
    $: groupedMessages = (() => {
        const out: { msg: any; showDate: boolean; msgDate: Date }[] = [];
        let prev: Date | null = null;
        $messages.forEach((m) => {
            const msgDate = new Date(m.created_at);
            const showDate = !prev || extractDayFromDate(prev) !== extractDayFromDate(msgDate);
            out.push({ msg: m, showDate, msgDate });
            prev = msgDate;
        });
        return out;
    })();
</script>

<div class="flex flex-col overflow-hidden"
	style="height: calc(100dvh - var(--nav-h, 72px));"
>
    <div class="shrink-0 p-2">
        <button class="text-sm text-blue-600 hover:underline " on:click={backToGame}>
            ← Retour à la partie
        </button>
    </div>

    <header class="w-full flex flex-col items-center">
        <div class="flex items-center gap-2">
            <h1 class="text-md font-semibold">Chat</h1>
            <span class="text-[10px] px-2 py-0.5 bg-yellow-50 text-yellow-800 rounded-full border border-yellow-100">Bêta</span>
        </div>
        <div class="mt-1">
            <p class="text-[11px] text-gray-500 opacity-80 text-center max-w-[80vw]">En développement — signalez les problèmes si vous en rencontrez.</p>
        </div>
    </header>

    <div class="flex-1 overflow-y-auto px-4" id="messages">
        {#each groupedMessages as item}
            {#if item.msg}
                {#if item.showDate}
                    <div class="w-full flex justify-center my-3">
                        <div class="inline-block px-3 py-1 bg-gray-100 text-xs text-gray-700 rounded-full shadow-sm border border-gray-200 max-w-[90%] text-center">
                            {extractDayFromDate(item.msgDate)}
                        </div>
                    </div>
                {/if}
                {@const msg_username = usernameMap[String(item.msg.user_id)] ?? item.msg.username ?? '–'}
                {@const current_username = currentUser?.username ?? '–'}
                <ChatBubble
                    msg={item.msg}
                    username={msg_username}
                    isOwn={msg_username === current_username}
                    onDelete={item.msg.user_id === currentUser?.id ? removeMessage : null}
                />
            {/if}
        {/each}
    </div>

    <div
        class="shrink-0 bg-white px-4 pt-3"
        style="padding-bottom: calc(env(safe-area-inset-bottom) + 12px);"
    >
        <div class="flex items-center gap-2">
            <textarea
                bind:value={content}
                rows={1}
                placeholder="Écrire un message..."
                class="flex-1 p-2 border rounded resize-none min-h-[44px]"
                on:keydown={(e) => {
                    if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); send(); }
                }}
            ></textarea>
            <button class="px-4 py-2 bg-green-600 text-white rounded" on:click={send} disabled={sending}>
                Envoyer
            </button>
        </div>
    </div>
</div>
