<script lang="ts">
    import type { UserInfos } from "$lib/types/user_infos";
  import GameList from "./GameList.svelte";

    let { userInfos }: { userInfos: UserInfos } = $props();
</script>

<div>
    <h2 class="text-lg font-medium mb-3">Statistiques</h2>

    <div class="space-y-3">
        <div class="p-3 bg-gray-50 rounded">
            <div class="flex items-center justify-between">
                <span class="text-sm text-gray-700">Parties jouées</span>
                <span class="text-sm font-medium text-gray-900">{userInfos.games_count}</span>
            </div>
            {#if userInfos.games_count_top_percent !== undefined && (userInfos.games_count_top_percent ?? 0) > 0}
                <div class="text-[11px] text-gray-500 mt-1">Top {Math.round(userInfos.games_count_top_percent)}%</div>
            {/if}
        </div>
        <div class="p-3 bg-gray-50 rounded">
            <div class="flex items-center justify-between">
                <span class="text-sm text-gray-700">Victoires</span>
                <span class="text-sm font-medium text-gray-900">{userInfos.victories}</span>
            </div>
            {#if userInfos.victories_top_percent !== undefined && (userInfos.victories_top_percent ?? 0) > 0}
                <div class="text-[11px] text-gray-500 mt-1">Top {Math.round(userInfos.victories_top_percent)}%</div>
            {/if}
        </div>
        <div class="p-3 bg-gray-50 rounded">
            <div class="flex items-center justify-between">
                <span class="text-sm text-gray-700">Score moyen</span>
                <span class="text-sm font-medium text-gray-900">{Math.round(userInfos.avg_score ?? 0)}</span>
            </div>
            {#if userInfos.avg_score_top_percent !== undefined && (userInfos.avg_score_top_percent ?? 0) > 0}
                <div class="text-[11px] text-gray-500 mt-1">Top {Math.round(userInfos.avg_score_top_percent)}%</div>
            {/if}
        </div>
        <div class="p-3 bg-gray-50 rounded">
            <div class="flex items-center justify-between">
                <span class="text-sm text-gray-700">Meilleur score</span>
                <span class="text-sm font-medium text-gray-900">{userInfos.best_score}</span>
            </div>
            {#if userInfos.best_score_top_percent !== undefined && (userInfos.best_score_top_percent ?? 0) > 0}
                <div class="text-[11px] text-gray-500 mt-1">Top {Math.round(userInfos.best_score_top_percent)}%</div>
            {/if}
        </div>
        <div class="p-3 bg-gray-50 rounded">
            <div class="flex items-center justify-between">
                <span class="text-sm text-gray-700">Moyenne par coup</span>
                <span class="text-sm font-medium text-gray-900">{Math.round(userInfos.avg_points_per_move ?? 0)}</span>
            </div>
            {#if userInfos.avg_points_per_move_top_percent !== undefined && (userInfos.avg_points_per_move_top_percent ?? 0) > 0}
                <div class="text-[11px] text-gray-500 mt-1">Top {Math.round(userInfos.avg_points_per_move_top_percent)}%</div>
            {/if}
        </div>
        <div class="p-3 bg-gray-50 rounded">
            <div class="flex items-center justify-between">
                <span class="text-sm text-gray-700">Meilleur coup</span>
                <span class="text-sm font-medium text-gray-900">{userInfos.best_move_score ?? 0}</span>
            </div>
            {#if userInfos.best_move_score_top_percent !== undefined && (userInfos.best_move_score_top_percent ?? 0) > 0}
                <div class="text-[11px] text-gray-500 mt-1">Top {Math.round(userInfos.best_move_score_top_percent)}%</div>
            {/if}
        </div>
    </div>
</div>
