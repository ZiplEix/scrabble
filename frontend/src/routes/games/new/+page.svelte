<script lang="ts">
	import { api } from '$lib/api';
	import { onDestroy } from 'svelte';
	import HeaderBar from '$lib/components/HeaderBar.svelte';

	let name = '';
	let error = '';
	let loading = false;

	let newPlayer = '';
	let players: string[] = [];

	let suggestions: string[] = [];

	let debounceTimeout: ReturnType<typeof setTimeout>;

	$: if (newPlayer.length >= 2) {
		clearTimeout(debounceTimeout);
		debounceTimeout = setTimeout(async () => {
			try {
				const res = await api.get('/users/suggest?q=' + encodeURIComponent(newPlayer));
				suggestions = res.data.map((u: any) => u.username);
			} catch (e) {
				suggestions = [];
			}
		}, 200);
	} else {
		suggestions = [];
	}

	onDestroy(() => {
		clearTimeout(debounceTimeout);
	});

	function addPlayer() {
		const trimmed = newPlayer.trim();
		if (trimmed && !players.includes(trimmed)) {
			players = [...players, trimmed];
			newPlayer = '';
		}
	}

	function removePlayer(player: string) {
		players = players.filter(p => p !== player);
	}

	async function createGame(event: Event) {
        event.preventDefault();
		error = '';
		if (name.trim().length === 0) {
			error = 'Le nom de la partie est obligatoire';
			return;
		}

		loading = true;
		try {
			const res = await api.post('/game', {
				name,
				players,
			});
			const gameId = res.data.game_id;
			window.location.href = `/games/${gameId}`; // for some reason goto doesn't work here
		} catch (e: any) {
			alert(e?.response?.data?.message)
			error = e?.response?.data?.error || 'Erreur lors de la création de la partie';
		} finally {
			loading = false;
		}
	}
	let _ = HeaderBar;
</script>

<HeaderBar title="Nouvelle partie" back={true} />
<main class="max-w-sm mx-auto px-4 py-6">
	<h1 class="text-2xl font-bold mb-6 text-center text-gray-800">Créer une nouvelle partie</h1>

	<form onsubmit={createGame} class="flex flex-col gap-4">
		<input
			class="border rounded px-4 py-3 text-sm focus:outline-none focus:ring-2 focus:ring-green-500"
			type="text"
			placeholder="Nom de la partie"
			bind:value={name}
			required
		/>

		<div class="flex gap-2">
			<input
				list="user-suggestions"
				class="border rounded px-4 py-3 text-sm flex-grow focus:outline-none focus:ring-2 focus:ring-green-500"
				type="text"
				placeholder="Ajouter un joueur (ex: alice)"
				bind:value={newPlayer}
			/>
			<datalist id="user-suggestions">
				{#each suggestions as user}
					<option value={user} ></option>
				{/each}
			</datalist>

			<button
				type="button"
						onclick={addPlayer}
				class="bg-blue-600 hover:bg-blue-700 text-white px-4 py-3 text-sm font-semibold rounded"
			>
				Ajouter
			</button>
		</div>

		{#if players.length > 0}
			<ul class="space-y-2">
				{#each players as player}
					<li class="flex justify-between items-center bg-gray-100 px-3 py-2 rounded text-sm">
						<span>{player}</span>
						<button
							type="button"
							class="text-red-500 hover:text-red-700 font-bold"
								onclick={() => removePlayer(player)}
						>
							✕
						</button>
					</li>
				{/each}
			</ul>
		{/if}

		{#if error}
			<p class="text-sm text-red-600 text-center">{error}</p>
		{/if}

		<button
			type="submit"
			class="bg-green-600 hover:bg-green-700 text-white rounded py-3 font-semibold transition disabled:opacity-50"
			disabled={loading}
		>
			{loading ? 'Création...' : 'Créer la partie'}
		</button>
	</form>
</main>
