<script lang="ts">
    import { goto } from "$app/navigation";

    let {
        title = "",
        back = false,
        right = null as (() => any) | null,
        topOffset = "0px",
    }: {
        title?: string;
        back?: boolean;
        right?: (() => any) | null;
        topOffset?: string;
    } = $props();

    function goBack() {
        if (history.length > 1) history.back();
        else goto("/");
    }
</script>

<header
    class="sticky z-20 bg-white/90 backdrop-blur supports-backdrop-blur:border-b border-b"
    style={`top: ${topOffset}`}
>
    <div class="max-w-[720px] mx-auto px-3 py-2 flex items-center gap-2">
        {#if back}
            <button
                class="p-2 rounded hover:bg-gray-100"
                aria-label="Retour"
                onclick={goBack}
            >
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none">
                    <path d="M15 18l-6-6 6-6" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                </svg>
            </button>
        {/if}
        <h1 class="text-base font-semibold flex-1 truncate">{title}</h1>
        <div class="flex items-center gap-2">
            {#if right}{@render right()}{/if}
        </div>
    </div>

    <div class="h-[4px] bg-gradient-to-r from-green-600/30 via-green-600/60 to-green-600/30"></div>
</header>
