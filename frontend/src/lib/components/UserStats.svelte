<script lang="ts">
    import type { UserInfos } from "$lib/types/user_infos";
    export let userInfos: UserInfos;
</script>

<div class="rounded-2xl bg-white/90 backdrop-blur-md ring-1 ring-black/5 shadow-lg p-4">
    <h2 class="text-lg font-medium mb-3">Statistiques</h2>
    <div class="grid grid-cols-2 sm:grid-cols-3 gap-3">
        <div class="p-3 bg-emerald-50 rounded-lg text-center">
            <div class="text-sm text-gray-600">Parties jouées</div>
            <div class="text-xl font-bold text-gray-900">{userInfos.games_count}</div>
            {#if userInfos.games_count_top_percent !== undefined && (userInfos.games_count_top_percent ?? 0) > 0}
                <div class="text-[11px] text-gray-500 mt-1">Top {Math.round(userInfos.games_count_top_percent)}%</div>
            {/if}
        </div>

        <div class="p-3 bg-emerald-50 rounded-lg text-center">
            <div class="text-sm text-gray-600">Victoires</div>
            <div class="text-xl font-bold text-gray-900">{userInfos.victories}</div>
            {#if userInfos.victories_top_percent !== undefined && (userInfos.victories_top_percent ?? 0) > 0}
                <div class="text-[11px] text-gray-500 mt-1">Top {Math.round(userInfos.victories_top_percent)}%</div>
            {/if}
        </div>

        <div class="p-3 bg-amber-50 rounded-lg text-center">
            <div class="text-sm text-gray-600">Score moyen</div>
            <div class="text-xl font-bold text-gray-900">{Math.round(userInfos.avg_score ?? 0)}</div>
            {#if userInfos.avg_score_top_percent !== undefined && (userInfos.avg_score_top_percent ?? 0) > 0}
                <div class="text-[11px] text-gray-500 mt-1">Top {Math.round(userInfos.avg_score_top_percent)}%</div>
            {/if}
        </div>

        <div class="p-3 bg-amber-50 rounded-lg text-center">
            <div class="text-sm text-gray-600">Meilleur score</div>
            <div class="text-xl font-bold text-gray-900">{userInfos.best_score}</div>
            {#if userInfos.best_score_top_percent !== undefined && (userInfos.best_score_top_percent ?? 0) > 0}
                <div class="text-[11px] text-gray-500 mt-1">Top {Math.round(userInfos.best_score_top_percent)}%</div>
            {/if}
        </div>

        <div class="p-3 bg-sky-50 rounded-lg text-center">
            <div class="text-sm text-gray-600">Moyenne / coup</div>
            <div class="text-xl font-bold text-gray-900">{Math.round(userInfos.avg_points_per_move ?? 0)}</div>
            {#if userInfos.avg_points_per_move_top_percent !== undefined && (userInfos.avg_points_per_move_top_percent ?? 0) > 0}
                <div class="text-[11px] text-gray-500 mt-1">Top {Math.round(userInfos.avg_points_per_move_top_percent)}%</div>
            {/if}
        </div>

        <div class="p-3 bg-sky-50 rounded-lg text-center">
            <div class="text-sm text-gray-600">Meilleur coup</div>
            <div class="text-xl font-bold text-gray-900">{userInfos.best_move_score ?? 0}</div>
            {#if userInfos.best_move_score_top_percent !== undefined && (userInfos.best_move_score_top_percent ?? 0) > 0}
                <div class="text-[11px] text-gray-500 mt-1">Top {Math.round(userInfos.best_move_score_top_percent)}%</div>
            {/if}
        </div>
    </div>
</div>
