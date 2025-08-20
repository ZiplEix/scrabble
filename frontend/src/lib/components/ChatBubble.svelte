<script lang="ts">
    import { extractTextFromMessage, formatMessage } from '$lib/utils/typography';

    export let msg: any;
    export let username: string = '';
    export let isOwn: boolean = false;
    export let onDelete: ((id: number) => void) | null = null;

    function initials(name: string) {
        if (!name) return '?';
        return name
            .split(' ')
            .map((n) => n[0] || '')
            .slice(0, 2)
            .join('')
            .toUpperCase();
    }

    function safeTime(dateLike: any) {
        try {
            const d = new Date(dateLike);
            if (isNaN(d.getTime())) return '';
            return d.toLocaleTimeString(undefined, { hour: '2-digit', minute: '2-digit' });
        } catch (e) {
            return '';
        }
    }

    $: formatted = formatMessage(msg?.content ?? '');

    // long-press / context menu handling
    let pressTimer: any = null;
    let menuOpen = false;
    let menuX = 0;
    let menuY = 0;
    let showRaw = false;

    function startPress(e: MouseEvent | TouchEvent) {
        // do not start if already open
        if (pressTimer) return;
        pressTimer = setTimeout(() => {
            pressTimer = null;
            openMenuAtEvent(e);
        }, 500);
    }

    function cancelPress() {
        if (pressTimer) {
            clearTimeout(pressTimer);
            pressTimer = null;
        }
    }

    function openMenuAtEvent(e: any) {
        e.preventDefault?.();
        const ev = e.touches?.[0] ?? e;
        menuX = ev.clientX || 0;
        menuY = ev.clientY || 0;
        menuOpen = true;
    }

    function openMenuFromEvent(e: MouseEvent) {
        openMenuAtEvent(e);
    }

    function openMenuFromKeyboard(e: KeyboardEvent) {
        // estimate position from element bounding rect
        const el = (e.currentTarget as HTMLElement);
        const rect = el.getBoundingClientRect();
        menuX = rect.left + rect.width / 2;
        menuY = rect.top + rect.height / 2;
        menuOpen = true;
    }

    function closeMenu() {
        menuOpen = false;
    }

    function onViewRaw() {
        closeMenu();
        showRaw = true;
    }

    function onCloseRaw() {
        showRaw = false;
    }

    function onDeleteClick() {
        closeMenu();
        if (onDelete && msg?.id) onDelete(msg.id);
    }
</script>

<div class={`w-full flex ${isOwn ? 'justify-end' : 'justify-start'} mb-2`}>
    {#if !isOwn}
        <div class="flex items-start gap-2">
            <div class="w-8 h-8 rounded-full flex-none bg-gray-200 text-gray-700 flex items-center justify-center text-sm font-semibold">
                {initials(username)}
            </div>
            <div
                role="button"
                tabindex="0"
                class="rounded-lg px-3 py-2 bg-white shadow-sm border border-gray-100 text-gray-900 min-w-0 flex-grow max-w-[70%] break-words"
                style="overflow-wrap:anywhere;"
                on:touchstart|passive={startPress}
                on:touchend={cancelPress}
                on:mousedown={startPress}
                on:mouseup={cancelPress}
                on:mouseleave={cancelPress}
                on:contextmenu|preventDefault={openMenuFromEvent}
                on:keydown={(e) => { if (e.key === 'Enter') openMenuFromKeyboard(e); if (e.key === 'Escape') closeMenu(); }}
            >
                <div class="text-sm">{@html formatted}</div>
                {#if msg.created_at}
                    <div class="text-xs text-gray-500 mt-1">{safeTime(msg.created_at)}</div>
                {/if}
            </div>
        </div>
    {:else}
        <div class="flex items-end gap-2 w-full justify-end">
            <div
                role="button"
                tabindex="0"
                class="rounded-lg px-3 py-2 bg-green-600 text-white min-w-0 flex-grow max-w-[70%] break-words"
                style="overflow-wrap:anywhere;"
                on:touchstart|passive={startPress}
                on:touchend={cancelPress}
                on:mousedown={startPress}
                on:mouseup={cancelPress}
                on:mouseleave={cancelPress}
                on:contextmenu|preventDefault={openMenuFromEvent}
                on:keydown={(e) => { if (e.key === 'Enter') openMenuFromKeyboard(e); if (e.key === 'Escape') closeMenu(); }}
            >
                <div class="text-sm">{@html formatted}</div>
                <div class="flex items-center gap-2 mt-1">
                    {#if msg.created_at}
                        <div class="text-xs text-green-200">{safeTime(msg.created_at)}</div>
                    {/if}
                </div>
            </div>
            <div class="w-8 h-8 rounded-full flex-none bg-green-600 flex items-center justify-center text-sm font-semibold text-white">
                {initials(username)}
            </div>
        </div>
    {/if}
</div>

{#if menuOpen}
    <!-- Fixed top banner under the navbar; uses --nav-h defaulting to 72px -->
    <div class="fixed left-0 right-0 z-50" style="top: var(--nav-h, 72px);">
        <div class="mx-auto max-w-[640px] px-3 py-2">
            <div class="w-full rounded-2xl bg-white/95 backdrop-blur-md shadow-lg ring-1 ring-black/5 flex items-center justify-between px-4 py-2">
                <div class="flex items-center gap-2">
                    {#if isOwn}
                        <button
                            class="px-3 py-1 bg-red-50 text-red-600 rounded hover:bg-red-100"
                            on:click={onDeleteClick}
                        >
                            Supprimer
                        </button>
                    {/if}
                    <button
                        class="px-3 py-1 bg-gray-100 text-gray-900 rounded hover:bg-gray-100"
                        on:click={onViewRaw}
                    >
                        Text brut
                    </button>
                    <button
                        class="px-3 py-1 bg-gray-100 text-gray-900 rounded hover:bg-gray-100"
                        on:click={() => {
                            navigator.clipboard.writeText(extractTextFromMessage(msg.content));
                            closeMenu();
                            alert('Message copié dans le presse-papiers');
                            closeMenu();
                        }}
                    >
                        Copier
                    </button>
                </div>
                <div>
                    <button class="text-sm text-gray-500" on:click={closeMenu} aria-label="Fermer le menu">Fermer</button>
                </div>
            </div>
        </div>
    </div>
{/if}

{#if showRaw}
    <div class="fixed inset-0 z-50 bg-black/40 flex items-center justify-center p-8" role="dialog" aria-modal="true" on:keydown={(e) => e.key === 'Escape' && onCloseRaw()} tabindex="-1">
        <button class="absolute inset-0" on:click={onCloseRaw} aria-label="Fermer la vue brute"></button>
        <div class="bg-white max-w-lg w-full p-4 rounded-md shadow" role="document">
            <div class="flex justify-between items-center mb-2">
                <strong>Message brut</strong>
                <button class="text-sm text-gray-900" on:click={onCloseRaw}>Fermer</button>
            </div>
            <pre class="whitespace-pre-wrap text-sm bg-gray-50 p-3 rounded">{msg.content}</pre>
        </div>
    </div>
{/if}
