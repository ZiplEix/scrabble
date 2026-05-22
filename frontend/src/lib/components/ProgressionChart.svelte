<script lang="ts">
	import type { RatingHistoryResponse } from '$lib/types/user_infos';

	// Svelte 5 Runes Props
	let { history = [] }: { history: RatingHistoryResponse[] } = $props();

	// Graph parameters
	const width = 600;
	const height = 240;
	const paddingX = 50;
	const paddingY = 30;

	// State for the hovered point
	let hoveredPoint = $state<{ x: number; y: number; data: RatingHistoryResponse; index: number } | null>(null);

	// Derived values for coordinates calculation
	let ratings = $derived(history.map((h) => h.rating));
	let minRating = $derived(ratings.length > 0 ? Math.min(...ratings) : 1600);
	let maxRating = $derived(ratings.length > 0 ? Math.max(...ratings) : 1600);

	let ratingRange = $derived(maxRating - minRating);
	let margin = $derived(ratingRange > 0 ? Math.max(15, Math.round(ratingRange * 0.25)) : 50);

	let yMin = $derived(minRating - margin);
	let yMax = $derived(maxRating + margin);

	// Calculate (x, y) coordinates for each point in history
	let points = $derived(
		history.map((h, i) => {
			const x =
				history.length > 1
					? paddingX + (i / (history.length - 1)) * (width - 2 * paddingX)
					: width / 2;
			const y = height - paddingY - ((h.rating - yMin) / (yMax - yMin)) * (height - 2 * paddingY);
			return { x, y, data: h, index: i };
		})
	);

	// Draw lines
	let linePath = $derived(
		points.length > 0 ? 'M ' + points.map((p) => `${p.x} ${p.y}`).join(' L ') : ''
	);

	let areaPath = $derived(
		points.length > 0
			? linePath +
					` L ${points[points.length - 1].x} ${height - paddingY} L ${points[0].x} ${height - paddingY} Z`
			: ''
	);

	// Helper to format dates
	function formatDate(dateStr: string): string {
		try {
			const d = new Date(dateStr);
			return d.toLocaleDateString('fr-FR', { day: 'numeric', month: 'short', year: 'numeric' });
		} catch {
			return dateStr;
		}
	}
</script>

