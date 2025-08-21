<script lang='ts'>
    let { checked = $bindable(), onToggle }: {
        checked: boolean,
        onToggle: () => void
    } = $props();
</script>

<label class="inline-flex items-center cursor-pointer">
    <input
        type="checkbox"
        class="sr-only"
        bind:checked={checked}
        aria-label="checkbox input"
        onchange={onToggle}
    />
    <div class="w-12 h-7 rounded-full relative transition-colors duration-150"
            aria-hidden="true"
            class:bg-green-500={checked} class:bg-gray-300={!checked}>
        <span
            class="absolute top-1 left-1 bg-white w-5 h-5 rounded-full shadow transition-transform duration-150"
            style="transform-origin: center"
            class:translate-x-5={checked}
        ></span>
    </div>
</label>
