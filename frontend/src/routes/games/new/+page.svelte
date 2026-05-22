<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { api } from '$lib/api';
	import HeaderBar from '$lib/components/HeaderBar.svelte';
	import RankBadge from '$lib/components/RankBadge.svelte';
	import { goto } from '$app/navigation';
	import type { FriendInfo } from '$lib/types/user_infos';

	let name = $state('');
	let error = $state('');
	let loading = $state(false);

	let newPlayer = $state('');
	let players = $state<string[]>([]);
	let suggestions = $state<string[]>([]);

	let friends = $state<FriendInfo[]>([]);
	let recentOpponents = $state<FriendInfo[]>([]);
	let activeTab = $state<'friends' | 'recent'>('friends');
	let loadingSocial = $state(true);

	let debounceTimeout: ReturnType<typeof setTimeout>;

	onMount(async () => {
		try {
			const [friendsRes, recentRes] = await Promise.all([
				api.get('/users/friends'),
				api.get('/users/recent-opponents')
			]);
			friends = friendsRes.data;
			recentOpponents = recentRes.data;
			// Basculer sur l'onglet adversaires s'il n'y a pas d'amis mais qu'on a déjà joué
			if (friends.length === 0 && recentOpponents.length > 0) {
				activeTab = 'recent';
			}
		} catch (e) {
			console.error('Failed to load social lists:', e);
		} finally {
			loadingSocial = false;
		}
	});

	$effect(() => {
		if (newPlayer.length >= 2) {
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
	});

	onDestroy(() => {
		clearTimeout(debounceTimeout);
	});

	function handleAddManual() {
		const trimmed = newPlayer.trim();
		if (trimmed) {
			if (!players.includes(trimmed)) {
				players = [...players, trimmed];
			}
			newPlayer = '';
			suggestions = [];
		}
	}

	function togglePlayer(username: string) {
		const trimmed = username.trim();
		if (players.includes(trimmed)) {
			players = players.filter(p => p !== trimmed);
		} else {
			players = [...players, trimmed];
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
			goto(`/games/${gameId}`);
		} catch (e: any) {
			error = e?.response?.data?.error || 'Erreur lors de la création de la partie';
		} finally {
			loading = false;
		}
	}

	// Liste des couleurs d'initiales pour rendre l'UI dynamique
	const bgGradients = [
		'from-emerald-400 to-brand-emerald',
		'from-blue-400 to-blue-600',
		'from-purple-400 to-purple-600',
		'from-amber-400 to-brand-gold',
		'from-rose-400 to-rose-600'
	];

	function getAvatarColor(username: string): string {
		let hash = 0;
		for (let i = 0; i < username.length; i++) {
			hash = username.charCodeAt(i) + ((hash << 5) - hash);
		}
		const index = Math.abs(hash) % bgGradients.length;
		return bgGradients[index];
	}
</script>

<HeaderBar title="Nouvelle partie" back={true} />

<main class="max-w-xl mx-auto px-4 py-6 flex flex-col gap-6">
	<!-- HEADER DE LA PAGE -->
	<header class="text-center">
		<h1 class="text-2xl font-black text-stone-800">Lancer un défi 🎲</h1>
		<p class="text-xs text-stone-500 mt-1">Créez une grille de Scrabble Club et invitez vos adversaires.</p>
	</header>

	<form onsubmit={createGame} class="flex flex-col gap-5">
		<!-- CONFIGURATION DE BASE -->
		<div class="rounded-3xl bg-white border border-stone-200/50 p-5 shadow-sm flex flex-col gap-4">
			<div>
				<label for="game-name" class="block text-xs font-black text-stone-500 uppercase tracking-wider mb-2">Nom de la partie</label>
				<input
					id="game-name"
					class="w-full bg-stone-50/50 border border-stone-200 rounded-2xl px-4 py-3 text-sm placeholder-stone-400 focus:outline-none focus:ring-2 focus:ring-brand-emerald/40 focus:border-brand-emerald shadow-inner transition"
					type="text"
					placeholder="Ex: Duel au sommet, Partie du Dimanche..."
					bind:value={name}
					required
				/>
			</div>

			<!-- RECHERCHE MANUELLE -->
			<div class="relative">
				<label for="manual-player" class="block text-xs font-black text-stone-500 uppercase tracking-wider mb-2">Inviter un autre joueur</label>
				<div class="flex gap-2">
					<div class="relative flex-grow">
						<input
							id="manual-player"
							list="user-suggestions"
							class="w-full bg-stone-50/50 border border-stone-200 rounded-2xl px-4 py-3 text-sm placeholder-stone-400 focus:outline-none focus:ring-2 focus:ring-brand-emerald/40 focus:border-brand-emerald shadow-inner transition"
							type="text"
							placeholder="Rechercher par pseudo (ex: alice)"
							bind:value={newPlayer}
							onkeydown={(e) => { if (e.key === 'Enter') { e.preventDefault(); handleAddManual(); } }}
						/>
						<datalist id="user-suggestions">
							{#each suggestions as user}
								<option value={user}></option>
							{/each}
						</datalist>
					</div>

					<button
						type="button"
						onclick={handleAddManual}
						class="bg-brand-emerald hover:bg-brand-emerald-hover text-white px-5 py-3 text-sm font-black rounded-2xl transition shadow-sm active:scale-95 cursor-pointer shrink-0"
					>
						Ajouter
					</button>
				</div>
			</div>
		</div>

		<!-- JOUEURS ACTUELLEMENT INVITÉS -->
		<div class="rounded-3xl bg-white border border-stone-200/50 p-5 shadow-sm">
			<span class="block text-xs font-black text-stone-500 uppercase tracking-wider mb-3">Participants invités ({players.length})</span>
			
			{#if players.length === 0}
				<div class="py-4 text-center rounded-2xl border border-dashed border-stone-200 bg-stone-50/30">
					<span class="text-xl block">👥</span>
					<p class="text-xs text-stone-400 font-bold mt-1">Aucun participant invité pour l'instant</p>
					<p class="text-[10px] text-stone-400 mt-0.5">Sélectionnez vos amis ci-dessous ou cherchez par pseudo.</p>
				</div>
			{:else}
				<div class="flex flex-wrap gap-2">
					{#each players as player}
						<div class="inline-flex items-center gap-2 bg-brand-emerald/10 border border-brand-emerald/20 pl-2 pr-3 py-1.5 rounded-full text-xs font-extrabold text-brand-emerald shadow-sm transition animate-fade-in animate-duration-200">
							<!-- Initiale -->
							<span class={`w-5 h-5 rounded-full bg-gradient-to-br ${getAvatarColor(player)} text-white flex items-center justify-center text-[10px] font-black`}>
								{player.charAt(0).toUpperCase()}
							</span>
							<span>{player}</span>
							<button
								type="button"
								class="text-brand-emerald hover:text-rose-600 font-black cursor-pointer transition ml-1"
								onclick={() => removePlayer(player)}
								aria-label={`Retirer ${player}`}
							>
								✕
							</button>
						</div>
					{/each}
				</div>
			{/if}
		</div>

		<!-- ACTION FINALE CREATION -->
		{#if error}
			<p class="text-xs font-bold text-rose-600 text-center bg-rose-50 border border-rose-200 p-3 rounded-2xl">{error}</p>
		{/if}

		<button
			type="submit"
			class="bg-brand-emerald hover:bg-brand-emerald-hover text-white rounded-2xl py-4 font-black shadow-lg shadow-brand-emerald/10 transition active:scale-[0.99] disabled:opacity-50 cursor-pointer text-center"
			disabled={loading}
		>
			{loading ? 'Création de la partie en cours...' : 'Créer et lancer la partie !'}
		</button>
	</form>

	<!-- ESPACE SOCIAL (AMIS & ADVERSAIRES RECENTS) -->
	<section class="mt-4 flex flex-col gap-4">
		<h2 class="text-xs font-black text-stone-500 uppercase tracking-wider">Inviter en un clic</h2>

		<!-- Onglets -->
		<div class="rounded-2xl bg-stone-200/50 p-1 border border-stone-200/20">
			<div class="grid grid-cols-2 gap-1">
				<button
					type="button"
					class="py-2.5 px-3 text-xs font-bold rounded-xl text-center transition cursor-pointer {activeTab === 'friends' ? 'bg-white text-brand-emerald shadow-sm' : 'text-stone-600 hover:text-stone-800'}"
					onclick={() => activeTab = 'friends'}
				>
					👤 Mes Amis ({friends.length})
				</button>
				<button
					type="button"
					class="py-2.5 px-3 text-xs font-bold rounded-xl text-center transition cursor-pointer {activeTab === 'recent' ? 'bg-white text-brand-emerald shadow-sm' : 'text-stone-600 hover:text-stone-800'}"
					onclick={() => activeTab = 'recent'}
				>
					⚔️ Adversaires Récents ({recentOpponents.length})
				</button>
			</div>
		</div>

		<!-- Listes dynamiques -->
		{#if loadingSocial}
			<div class="flex items-center justify-center py-10 gap-2">
				<div class="w-5 h-5 border-2 border-brand-emerald border-t-transparent rounded-full animate-spin"></div>
				<p class="text-xs text-stone-400 font-bold">Chargement des joueurs…</p>
			</div>
		{:else}
			{@const currentList = activeTab === 'friends' ? friends : recentOpponents}

			{#if currentList.length === 0}
				<div class="py-8 px-6 text-center rounded-3xl bg-white border border-stone-200/50 shadow-sm">
					{#if activeTab === 'friends'}
						<span class="text-2xl block mb-1">🔍</span>
						<p class="text-xs text-stone-700 font-black">Aucun ami pour le moment</p>
						<p class="text-[10px] text-stone-400 mt-1 leading-normal max-w-xs mx-auto">
							Allez sur le profil d'un joueur ou d'un membre du classement et cliquez sur <strong>👤+ Ajouter en ami</strong> pour l'afficher ici !
						</p>
					{:else}
						<span class="text-2xl block mb-1">🎮</span>
						<p class="text-xs text-stone-700 font-black">Aucun historique de partie</p>
						<p class="text-[10px] text-stone-400 mt-1 leading-normal max-w-xs mx-auto">
							Vos anciens adversaires s'afficheront automatiquement ici dès que vous aurez joué avec eux.
						</p>
					{/if}
				</div>
			{:else}
				<div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
					{#each currentList as player}
						{@const isSelected = players.includes(player.username)}
						<button
							type="button"
							onclick={() => togglePlayer(player.username)}
							class="w-full text-left rounded-2xl p-3 bg-white border transition shadow-sm hover:shadow active:scale-[0.98] cursor-pointer flex items-center justify-between gap-3
							{isSelected ? 'border-brand-emerald ring-1 ring-brand-emerald bg-brand-emerald/5' : 'border-stone-200/60 hover:bg-stone-50/50'}"
						>
							<div class="flex items-center gap-3 min-w-0">
								<!-- Avatar monogramme -->
								<div class={`w-9 h-9 rounded-xl bg-gradient-to-br ${getAvatarColor(player.username)} text-white flex items-center justify-center font-black shadow-sm shrink-0`}>
									{player.username.charAt(0).toUpperCase()}
								</div>
								
								<div class="min-w-0">
									<p class="text-xs font-black text-stone-800 truncate">{player.username}</p>
									<div class="flex items-center gap-1 mt-0.5">
										<RankBadge rating={player.rating} size="sm" />
										<span class="text-[9px] font-bold text-stone-500">{player.rating} IPS</span>
									</div>
								</div>
							</div>

							<!-- Check / Plus -->
							<div class="shrink-0">
								{#if isSelected}
									<span class="w-6 h-6 rounded-full bg-brand-emerald text-white flex items-center justify-center text-xs font-black shadow-sm">
										✓
									</span>
								{:else}
									<span class="w-6 h-6 rounded-full bg-stone-100 hover:bg-brand-emerald/10 text-stone-500 hover:text-brand-emerald flex items-center justify-center text-sm font-black border border-stone-200 transition">
										+
									</span>
								{/if}
							</div>
						</button>
					{/each}
				</div>
			{/if}
		{/if}
	</section>
</main>
