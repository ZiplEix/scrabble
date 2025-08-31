<script lang="ts">
    type ButtonType = 'button' | 'submit' | 'reset';
    type Variant = 'primary' | 'secondary' | 'ghost' | 'danger';
    type Size = 'sm' | 'md' | 'lg';

    const {
        variant = 'primary',
        size = 'md',
        full = false,
        disabled = false,
        ariaLabel = undefined as string | undefined,
        type = 'button' as ButtonType,
        onclick,
        children
    } = $props<{
        variant?: Variant;
        size?: Size;
        full?: boolean;
        disabled?: boolean;
        ariaLabel?: string;
        type?: ButtonType;
        onclick?: (e: MouseEvent) => void;
        children?: () => any;
    }>();

    export { disabled };

    const base = 'inline-flex items-center justify-center gap-2 rounded-full font-medium transition active:scale-95 disabled:opacity-60 disabled:pointer-events-none shadow-sm';
    const sizes: Record<string, string> = {
        sm: 'px-3 py-1.5 text-[13px]',
        md: 'px-4 py-2 text-[14px]',
        lg: 'px-5 py-3 text-[15px]'
    };
    const variants: Record<string, string> = {
        primary: 'bg-green-600 text-white hover:bg-green-700',
        secondary: 'bg-gray-100 text-gray-900 hover:bg-gray-200',
        ghost: 'bg-transparent text-gray-800 hover:bg-gray-100',
        danger: 'bg-red-600 text-white hover:bg-red-700'
    };
</script>

<button
    type={type}
    class={`${base} ${sizes[size] ?? sizes.md} ${variants[variant] ?? variants.primary} ${full ? 'w-full' : ''}`}
    aria-label={ariaLabel}
    disabled={disabled}
    onclick={onclick}
>
    {@render children?.()}
</button>
