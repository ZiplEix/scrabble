<script lang='ts'>
  import type { Writable } from "svelte/store";
  import GameMenu from "./GameMenu.svelte";

    let { game, showScores, gameId }: {
        game: GameInfo;
        showScores: Writable<boolean>;
        gameId: string;
    } = $props()
</script>

<div class="flex items-center w-full justify-between">
    <!-- Title & infos -->
    <div class="w-full">
        <div class="px-3 pt-2 pb-1 w-full relative">
            <div class="flex items-center gap-2">
                <h2 class="flex-1 text-base font-semibold text-gray-800 truncate">{game.name}</h2>
            </div>
        </div>
        <!-- Sous-ligne compacte (hauteur minimale) -->
        <p class="px-3 mt-0.5 text-[11px] leading-tight text-gray-600 flex items-center justify-between">
            <span>
                Lettres restantes : <strong class="font-semibold">{game.remaining_letters}</strong>
            </span>
            <span>
                Tour : <strong class="font-semibold">{game.current_turn_username}</strong>
            </span>
        </p>
    </div>

    <!-- Menu bouton -->
    <GameMenu showScores={showScores} gameId={game.id} />
</div>
