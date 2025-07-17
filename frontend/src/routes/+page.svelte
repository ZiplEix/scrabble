<script lang="ts">
	import { user } from '$lib/stores/user';
  	import { onMount } from 'svelte';
	import LoginedPage from './LoginedPage.svelte';
	import NoLoginPage from './NoLoginPage.svelte';

	let deferredPrompt: any = null;
	let canInstall = false;

	onMount(() => {
		window.addEventListener('beforeinstallprompt', (e) => {
			e.preventDefault();
			deferredPrompt = e;
			canInstall = true;
		});
	});

	async function installApp() {
		if (deferredPrompt) {
			deferredPrompt.prompt();
			const { outcome } = await deferredPrompt.userChoice;
			if (outcome === 'accepted') {
				console.log('PWA installée');
			}
			deferredPrompt = null;
			canInstall = false;
		}
	}
</script>

<main class="max-w-4xl mx-auto px-6 py-10">
	{#if canInstall}
		<div class="bg-green-100 border border-green-300 text-green-800 px-4 py-3 rounded mb-6 flex items-center justify-between shadow">
			<div class="flex items-center gap-2">
				<span class="text-lg">📲</span>
				<p class="text-sm">
					Ajoutez ce site à votre écran d’accueil pour l’utiliser comme une app.
				</p>
			</div>
			<button
				on:click={installApp}
				class="ml-4 bg-green-600 hover:bg-green-700 text-white text-sm px-4 py-2 rounded"
			>
				Installer
			</button>
		</div>
	{/if}
	{#if !$user}
		<NoLoginPage />
	{:else}
		<LoginedPage />
	{/if}
</main>
