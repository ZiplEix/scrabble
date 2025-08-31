<script lang="ts">
    import type { UserInfos } from "$lib/types/user_infos";
  import GameList from "./GameList.svelte";

    let { userInfos }: { userInfos: UserInfos } = $props();
</script>

<div>
    <h2 class="text-lg font-medium mb-3">Statistiques</h2>

    <div class="space-y-3">
        <div class="flex items-center justify-between p-3 bg-gray-50 rounded">
            <span class="text-sm text-gray-700">Parties jouées</span>
            <span class="text-sm font-medium text-gray-900">{userInfos.games_count}</span>
        </div>
        <div class="flex items-center justify-between p-3 bg-gray-50 rounded">
            <span class="text-sm text-gray-700">Victoires</span>
            <span class="text-sm font-medium text-gray-900">{userInfos.victories}</span>
        </div>
        <div class="flex items-center justify-between p-3 bg-gray-50 rounded">
            <span class="text-sm text-gray-700">Score moyen</span>
            <span class="text-sm font-medium text-gray-900">{Math.round(userInfos.avg_score ?? 0)}</span>
        </div>
        <div class="flex items-center justify-between p-3 bg-gray-50 rounded">
            <span class="text-sm text-gray-700">Meilleur score</span>
            <span class="text-sm font-medium text-gray-900">{userInfos.best_score}</span>
        </div>
        <div class="flex items-center justify-between p-3 bg-gray-50 rounded">
            <span class="text-sm text-gray-700">Moyenne par coup</span>
            <span class="text-sm font-medium text-gray-900">{Math.round(userInfos.avg_points_per_move ?? 0)}</span>
        </div>
        <div class="flex items-center justify-between p-3 bg-gray-50 rounded">
            <span class="text-sm text-gray-700">Meilleur coup</span>
            <span class="text-sm font-medium text-gray-900">{userInfos.best_move_score ?? 0}</span>
        </div>
    </div>
</div>