<div class="relative w-full overflow-visible select-none rounded-2xl bg-white border border-stone-200/60 p-5 shadow-sm">
	{#if history.length === 0}
		<div class="flex flex-col items-center justify-center py-16 text-stone-400">
			<span class="text-3xl">📈</span>
			<p class="mt-2.5 text-sm font-semibold text-stone-600">Pas encore d'historique d'IPS</p>
			<p class="text-xs text-stone-400 mt-1 max-w-[280px] text-center">
				Terminez votre première partie pour voir votre courbe de progression s'animer.
			</p>
		</div>
	{:else}
		<!-- SVG Chart -->
		<svg
			viewBox="0 0 {width} {height}"
			class="w-full h-auto overflow-visible"
		>
			<defs>
				<!-- Gradient fill under line -->
				<linearGradient id="area-gradient" x1="0" y1="0" x2="0" y2="1">
					<stop offset="0%" stop-color="#10b981" stop-opacity="0.16" />
					<stop offset="100%" stop-color="#10b981" stop-opacity="0.0" />
				</linearGradient>
				<!-- Glowing shadow under line -->
				<filter id="line-glow" x="-20%" y="-20%" width="140%" height="140%">
					<feDropShadow dx="0" dy="3" stdDeviation="4" flood-color="#10b981" flood-opacity="0.25" />
				</filter>
			</defs>

			<!-- Grid Lines (Horizontal) -->
			<!-- Max Line -->
			<line
				x1={paddingX}
				y1={height - paddingY - ((maxRating - yMin) / (yMax - yMin)) * (height - 2 * paddingY)}
				x2={width - paddingX}
				y2={height - paddingY - ((maxRating - yMin) / (yMax - yMin)) * (height - 2 * paddingY)}
				class="stroke-stone-200/50 stroke-[1] stroke-dasharray-[4_4]"
				stroke-dasharray="4 4"
			/>
			<text
				x={paddingX - 10}
				y={height - paddingY - ((maxRating - yMin) / (yMax - yMin)) * (height - 2 * paddingY) + 4}
				class="text-[9px] font-black text-stone-400 text-right"
				text-anchor="end"
			>
				{maxRating}
			</text>

			<!-- Mid Line (only if max and min differ significantly) -->
			{#if maxRating - minRating > 10}
				{@const midVal = Math.round((maxRating + minRating) / 2)}
				<line
					x1={paddingX}
					y1={height - paddingY - ((midVal - yMin) / (yMax - yMin)) * (height - 2 * paddingY)}
					x2={width - paddingX}
					y2={height - paddingY - ((midVal - yMin) / (yMax - yMin)) * (height - 2 * paddingY)}
					class="stroke-stone-100/50 stroke-[1]"
				/>
				<text
					x={paddingX - 10}
					y={height - paddingY - ((midVal - yMin) / (yMax - yMin)) * (height - 2 * paddingY) + 4}
					class="text-[9px] font-bold text-stone-400 text-right"
					text-anchor="end"
				>
					{midVal}
				</text>
			{/if}

			<!-- Min Line -->
			<line
				x1={paddingX}
				y1={height - paddingY - ((minRating - yMin) / (yMax - yMin)) * (height - 2 * paddingY)}
				x2={width - paddingX}
				y2={height - paddingY - ((minRating - yMin) / (yMax - yMin)) * (height - 2 * paddingY)}
				class="stroke-stone-200/50 stroke-[1] stroke-dasharray-[4_4]"
				stroke-dasharray="4 4"
			/>
			<text
				x={paddingX - 10}
				y={height - paddingY - ((minRating - yMin) / (yMax - yMin)) * (height - 2 * paddingY) + 4}
				class="text-[9px] font-black text-stone-400 text-right"
				text-anchor="end"
			>
				{minRating}
			</text>

			<!-- Filled Area Under the Line -->
			<path
				d={areaPath}
				fill="url(#area-gradient)"
			/>

			<!-- The Progression Line -->
			<path
				d={linePath}
				fill="none"
				stroke="#10b981"
				stroke-width="3.5"
				stroke-linecap="round"
				stroke-linejoin="round"
				filter="url(#line-glow)"
			/>

			<!-- Interactive Points Circles -->
			{#each points as p (p.index)}
				<!-- Transparent larger hover area for easier mouse targeting -->
				<circle
					cx={p.x}
					cy={p.y}
					r="14"
					fill="transparent"
					class="cursor-pointer"
					role="img"
					aria-label="Détail de la partie"
					onmouseenter={() => (hoveredPoint = p)}
					onmouseleave={() => (hoveredPoint = null)}
				/>
				<!-- Actual visible dot -->
				<circle
					cx={p.x}
					cy={p.y}
					r={hoveredPoint?.index === p.index ? '6.5' : '4.5'}
					class="transition-all duration-150 cursor-pointer pointer-events-none"
					fill={hoveredPoint?.index === p.index ? '#10b981' : '#ffffff'}
					stroke="#10b981"
					stroke-width={hoveredPoint?.index === p.index ? '2.5' : '3.5'}
				/>
			{/each}
		</svg>

		<!-- Premium Dynamic Floating Tooltip -->
		{#if hoveredPoint}
			{@const gInfo = hoveredPoint.data.game_info}
			<div
				class="absolute flex flex-col gap-1.5 min-w-[180px] bg-stone-900 border border-stone-800 text-stone-100 p-3 rounded-2xl shadow-xl transition-all duration-150 text-left pointer-events-none"
				style="
					left: {((hoveredPoint.x - paddingX) / (width - 2 * paddingX)) * 88 + 6}%;
					bottom: {(1 - (hoveredPoint.y / height)) * 80 + 20}%;
					transform: translateX(-50%);
					z-index: 50;
				"
			>
				<!-- Tooltip Header -->
				<div class="flex items-center justify-between gap-3 text-[10px] text-stone-400 font-bold tracking-wider">
					<span>{formatDate(hoveredPoint.data.created_at)}</span>
					<span class="text-emerald-400">{hoveredPoint.data.rating} IPS</span>
				</div>

				<!-- Game Details -->
				{#if gInfo}
					<div class="flex flex-col gap-0.5 mt-0.5 border-t border-stone-800 pt-1.5">
						<div class="flex items-center gap-1.5">
							{#if gInfo.won}
								<span class="px-1.5 py-0.5 text-[8px] font-black uppercase rounded-md bg-emerald-500/20 text-emerald-400 border border-emerald-500/10">
									Victoire
								</span>
							{:else}
								<span class="px-1.5 py-0.5 text-[8px] font-black uppercase rounded-md bg-rose-500/20 text-rose-400 border border-rose-500/10">
									Défaite
								</span>
							{/if}
							<span class="text-[11px] font-medium truncate text-stone-300">
								vs {gInfo.opponent_username}
							</span>
						</div>
						<div class="text-[10px] text-stone-400 font-medium mt-0.5">
							Score : <strong class="text-stone-200">{gInfo.user_score}</strong> - {gInfo.opponent_score}
						</div>
					</div>
				{:else}
					<div class="text-[10px] text-stone-400 mt-1 border-t border-stone-800 pt-1.5 italic">
						Mise à jour du classement
					</div>
				{/if}
			</div>
		{/if}
	{/if}
</div>

<style>
	/* Subtle transition styles if needed */
	circle {
		transition: r 0.15s ease, fill 0.15s ease, stroke-width 0.15s ease;
	}
</style>
