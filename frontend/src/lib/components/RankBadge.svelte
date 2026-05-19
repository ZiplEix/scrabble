<script lang="ts">
    import { getRankInfo } from '$lib/ranks';

    let {
        rating,
        size = 'md'
    }: {
        rating: number;
        size?: 'sm' | 'md' | 'lg' | 'xl';
    } = $props();

    const rank = $derived(getRankInfo(rating));
    
    const sizeClasses = {
        sm: 'w-8 h-8',
        md: 'w-10 h-10',
        lg: 'w-12 h-12',
        xl: 'w-14 h-14'
    };

    const badgeStyles = {
        fer: {
            bg: 'bg-stone-700',
            accent: 'text-stone-300',
            ring: 'ring-stone-600'
        },
        bronze: {
            bg: 'bg-amber-700',
            accent: 'text-amber-200',
            ring: 'ring-amber-500'
        },
        argent: {
            bg: 'bg-slate-500',
            accent: 'text-slate-200',
            ring: 'ring-slate-400'
        },
        or: {
            bg: 'bg-amber-500',
            accent: 'text-amber-100',
            ring: 'ring-amber-400'
        },
        platine: {
            bg: 'bg-cyan-400',
            accent: 'text-cyan-900',
            ring: 'ring-cyan-300'
        }
    };

    const style = badgeStyles[rank.key];
</script>

<div class="inline-flex items-center justify-center {sizeClasses[size]} rounded-full {style.bg} {style.ring} ring-2 shadow-md relative overflow-hidden" title={rank.label}>
    <!-- Background shimmer effect for platine -->
    {#if rank.key === 'platine'}
        <div class="absolute inset-0 bg-gradient-to-br from-white/40 to-transparent opacity-60"></div>
    {/if}
    
    <!-- Inner design -->
    <svg viewBox="0 0 24 24" class="w-full h-full {style.accent} relative z-10" fill="currentColor">
        {#if rank.key === 'fer'}
            <!-- Iron: Simple square tile -->
            <rect x="5" y="5" width="14" height="14" fill="currentColor" opacity="0.6" />
            <rect x="6" y="6" width="12" height="12" fill="currentColor" opacity="0.3" stroke="currentColor" stroke-width="1" />
        {:else if rank.key === 'bronze'}
            <!-- Bronze: simple tile + centered star with softer contrast -->
            <rect x="4" y="4" width="16" height="16" rx="2" fill="currentColor" opacity="0.35" />
            <rect x="6" y="6" width="12" height="12" rx="1" fill="none" stroke="currentColor" stroke-width="1" opacity="0.55" />
            <path d="M12 9.2l2.8 2.8-2.8 2.8-2.8-2.8z" fill="currentColor" opacity="0.9" />
        {:else if rank.key === 'argent'}
            <!-- Silver: Square tile with double border -->
            <rect x="3" y="3" width="18" height="18" fill="none" stroke="currentColor" stroke-width="1.5" opacity="0.6" />
            <rect x="5" y="5" width="14" height="14" fill="currentColor" opacity="0.25" stroke="currentColor" stroke-width="1" />
            <circle cx="12" cy="12" r="3" fill="currentColor" opacity="0.8" />
        {:else if rank.key === 'or'}
            <!-- Gold: Filled square with shine -->
            <rect x="4" y="4" width="16" height="16" rx="2" fill="currentColor" />
            <rect x="6" y="6" width="12" height="12" rx="1" fill="currentColor" opacity="0.5" />
            <path d="M8 8l8 8M16 8l-8 8" stroke="currentColor" stroke-width="1" opacity="0.4" />
            <circle cx="12" cy="12" r="4.8" fill="white" opacity="0.2" />
            <path d="M12 7.2l1.45 4.2h4.3l-3.45 2.55 1.35 4.05L12 15.6l-3.65 2.6 1.35-4.05-3.45-2.55h4.3z" fill="white" opacity="1" />
        {:else if rank.key === 'platine'}
            <!-- Platine: Premium diamond/gem shape -->
            <path d="M12 3l6 9l-6 9l-6-9z" fill="currentColor" />
            <path d="M12 5l4 7l-4 7l-4-7z" fill="currentColor" opacity="0.5" stroke="currentColor" stroke-width="0.5" />
            <path d="M12 5l3 4M12 5l-3 4" stroke="currentColor" stroke-width="0.8" opacity="0.6" />
        {/if}
    </svg>
</div>
