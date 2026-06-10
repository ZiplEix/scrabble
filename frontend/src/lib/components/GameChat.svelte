<script lang="ts">
    import { onMount, onDestroy } from 'svelte';
    import { getMessages, markMessagesAsRead, sendMessage, deleteChatMessage } from '$lib/api';
    import { user } from '$lib/stores/user';
    import { tick } from 'svelte';
    import ChatBubble from '$lib/components/ChatBubble.svelte';
    import type { GameInfo } from '$lib/types/game_infos';

    let { gameId, game, onNewMessage }: { 
        gameId: string; 
        game: GameInfo;
        onNewMessage?: () => void;
    } = $props();

    let messages = $state<any[]>([]);
    let content = $state('');
    let sending = $state(false);
    let loading = $state(true);
    let error = $state('');
    let usernameMap = $derived.by(() => {
        const map: Record<string, string> = {};
        if (game?.players) {
            game.players.forEach((p: any) => { map[String(p.id)] = p.username; });
        }
        return map;
    });

    let scrollContainer = $state<HTMLDivElement | null>(null);
    let pollInterval: any;

    async function loadMessages(silent = false) {
        if (!silent) loading = true;
        try {
            const newMessages = await getMessages(gameId);
            
            // Check if there are new messages
            if (newMessages.length > messages.length) {
                if (messages.length > 0 && typeof onNewMessage === 'function') {
                    onNewMessage();
                }
                messages = newMessages;
                await tick();
                scrollToBottom();
            } else {
                messages = newMessages;
            }
        } catch (e: any) {
            console.error('failed to load messages', e);
            error = 'Impossible de charger les messages';
        } finally {
            if (!silent) loading = false;
        }
    }

    async function markMessagesRead() {
        try {
            const lastId = messages.length ? messages[messages.length - 1].id : 0;
            await markMessagesAsRead(gameId, lastId);
        } catch (e) {
            console.warn('failed to mark messages read', e);
        }
    }

    async function send() {
        const trimmed = content.trim();
        if (!trimmed) return;
        sending = true;
        try {
            await sendMessage(gameId, trimmed);
            content = '';
            await loadMessages(true);
            await markMessagesRead();
            await tick();
            scrollToBottom();
        } catch (e: any) {
            error = e?.response?.data?.message || 'Erreur lors de l\'envoi';
        } finally {
            sending = false;
        }
    }

    async function removeMessage(msgId: number) {
        try {
            await deleteChatMessage(gameId, msgId);
            await loadMessages(true);
            await markMessagesRead();
            await tick();
            scrollToBottom();
        } catch (e) {
            console.error('delete failed', e);
        }
    }

    function scrollToBottom() {
        if (scrollContainer) {
            scrollContainer.scrollTop = scrollContainer.scrollHeight;
        }
    }

    onMount(async () => {
        await loadMessages();
        await markMessagesRead();
        await tick();
        scrollToBottom();

        // Poll messages every 6 seconds to feel live
        pollInterval = setInterval(async () => {
            await loadMessages(true);
        }, 6000);
    });

    onDestroy(() => {
        if (pollInterval) clearInterval(pollInterval);
    });

    function extractDayFromDate(date: Date): string {
        return date.toLocaleDateString('fr-FR', { weekday: 'long', year: 'numeric', month: 'long', day: 'numeric' });
    }

    // Group messages by day
    let groupedMessages = $derived.by(() => {
        const out: { msg: any; showDate: boolean; msgDate: Date }[] = [];
        let prev: Date | null = null;
        messages.forEach((m) => {
            const msgDate = new Date(m.created_at);
            const showDate = !prev || extractDayFromDate(prev) !== extractDayFromDate(msgDate);
            out.push({ msg: m, showDate, msgDate });
            prev = msgDate;
        });
        return out;
    });
</script>

<div class="flex-1 flex flex-col min-h-0 bg-stone-50/30">
    {#if loading}
        <div class="flex-1 flex items-center justify-center">
            <div class="flex flex-col items-center gap-3">
                <svg class="animate-spin text-brand-emerald w-8 h-8" fill="none" stroke="currentColor" stroke-width="3" viewBox="0 0 24 24">
                    <circle cx="12" cy="12" r="10" stroke="rgba(12,106,77,0.15)" stroke-width="3"></circle>
                    <path d="M22 12a10 10 0 0 1-10 10" stroke="currentColor" stroke-width="3" stroke-linecap="round"></path>
                </svg>
                <p class="text-xs font-semibold text-stone-500">Chargement des messages...</p>
            </div>
        </div>
    {:else if error && messages.length === 0}
        <div class="flex-1 flex items-center justify-center p-6 text-center">
            <p class="text-sm font-semibold text-red-600">⚠️ {error}</p>
        </div>
    {:else}
        <!-- Message list -->
        <div 
            bind:this={scrollContainer}
            class="flex-1 overflow-y-auto px-4 py-4 space-y-2 flex flex-col no-scrollbar"
        >
            {#each groupedMessages as item}
                {#if item.showDate}
                    <div class="w-full flex justify-center my-3">
                        <div class="inline-block px-3 py-1 bg-stone-200/50 text-[10px] font-bold text-stone-600 rounded-full border border-stone-200/20">
                            {extractDayFromDate(item.msgDate)}
                        </div>
                    </div>
                {/if}
                
                {@const msg_username = usernameMap[String(item.msg.user_id)] ?? item.msg.username ?? '–'}
                {@const isOwn = msg_username === $user?.username}
                
                <ChatBubble
                    msg={item.msg}
                    username={msg_username}
                    userId={item.msg.user_id}
                    {isOwn}
                    onDelete={item.msg.user_id === $user?.id ? removeMessage : null}
                />
            {/each}

            {#if messages.length === 0}
                <div class="flex-1 flex flex-col items-center justify-center text-center p-8 select-none">
                    <span class="text-3xl mb-2">💬</span>
                    <p class="text-xs font-bold text-stone-500">Aucun message pour l'instant</p>
                    <p class="text-[10px] text-stone-400 mt-0.5">Envoyez le premier mot de la partie !</p>
                </div>
            {/if}
        </div>

        <!-- Chat Input -->
        <div class="shrink-0 bg-white/80 backdrop-blur-md border-t border-stone-200/50 p-3 pb-[calc(env(safe-area-inset-bottom)+12px)]">
            <div class="flex items-center gap-2 max-w-lg mx-auto">
                <textarea
                    bind:value={content}
                    rows={1}
                    placeholder="Écrire à vos proches..."
                    class="flex-1 p-3 border border-stone-200 rounded-2xl resize-none text-xs placeholder-stone-400 bg-white focus:outline-none focus:ring-2 focus:ring-brand-emerald/40 focus:border-brand-emerald transition max-h-20 min-h-[42px]"
                    onkeydown={(e) => {
                        if (e.key === 'Enter' && !e.shiftKey) { 
                            e.preventDefault(); 
                            send(); 
                        }
                    }}
                ></textarea>
                <button 
                    class="shrink-0 inline-flex items-center justify-center h-10 px-5 rounded-2xl bg-brand-emerald hover:bg-brand-emerald-hover text-white text-xs font-bold shadow-md shadow-brand-emerald/10 transition active:scale-95 disabled:opacity-50 cursor-pointer"
                    onclick={send} 
                    disabled={sending || !content.trim()}
                >
                    Envoyer
                </button>
            </div>
        </div>
    {/if}
</div>
