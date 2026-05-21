<script lang="ts">
	import { api } from '$lib/api';
	import { user } from '$lib/stores/user';
	import { goto } from '$app/navigation';
	import HeaderBar from '$lib/components/HeaderBar.svelte';

	let username = $state('');
	let password = $state('');
	let error = $state('');
	let showPassword = $state(false);

	async function handleRegister() {
		error = '';
		try {
			const userNameToStore = username.trim().toLowerCase();
			const res = await api.post('/auth/register', { username: userNameToStore, password });
			user.set({ username: userNameToStore, token: res.data.token });
			goto('/');
		} catch (err: any) {
			error = err?.response?.data?.message || 'Échec de l’inscription';
		}
	}

	function togglePasswordVisibility() {
		showPassword = !showPassword;
	}
</script>

<HeaderBar title="Inscription" back={true} />

<main class="max-w-sm mx-auto px-4 py-8">
	<div class="text-center mb-8">
		<div class="inline-flex items-center justify-center w-14 h-14 rounded-2xl bg-brand-gold-light text-brand-gold shadow-sm mb-3 text-2xl font-bold select-none">
			✨
		</div>
		<h1 class="text-2xl font-extrabold text-stone-800 tracking-tight">Rejoignez le jeu !</h1>
		<p class="text-stone-500 text-xs mt-1">Créez votre compte en quelques secondes</p>
	</div>

	<div class="glass-card rounded-3xl p-6 border border-white/60 shadow-xl relative overflow-hidden">
		<form onsubmit={(e) => { e.preventDefault(); handleRegister(); }} class="flex flex-col gap-4">
			
			<div class="flex flex-col gap-1">
				<label for="username" class="text-xs font-semibold text-stone-600 px-1">Nom d'utilisateur</label>
				<input
					id="username"
					class="w-full bg-white/70 border border-stone-200/80 rounded-2xl px-4 py-3 text-sm placeholder-stone-400 focus:outline-none focus:ring-2 focus:ring-brand-emerald/40 focus:border-brand-emerald transition"
					type="text"
					placeholder="Ex: baptiste"
					bind:value={username}
					required
				/>
				<p class="text-[10px] text-stone-500 mt-1 px-1">
					Ce nom sera affiché pour vos proches et servira à vous inviter.
				</p>
			</div>

			<div class="flex flex-col gap-1">
				<label for="password" class="text-xs font-semibold text-stone-600 px-1">Mot de passe</label>
				<div class="relative">
					<input
						id="password"
						class="w-full bg-white/70 border border-stone-200/80 rounded-2xl px-4 py-3 text-sm placeholder-stone-400 focus:outline-none focus:ring-2 focus:ring-brand-emerald/40 focus:border-brand-emerald pr-11 transition"
						type={showPassword ? 'text' : 'password'}
						placeholder="••••••••"
						bind:value={password}
						required
					/>
					<button
						type="button"
						onclick={togglePasswordVisibility}
						class="absolute inset-y-0 right-0 px-3.5 flex items-center text-stone-400 hover:text-stone-600 active:scale-95 transition-all"
						aria-label={showPassword ? 'Cacher le mot de passe' : 'Montrer le mot de passe'}
					>
						{#if showPassword}
							<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
								<path stroke-linecap="round" stroke-linejoin="round" d="M13.875 18.825A10.05 10.05 0 0112 19c-5.523 0-10-4.477-10-10a9.96 9.96 0 012.277-6.176" />
								<path stroke-linecap="round" stroke-linejoin="round" d="M3 3l18 18" />
							</svg>
						{:else}
							<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
								<path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
								<path stroke-linecap="round" stroke-linejoin="round" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.477 0 8.268 2.943 9.542 7-1.274 4.057-5.065 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
							</svg>
						{/if}
					</button>
				</div>
			</div>

			{#if error}
				<div class="bg-red-50 border border-red-200 text-red-700 text-xs px-3.5 py-2.5 rounded-xl font-medium text-center">
					⚠️ {error}
				</div>
			{/if}

			<button
				type="submit"
				class="w-full mt-2 inline-flex items-center justify-center bg-brand-emerald hover:bg-brand-emerald-hover text-white py-3.5 px-6 rounded-2xl font-bold shadow-lg shadow-brand-emerald/10 active:scale-[0.98] transition-all cursor-pointer"
			>
				Créer mon compte
			</button>
		</form>
	</div>

	<p class="mt-8 text-sm text-center text-stone-500">
		Déjà inscrit ?
		<a href="/login" class="text-brand-emerald font-bold hover:underline">Se connecter</a>
	</p>
</main>
