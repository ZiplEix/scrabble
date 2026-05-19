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
        <h1 class="text-xl font-bold text-gray-900">Comment fonctionne le système Elo</h1>
        <p class="text-sm text-gray-700 mt-2">
            Le Elo mesure ton niveau relatif. Après chaque partie, ton score monte ou descend selon le résultat
            et le niveau des adversaires.
        </p>
    </section>

    <section class="rounded-2xl bg-white ring-1 ring-black/5 p-4 shadow-sm">
        <h2 class="text-lg font-semibold text-gray-900">Règles principales</h2>
        <ul class="mt-3 space-y-2 text-sm text-gray-700 list-disc pl-5">
            <li>Si tu bats un joueur mieux classé, tu gagnes plus de points.</li>
            <li>Si tu perds contre un joueur moins bien classé, tu perds plus de points.</li>
            <li>Contre un joueur de niveau proche, les variations sont plus modérées.</li>
            <li>Dans les parties à plusieurs joueurs, le calcul est fait par comparaisons entre chaque paire de joueurs.</li>
        </ul>
    </section>

    <section class="rounded-2xl bg-white ring-1 ring-black/5 p-4 shadow-sm">
        <h2 class="text-lg font-semibold text-gray-900">Rangs</h2>
        <p class="text-sm text-gray-600 mt-1">Ton rang dépend directement de ton Elo actuel.</p>

        <div class="mt-4 space-y-2">
            {#each ranks as rank, i}
                <div class="flex items-center justify-between gap-3 rounded-xl border border-gray-100 p-3">
                    <div class="flex items-center gap-3 min-w-0">
                        <RankBadge rating={rank.minRating} size="md" />
                        <div class="min-w-0">
                            <div class="font-semibold text-gray-900">{rank.label}</div>
                            <div class="text-xs text-gray-500">Elo {rankRangeLabel(rank, i)}</div>
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
        <h2 class="text-lg font-semibold text-gray-900">Formule (version simplifiée)</h2>
        <div class="mt-3 text-sm text-gray-700 space-y-2">
            <p>Espérance de victoire: E = 1 / (1 + 10^((R_adv - R_joueur) / 400))</p>
            <p>Variation: Δ = K × (S - E), avec K = 32 dans l'app</p>
            <p class="text-xs text-gray-500">S vaut 1 pour une victoire et 0 pour une défaite.</p>
        </div>
        <div class="mt-3 rounded-lg bg-gray-50 p-3 text-xs text-gray-600">
            Exemple: un joueur 1600 contre 1800 a ~{(expectedVsStronger * 100).toFixed(1)}% de chance théorique.
            Le joueur 1800 contre 1600 a ~{(expectedVsWeaker * 100).toFixed(1)}%.
        </div>
    </section>
</main>