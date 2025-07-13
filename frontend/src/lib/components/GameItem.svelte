<script lang="ts">
	export let game: {
		id: string;
		name: string;
		current_turn_username: string;
	};
</script>

<a href={`/games/${game.id}`} class="block p-4 border rounded hover:bg-gray-100 transition">
	<h2 class="text-lg font-semibold">{game.name}</h2>
	<p class="text-sm text-gray-600">Tour de : <strong>{game.current_turn_username}</strong></p>
</a>
