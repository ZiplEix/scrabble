<script lang='ts'>
    import { onMount } from "svelte";

    type AnnouncementLevel = 'info' | 'feature' | 'important';

    type Announcement = {
        id: string;
        title: string;
        message: string;
        date: string;
        level: AnnouncementLevel;
        actionLabel?: string;
        actionHref?: string;
    };

    const announcements: Announcement[] = [
        {
            id: '2026-05-23-scrabby-bot',
            title: '🤖 Nouveau : Défiez notre IA Scrabby !',
            message:
                "Un nouvel adversaire de taille rejoint la communauté ! Scrabby est une intelligence artificielle redoutable conçue pour chercher les meilleurs coups en temps réel.\n\n" +
                "Pour l'affronter :\n" +
                "1. Rendez-vous sur la page de création de partie.\n" +
                "2. Cliquez sur le bouton d'invitation rapide \"Défier Scrabby\" en un clic, ou ajoutez manuellement \"Scrabby\" dans les joueurs invités.\n\n" +
                "Êtes-vous prêts à relever le défi et battre l'ordinateur ?\n" +
                "(Même ChatGPT n'a pas réussi à la battre, je ne sais plus quoi faire pour gagner...)",
            date: '23 mai 2026',
            level: 'feature',
            actionLabel: 'Défier Scrabby',
            actionHref: '/games/new'
        },
        {
            id: '2026-05-22-systeme-amis',
            title: "👥 Nouveau : Système d'Amis",
            message:
                "Ajoutez vos partenaires de jeu favoris en un clic ! Visitez le profil public d'un joueur ou consultez le classement général pour l'ajouter en ami.\n\n" +
                "Votre liste d'amis est directement accessible sur votre profil pour suivre leur IPS et gérer vos relations.",
            date: '22 mai 2026',
            level: 'feature',
            actionLabel: 'Voir mes amis',
            actionHref: '/me'
        },
        {
            id: '2026-05-22-creation-partie-amelioree',
            title: '🎮 Création de partie en 1 clic !',
            message:
                "Inviter des adversaires n'a jamais été aussi rapide. La création de partie intègre désormais deux onglets sociaux :\n" +
                "1. 👤 Mes Amis : retrouvez instantanément vos amis.\n" +
                "2. ⚔️ Adversaires Récents : invitez à nouveau vos derniers rivaux.\n\n" +
                "Cliquez simplement sur leurs fiches pour les ajouter à votre partie !",
            date: '22 mai 2026',
            level: 'feature',
            actionLabel: 'Créer une partie',
            actionHref: '/games/new'
        },
        {
            id: '2026-05-22-succes-badges',
            title: '🏆 Succès & Badges : Relevez le défi !',
            message:
                "Gagnez des badges uniques et débloquez 21 succès inédits en temps réel au cours de vos parties (Bingo, Coup de Génie, Oiseau de Nuit, etc.). " +
                "Vos badges sont affichés avec fierté sur votre profil (/me) et visibles par toute la communauté !",
            date: '22 mai 2026',
            level: 'feature',
            actionLabel: 'Voir mes succès',
            actionHref: '/me'
        },
        {
            id: '2026-05-22-face-a-face',
            title: '⚔️ Nouveau : Statistiques Face-à-Face',
            message:
                "Analysez vos duels face aux autres joueurs directement sur leur profil public ! " +
                "Découvrez votre historique commun, le ratio de victoires respectives et vos scores moyens.",
            date: '22 mai 2026',
            level: 'feature'
        },
        {
            id: '2026-05-22-partage-canvas',
            title: '📸 Partagez vos grilles en image HD',
            message:
                "Un bouton de partage est désormais disponible sur l'écran de fin de partie. " +
                "Il génère instantanément une superbe carte PNG de votre grille finale au style Scrabble Club, idéale pour WhatsApp ou Signal.",
            date: '22 mai 2026',
            level: 'feature'
        },
        {
            id: '2026-05-21-pose-tactile',
            title: 'Pose rapide : Dites adieu au Drag & Drop !',
            message:
                "Posez vos tuiles d'une simple pression tactile ! Deux nouvelles méthodes ultra intuitives sont maintenant disponibles : \n" +
                "1. Sélectionnez d'abord une case vide sur le plateau (qui clignote en doré), puis touchez la lettre souhaitée sur votre chevalet.\n" +
                "2. Ou sélectionnez une lettre sur votre chevalet, puis touchez sa case de destination sur le plateau.\n" +
                "3. Le drag and drop est bien sur toujours disponible.",
            date: '21 mai 2026',
            level: 'feature',
            actionLabel: 'Découvrir maintenant',
            actionHref: '/'
        },
        {
            id: '2026-05-21-plateau-geant',
            title: 'Plateau géant optimisé pour mobile',
            message:
                "Pour un confort tactile maximal sur vos téléphones, le plateau occupe désormais 100% de la largeur disponible. " +
                "Les marges externes ont été optimisées et les arrondis de sélection ont été affinés pour un rendu épuré.",
            date: '21 mai 2026',
            level: 'info'
        },
        {
            id: '2026-05-21-puzzle-quotidien',
            title: 'Nouveau: défi quotidien',
            message:
                "Chaque jour, un puzzle est disponible avec un classement dédié. " +
                "Tu peux voir ta grille, comparer tes résultats et suivre ta progression.",
            date: '21 mai 2026',
            level: 'feature',
            actionLabel: 'Voir le puzzle du jour',
            actionHref: '/puzzles'
        }
    ];

    const LOCAL_KEY = 'dismissedAnnouncements';
    let dismissedIds = new Set<string>();
    let unseenAnnouncements: Announcement[] = [];
    let currentIndex = 0;
    let showNews = false;

    let currentAnnouncement: Announcement | null = null;
    let canGoPrev = false;
    let canGoNext = false;

    $: currentAnnouncement = unseenAnnouncements[currentIndex] ?? null;
    $: canGoPrev = currentIndex > 0;
    $: canGoNext = currentIndex < unseenAnnouncements.length - 1;

    onMount(() => {
        try {
            const raw = localStorage.getItem(LOCAL_KEY);
            const parsed = raw ? (JSON.parse(raw) as string[]) : [];
            dismissedIds = new Set(parsed);
        } catch (e) {
            console.warn('localStorage inaccessible pour les annonces', e);
            dismissedIds = new Set();
        }

        unseenAnnouncements = announcements.filter((a) => !dismissedIds.has(a.id));
        showNews = unseenAnnouncements.length > 0;
        currentIndex = 0;
    });

    function persistDismissed() {
        try {
            localStorage.setItem(LOCAL_KEY, JSON.stringify(Array.from(dismissedIds)));
        } catch (e) {
            console.warn('Impossible de sauvegarder les annonces lues', e);
        }
    }

    function closeCurrentAnnouncement() {
        if (!currentAnnouncement) {
            showNews = false;
            return;
        }

        dismissedIds.add(currentAnnouncement.id);
        persistDismissed();

        unseenAnnouncements = announcements.filter((a) => !dismissedIds.has(a.id));
        if (unseenAnnouncements.length === 0) {
            showNews = false;
            currentIndex = 0;
            return;
        }

        if (currentIndex >= unseenAnnouncements.length) {
            currentIndex = unseenAnnouncements.length - 1;
        }
    }

    function closeAllAnnouncements() {
        for (const announcement of unseenAnnouncements) {
            dismissedIds.add(announcement.id);
        }
        persistDismissed();
        unseenAnnouncements = [];
        currentIndex = 0;
        showNews = false;
    }

    function goPrev() {
        if (canGoPrev) {
            currentIndex -= 1;
        }
    }

    function goNext() {
        if (canGoNext) {
            currentIndex += 1;
        }
    }

    function onActionClick() {
        closeCurrentAnnouncement();
    }

    function levelClasses(level: AnnouncementLevel) {
        switch (level) {
            case 'important':
                return 'bg-rose-100 text-rose-800';
            case 'feature':
                return 'bg-emerald-100 text-emerald-800';
            default:
                return 'bg-sky-100 text-sky-800';
        }
    }

    function levelLabel(level: AnnouncementLevel) {
        switch (level) {
            case 'important':
                return 'Important';
            case 'feature':
                return 'Nouveaute';
            default:
                return 'Info';
        }
    }
