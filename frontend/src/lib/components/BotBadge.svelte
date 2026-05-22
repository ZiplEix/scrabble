<script lang="ts">
	// Svelte 5 — aucun prop nécessaire, c'est un badge statique
</script>

<span class="bot-badge">
	<span class="bot-icon">🤖</span>
	<span class="bot-label">Bot</span>
</span>

<style>
	.bot-badge {
		display: inline-flex;
		align-items: center;
		gap: 3px;
		padding: 2px 7px;
		border-radius: 999px;
		background: linear-gradient(135deg, hsl(270, 70%, 92%), hsl(250, 80%, 88%));
		border: 1px solid hsl(265, 60%, 78%);
		color: hsl(265, 55%, 45%);
		font-size: 10px;
		font-weight: 900;
		letter-spacing: 0.04em;
		text-transform: uppercase;
		white-space: nowrap;
		animation: bot-pulse 3s ease-in-out infinite;
	}

	.bot-icon {
		font-size: 11px;
		line-height: 1;
	}

	.bot-label {
		line-height: 1;
	}

	@keyframes bot-pulse {
		0%, 100% { box-shadow: 0 0 0 0 hsl(265, 60%, 78% / 0); }
		50%       { box-shadow: 0 0 0 4px hsl(265, 60%, 78% / 0.4); }
	}
</style>
