<script lang="ts">
	import { user } from '$lib/stores/user';
	import LoginedPage from './LoginedPage.svelte';
	import NoLoginPage from './NoLoginPage.svelte';

</script>

<main class="max-w-4xl mx-auto px-6 py-10">
	{#if !$user}
		<NoLoginPage />
	{:else}
		<LoginedPage />
	{/if}
</main>