</script>

{#if showNews && currentAnnouncement}
    <div class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-black/50"></div>

        <div class="relative z-10 w-full max-w-xl rounded-2xl bg-white shadow-2xl">
            <div class="p-6 sm:p-7">
                <div class="mb-4 flex items-center justify-between gap-3">
                    <span class={`inline-flex rounded-full px-3 py-1 text-xs font-semibold ${levelClasses(currentAnnouncement.level)}`}>
                        {levelLabel(currentAnnouncement.level)}
                    </span>
                    <span class="text-xs text-gray-500">{currentAnnouncement.date}</span>
                </div>

                <h2 class="text-xl font-semibold text-gray-900">{currentAnnouncement.title}</h2>
                <p class="mt-3 text-sm leading-relaxed text-gray-600 whitespace-pre-line">{currentAnnouncement.message}</p>

                {#if currentAnnouncement.actionLabel && currentAnnouncement.actionHref}
                    <a
                        class="mt-5 inline-flex rounded-lg bg-emerald-600 px-4 py-2 text-sm font-medium text-white hover:bg-emerald-700"
                        href={currentAnnouncement.actionHref}
                        onclick={onActionClick}
                    >
                        {currentAnnouncement.actionLabel}
                    </a>
                {/if}

                <div class="mt-6 flex flex-wrap items-center justify-between gap-2 border-t border-gray-100 pt-4">
                    <div class="text-xs text-gray-500">
                        {currentIndex + 1} / {unseenAnnouncements.length}
                    </div>

                    <div class="flex items-center gap-2">
                        <button
                            class="rounded-md border border-gray-200 px-3 py-2 text-sm text-gray-700 disabled:opacity-40"
                            onclick={goPrev}
                            disabled={!canGoPrev}
                        >
                            Precedent
                        </button>
                        <button
                            class="rounded-md border border-gray-200 px-3 py-2 text-sm text-gray-700 disabled:opacity-40"
                            onclick={goNext}
                            disabled={!canGoNext}
                        >
                            Suivant
                        </button>
                        <button
                            class="rounded-md bg-gray-900 px-3 py-2 text-sm text-white hover:bg-black"
                            onclick={closeAllAnnouncements}
                        >
                            Tout marquer comme lu
                        </button>
                    </div>
                </div>
            </div>
        </div>
    </div>
{/if}