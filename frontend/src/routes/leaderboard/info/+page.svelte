<script lang="ts">
    import HeaderBar from '$lib/components/HeaderBar.svelte';
    import RankBadge from '$lib/components/RankBadge.svelte';
    import { RANKS, type RankInfo } from '$lib/ranks';

    const ranks: RankInfo[] = RANKS;

    function rankRangeLabel(rank: RankInfo, index: number): string {
        const next = ranks[index + 1];
        if (!next) {
            return `${rank.minRating}+`;
        }
        return `${rank.minRating} - ${next.minRating - 1}`;
    }
    
    function expectedScore(playerRating: number, opponentRating: number): number {
        return 1 / (1 + Math.pow(10, (opponentRating - playerRating) / 400));
    }

    const expectedVsStronger = expectedScore(1600, 1800);
    const expectedVsWeaker = expectedScore(1800, 1600);
</script>

<HeaderBar title="Classement: infos" back={true} />

<main class="max-w-2xl mx-auto px-4 py-6 space-y-4">
    <section class="rounded-2xl bg-emerald-50 ring-1 ring-black/5 p-4">
        <h1 class="text-xl font-bold text-gray-900">Comment fonctionne l'IPS</h1>
        <p class="text-sm text-gray-700 mt-2">
            L'<b>Indice de Performance Scrabble (IPS)</b> mesure tes performances récentes. Il se base sur tes scores réels et récompense le volume de victoires.
        </p>
    </section>

    <section class="rounded-2xl bg-white ring-1 ring-black/5 p-4 shadow-sm">
        <h2 class="text-lg font-semibold text-gray-900">Calcul de l'IPS</h2>
        <p class="text-sm text-gray-600 mt-1">L'indice est recalculé à la fin de chaque partie sur la base de tes 10 derniers matchs.</p>
        <ul class="mt-3 space-y-2 text-sm text-gray-700 list-disc pl-5">
            <li><strong>Moyenne des scores</strong> : Moyenne de tes points sur les 10 dernières parties.</li>
            <li><strong>Bonus Victoire</strong> : Chaque victoire parmi ces 10 parties apporte <strong>+15 points</strong>.</li>
        </ul>
    </section>

    <section class="rounded-2xl bg-white ring-1 ring-black/5 p-4 shadow-sm">
        <h2 class="text-lg font-semibold text-gray-900">Rangs</h2>
        <p class="text-sm text-gray-600 mt-1">Ton rang dépend directement de ton IPS actuel.</p>

        <div class="mt-4 space-y-2">
            {#each ranks as rank, i}
                <div class="flex items-center justify-between gap-3 rounded-xl border border-gray-100 p-3">
                    <div class="flex items-center gap-3 min-w-0">
                        <RankBadge rating={rank.minRating} size="md" />
                        <div class="min-w-0">
                            <div class="font-semibold text-gray-900">{rank.label}</div>
                            <div class="text-xs text-gray-500">IPS {rankRangeLabel(rank, i)}</div>
                        </div>
                    </div>
                    <div class="text-xs px-2 py-1 rounded-full {rank.softClass}">
                        {rank.key}
                    </div>
                </div>
            {/each}
        </div>
    </section>

    <section class="rounded-2xl bg-white ring-1 ring-black/5 p-4 shadow-sm">
        <h2 class="text-lg font-semibold text-gray-900">Exemple</h2>
        <div class="mt-3 text-sm text-gray-700 space-y-2">
            <p>Score moyen: 320 pts, 4 victoires sur les 10 derniers matchs.</p>
            <p class="font-bold text-emerald-700">IPS = 320 + (4 × 15) = 380 (Argent)</p>
        </div>
    </section>
</main>