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
    import HeaderBar from '$lib/components/HeaderBar.svelte';
    import type { GameInfo } from '$lib/types/game_infos';

    const gameId = get(page).params.id;

    const messages = writable<any[]>([]);
    let content = '';
    let sending = false;
    let loading = writable(true);
    let error = '';
    let usernameMap: Record<string, string> = {};

    let currentGame: GameInfo | null = null;

    function backToGame() { goto(`/games/${gameId}`); }

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

    async function markMessagesRead() {
        try {
            const msgs = get(messages) as any[];
            const lastId = msgs.length ? msgs[msgs.length - 1].id : 0;
            if (lastId) {
                await api.post(`/game/${gameId}/messages/read`, { last_message_id: lastId });
            } else {
                // no messages -> still call endpoint to create row if needed (server will noop)
                await api.post(`/game/${gameId}/messages/read`, {});
            }
        } catch (e) {
            console.warn('failed to mark messages read', e);
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
            await markMessagesRead();
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
            await markMessagesRead();
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
        await markMessagesRead();
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

    // info modal state
    let showInfo = false;
    function openInfo() { showInfo = true; }
    function closeInfo() { showInfo = false; }
</script>

<div class="flex flex-col overflow-hidden"
	style="height: 100dvh;"
>
    <HeaderBar title="Chat" back={true} />
    <div class="px-3 py-2 flex justify-end">
        <button class="p-1 rounded hover:bg-gray-100" aria-label="Infos" title="Informations" onclick={openInfo}>
            <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5 text-gray-600" viewBox="0 0 20 20" fill="currentColor" aria-hidden="true">
                <path fill-rule="evenodd" d="M18 10A8 8 0 11 2 10a8 8 0 0116 0zm-9-1a1 1 0 112 0v5a1 1 0 11-2 0V9zm1-4a1.25 1.25 0 100 2.5A1.25 1.25 0 0010 5z" clip-rule="evenodd" />
            </svg>
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
                    userId={item.msg.user_id}
                    isOwn={msg_username === current_username}
                    onDelete={item.msg.user_id === currentUser?.id ? removeMessage : null}
                />
            {/if}
        {/each}
    </div>

    {#if showInfo}
        <div class="fixed inset-0 z-50 flex items-center justify-center p-4">
            <div class="absolute inset-0 bg-black/40" onclick={closeInfo} aria-hidden="true"></div>
            <div class="relative max-w-xl w-full bg-white rounded-lg shadow-lg ring-1 ring-black/5">
                <div class="flex items-center justify-between px-4 py-3 border-b">
                    <h2 class="text-sm font-semibold">À propos du chat (Bêta)</h2>
                    <button class="text-gray-500 hover:text-gray-700 p-1" aria-label="Fermer" onclick={closeInfo}>&times;</button>
                </div>
                <div class="px-4 py-3 text-sm text-gray-700 space-y-2">
                    <p>Ce chat est en version Bêta. Il peut contenir des bugs — n'hésitez pas à signaler tout problème.</p>
                    <p class="font-medium">Formatage léger supporté :</p>
                    <ul class="list-none ml-0 space-y-1">
                        <li><span class="font-mono">**gras**</span> → <b>gras</b></li>
                        <li><span class="font-mono">*italique*</span> → <i>italique</i></li>
                        <li><span class="font-mono">__souligné__</span> → <u>souligné</u></li>
                        <li><span class="font-mono">~~barré~~</span> → <s>barré</s></li>
                        <li><span class="font-mono">`code`</span> → <code>inline code</code></li>
                    </ul>
                    <p>Les liens (http/https/www) sont détectés automatiquement. Les sauts de ligne sont préservés.</p>
                    <p class="text-xs text-gray-500">Astuce : le texte est automatiquement nettoyé à l'envoi (trim) et la ponctuation française est protégée.</p>
                </div>
                <div class="px-4 py-3 border-t flex justify-end gap-2">
                    <button class="px-3 py-1 text-sm bg-gray-100 rounded" onclick={closeInfo}>Fermer</button>
                </div>
            </div>
        </div>
    {/if}

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
                onkeydown={(e) => {
                    if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); send(); }
                }}
            ></textarea>
            <button class="px-4 py-2 bg-green-600 text-white rounded" onclick={send} disabled={sending}>
                Envoyer
            </button>
        </div>
    </div>
</div>
