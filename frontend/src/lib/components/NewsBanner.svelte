<script lang='ts'>
    import { onMount } from "svelte";

    // --- Nouveautés / banderole ---
	type NewsBanner = {
		title: string;
		message: string;
		image?: string | null;
	};

	// TODO: Replace this object when you want to show a different nouveauté.
	const newsBanner: NewsBanner = {
        title: 'Nouveau classement Elo et système de rangs',
        message:
            "Le classement Elo est maintenant disponible, avec un système de rangs: Fer, Bronze, Argent, Or et Platine. " +
            "Ton rang évolue automatiquement selon tes résultats en partie. " +
            "Retrouve tous les détails dans la page Classement > Infos.",
        image: '/news/elo_ranks.png'
	};

	const LOCAL_KEY = 'closedBannerTitle';
	let showNews = false;

    onMount(() => {
        try {
            const closed = localStorage.getItem(LOCAL_KEY);
            if (closed !== newsBanner.title) {
                showNews = true;
            }
        } catch (e) {
            console.warn('localStorage inaccessible pour la banderole', e);
        }
    })

    function closeNews() {
		showNews = false;

		try {
			localStorage.setItem(LOCAL_KEY, newsBanner.title);
		} catch (e) {
			console.warn('Impossible de sauvegarder la préférence de la banderole', e);
		}
	}
</script>

{#if showNews}
    <div class="fixed inset-0 z-50 flex items-center justify-center">
        <div class="absolute inset-0 bg-black/40"></div>
        <div class="relative bg-white rounded-lg shadow-lg max-w-2xl w-full mx-4 z-10">
            <button aria-label="Fermer" class="absolute top-3 right-3 text-gray-300 hover:text-gray-700" onclick={closeNews}>
                <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                </svg>
            </button>

            <div class="flex flex-col items-center justify-center">
                {#if newsBanner.image}
                    <img src={newsBanner.image} alt="Annonce : white tile" class="w-full h-auto rounded" />
                {/if}
                <h2 class="text-lg font-semibold text-gray-900 mt-3 px-4 text-center">{newsBanner.title}</h2>
                <p class="text-sm text-gray-600 mt-2 px-4 pb-4 text-center">{newsBanner.message}</p>
            </div>
        </div>
    </div>
{/if}