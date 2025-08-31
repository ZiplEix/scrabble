<script lang="ts">
    const {
        padded = true,
        class: className = "",
        children,
    } = $props<{
        padded?: boolean;
        class?: string;
        children?: () => any;
    }>();
</script>

<div class={`rounded-2xl bg-white ring-1 ring-black/5 shadow ${className}`}>
    <div class={padded ? "p-4" : ""}>
        {@render children?.()}
    </div>
</div>
