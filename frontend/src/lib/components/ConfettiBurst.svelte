<script lang="ts">
        import { onMount } from 'svelte';

        type Particle = {
                left: number;
                delay: number;
                duration: number;
                color: string;
                rotation: number;
                scale: number;
                drift: number;
        };

        const COLORS = ['#fbbf24', '#34d399', '#60a5fa', '#f472b6', '#f87171', '#c084fc', '#facc15'];

        let particles: Particle[] = [];

        onMount(() => {
                particles = Array.from({ length: 120 }, () => ({
                        left: Math.random() * 100,
                        delay: Math.random() * 0.4,
                        duration: 1.6 + Math.random() * 1.2,
                        color: COLORS[Math.floor(Math.random() * COLORS.length)],
                        rotation: Math.random() * 360,
                        scale: 0.6 + Math.random() * 0.8,
                        drift: Math.random(),
                }));
        });
</script>

<div class="confetti" aria-hidden="true">
        {#each particles as particle}
                <span
                        class="particle"
                        style={`left:${particle.left}%;animation-delay:${particle.delay}s;animation-duration:${particle.duration}s;background:${particle.color};--rotate:${particle.rotation}deg;--scale:${particle.scale};--drift:${particle.drift};`}
                ></span>
        {/each}
</div>

<style>
        .confetti {
                position: fixed;
                inset: 0;
                pointer-events: none;
                overflow: hidden;
                z-index: 60;
        }

        .particle {
                position: absolute;
                top: -12px;
                width: 8px;
                height: 14px;
                border-radius: 2px;
                opacity: 0;
                animation-name: fall;
                animation-timing-function: ease-out;
                animation-fill-mode: forwards;
        }

        @keyframes fall {
                0% {
                        opacity: 1;
                        transform: translate3d(0, -20px, 0) rotate(var(--rotate)) scale(var(--scale));
                }

                70% {
                        opacity: 1;
                }

                100% {
                        opacity: 0;
                        transform: translate3d(calc((var(--drift) - 0.5) * 320px), 110vh, 0)
                                rotate(calc(var(--rotate) + 360deg))
                                scale(var(--scale));
                }
        }
</style>
