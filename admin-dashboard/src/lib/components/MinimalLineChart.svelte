<script lang="ts">
    export let data: number[] = [];
    export let labels: string[] | null = null; // x labels (hours strings)
    export let xTickStep: number = 6; // show label every xTickStep
    export let color: string = '#7c3aed'; // violet-600
    export let strokeWidth: number = 2;
    // visual height in pixels (default 2x of previous VH)
    export let height: number = 80;

    // SVG logical width
    const VW = 100;

    function buildPath(values: number[]) {
            if (!values || values.length === 0) return '';
            const min = Math.min(...values);
            const max = Math.max(...values);
            const range = max - min || 1;
            const stepX = VW / Math.max(1, values.length - 1);

            // build points array in viewBox coordinates
            const pts = values.map((v, i) => ({
                x: +(i * stepX).toFixed(2),
                y: +((height - ((v - min) / range) * height)).toFixed(2)
            }));

            if (pts.length === 1) return `M ${pts[0].x} ${pts[0].y}`;
            if (pts.length === 2) return `M ${pts[0].x} ${pts[0].y} L ${pts[1].x} ${pts[1].y}`;

            // Catmull-Rom to bezier conversion for smooth curve
            let d = `M ${pts[0].x} ${pts[0].y}`;
            for (let i = 0; i < pts.length - 1; i++) {
                const p0 = i - 1 >= 0 ? pts[i - 1] : pts[i];
                const p1 = pts[i];
                const p2 = pts[i + 1];
                const p3 = i + 2 < pts.length ? pts[i + 2] : pts[i + 1];

                // control points
                const cp1x = p1.x + (p2.x - p0.x) / 6;
                const cp1y = p1.y + (p2.y - p0.y) / 6;
                const cp2x = p2.x - (p3.x - p1.x) / 6;
                const cp2y = p2.y - (p3.y - p1.y) / 6;

                d += ` C ${cp1x.toFixed(2)} ${cp1y.toFixed(2)}, ${cp2x.toFixed(2)} ${cp2y.toFixed(2)}, ${p2.x} ${p2.y}`;
            }
            return d;
    }

    function computeYLabels(values: number[]) {
        if (!values || values.length === 0) return [0, 0, 0];
        const min = Math.min(...values);
        const max = Math.max(...values);
        const mid = Math.round((min + max) / 2);
        return [min, mid, max];
    }

    $: pathD = buildPath(data);
    $: yLabels = computeYLabels(data);
    // compute point positions in viewBox units for interactivity
    $: pointsPos = (data || []).map((v, i) => {
        const min = Math.min(...(data.length ? data : [0]));
        const max = Math.max(...(data.length ? data : [0]));
        const range = max - min || 1;
        const stepX = VW / Math.max(1, data.length - 1);
        const x = +(i * stepX).toFixed(2);
        const y = +((height - ((v - min) / range) * height)).toFixed(2);
        return { x, y };
    });

    let svgEl: SVGSVGElement | null = null;
    let tooltipVisible = false;
    let tooltipX = 0;
    let tooltipY = 0;
    let tooltipValue: number | null = null;
    let hoveredIndex = -1;

    function updateTooltip(idx: number, ev?: MouseEvent) {
        if (!svgEl) return;
        const rect = svgEl.getBoundingClientRect();
        const p = pointsPos[idx];
        const px = rect.left + (p.x / VW) * rect.width;
        const py = rect.top + (p.y / height) * rect.height;
        tooltipX = px;
        tooltipY = py;
        tooltipValue = data[idx];
        tooltipVisible = true;
        hoveredIndex = idx;
    }

    function hideTooltip() {
        tooltipVisible = false;
        hoveredIndex = -1;
    }

    function keyTrigger(idx: number, e: KeyboardEvent) {
        if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault();
            updateTooltip(idx);
        }
    }
</script>

<div class="w-full">
    <div class="flex items-center gap-3">
        <!-- Y labels column (min, mid, max) -->
        <div class="flex flex-col justify-between text-xs text-white/60" style="height: {height}px">
            <div>{yLabels[2]}</div>
            <div>{yLabels[1]}</div>
            <div>{yLabels[0]}</div>
        </div>

        <!-- Chart -->
        <div class="flex-1">
            <div class="relative w-full">
                <svg bind:this={svgEl} viewBox={"0 0 100 " + height} preserveAspectRatio="none" class="w-full" style="height: {height}px">
                    <path d={pathD}
                        fill="none"
                        stroke={color}
                        stroke-width={strokeWidth}
                        stroke-linecap="round"
                        stroke-linejoin="round"
                        vector-effect="non-scaling-stroke"
                        class="opacity-95"
                    />

                    {#if data && data.length}
                        {#each pointsPos as p, idx}
                            <!-- invisible larger hit area -->
                            <circle
                                cx={p.x}
                                cy={p.y}
                                r={Math.max(6, (VW / Math.max(1, data.length - 1)) * 0.8)}
                                fill="transparent"
                                pointer-events="all"
                                role="button"
                                tabindex="0"
                                on:mouseenter={(e) => updateTooltip(idx, e)}
                                on:mousemove={(e) => updateTooltip(idx, e)}
                                on:mouseleave={hideTooltip}
                                on:keydown={(e) => keyTrigger(idx, e)}
                            />

                            <!-- small visible dot when hovered -->
                            {#if hoveredIndex === idx}
                                <circle cx={p.x} cy={p.y} r={3} fill={color} class="opacity-95" />
                            {/if}
                        {/each}
                    {/if}
                </svg>

                {#if tooltipVisible && tooltipValue !== null}
                    <div style="position:fixed; left: {tooltipX}px; top: {tooltipY - 36}px; transform: translateX(-50%); pointer-events: none;" class="z-50">
                        <div class="bg-white text-black text-xs px-2 py-1 rounded shadow">{tooltipValue}</div>
                    </div>
                {/if}
            </div>

            {#if labels}
                <div class="w-full mt-1 text-xs text-white/50 flex justify-between">
                    {#each labels as lab, idx}
                        {#if idx % xTickStep === 0}
                            <div style="width:0; transform:translateX(-50%);">{lab}</div>
                        {:else}
                            <div></div>
                        {/if}
                    {/each}
                </div>
            {/if}
        </div>
    </div>
</div>
