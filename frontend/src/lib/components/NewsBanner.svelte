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

    // Ajouter une nouvelle annonce en tete de liste.
    const announcements: Announcement[] = [
        {
            id: '2026-05-21-puzzle-quotidien',
            title: 'Nouveau: defi quotidien',
            message:
                "Chaque jour, un puzzle est disponible avec un classement dedie. " +
                "Tu peux voir ta grille, comparer tes resultats et suivre ta progression.",
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
                <p class="mt-3 text-sm leading-relaxed text-gray-600">{currentAnnouncement.message}</p>

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