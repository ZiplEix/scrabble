<script lang="ts">
	import { onMount, createEventDispatcher } from 'svelte';

	interface Props {
		// startedAt et timeoutSeconds viennent du serveur — le client ne fait que l'affichage
		startedAt: Date;
		timeoutSeconds: number;
	}

	let { startedAt, timeoutSeconds }: Props = $props();

	const dispatch = createEventDispatcher<{
		timeout: void;
	}>();

	let timeRemaining = $state(0);
	let timerInterval: number | null = null;

	onMount(() => {
		updateRemaining();
		timerInterval = window.setInterval(updateRemaining, 1000);

		return () => {
			if (timerInterval) clearInterval(timerInterval);
		};
	});

	function updateRemaining() {
		const elapsed = (Date.now() - startedAt.getTime()) / 1000;
		const remaining = timeoutSeconds - elapsed;

		if (remaining <= 0) {
			timeRemaining = 0;
			if (timerInterval) clearInterval(timerInterval);
			dispatch('timeout');
		} else {
			timeRemaining = Math.ceil(remaining);
		}
	}

	function formatTime(seconds: number): string {
		const mins = Math.floor(seconds / 60);
		const secs = seconds % 60;
		return `${mins}:${secs.toString().padStart(2, '0')}`;
	}

	function getTimerColor(): string {
		if (timeRemaining <= 30) return 'text-red-600';
		if (timeRemaining <= 60) return 'text-orange-600';
		return 'text-emerald-600';
	}

	function getProgressPercentage(): number {
		return Math.max(0, (timeRemaining / timeoutSeconds) * 100);
	}
</script>

<div class="flex flex-col items-center gap-2">
	<div class="text-center">
		<p class="text-sm text-gray-600 font-medium">Temps restant</p>
		<div class={`text-4xl font-bold ${getTimerColor()} font-mono`}>
			{formatTime(timeRemaining)}
		</div>
	</div>

	<!-- Progress bar -->
	<div class="w-40 h-2 bg-gray-200 rounded-full overflow-hidden">
		<div
			class={`h-full transition-all duration-1000 ${
				timeRemaining <= 30 ? 'bg-red-500' : timeRemaining <= 60 ? 'bg-orange-500' : 'bg-emerald-500'
			}`}
			style="width: {getProgressPercentage()}%"
		/>
	</div>

	{#if timeRemaining <= 30 && timeRemaining > 0}
		<p class="text-xs font-semibold text-red-600">Dépêchez-vous!</p>
	{:else if timeRemaining === 0}
		<p class="text-xs font-semibold text-red-700">Temps écoulé!</p>
	{/if}
</div>
