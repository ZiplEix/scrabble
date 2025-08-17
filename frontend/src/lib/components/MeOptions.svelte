<script lang="ts">
    import type { UserInfos } from "$lib/types/user_infos";

    let { userInfos = $bindable() }: { userInfos: UserInfos } = $props();

    let checked = $state(userInfos?.notifications_enabled ?? false);

    $effect(() => {
        if (userInfos && userInfos.notifications_enabled !== checked)
            checked = userInfos.notifications_enabled;
    });

    function onToggle() {
        const next = !(userInfos?.notifications_enabled ?? false);
        console.log('Toggling notifications:', userInfos?.notifications_enabled, '->', next);
        userInfos = { ...userInfos, notifications_enabled: next };
        checked = next;
    }
</script>

<div>
    <h2 class="text-lg font-medium mb-4">Options</h2>
    <div class="flex items-center justify-between mb-4">
        <div>
            <p class="text-sm font-medium">Notifications</p>
            <p class="text-xs text-gray-500">Activer ou désactiver les notifications push</p>
        </div>
        <label class="inline-flex items-center cursor-pointer">
            <input
                type="checkbox"
                class="sr-only"
                bind:checked={checked}
                aria-label="Activer les notifications"
                onchange={onToggle}
            />
            <div class="w-12 h-7 rounded-full relative transition-colors duration-150"
                 aria-hidden="true"
                 class:bg-green-500={checked} class:bg-gray-300={!checked}>
                <span
                    class="absolute top-1 left-1 bg-white w-5 h-5 rounded-full shadow transition-transform duration-150"
                    style="transform-origin: center"
                    class:translate-x-5={checked}
                ></span>
            </div>
        </label>
    </div>

    <!-- <div class="flex items-center justify-between mb-4">
        <div>
            <p class="text-sm font-medium">Thème</p>
            <p class="text-xs text-gray-500">Basculer entre clair / sombre (placeholder)</p>
        </div>
        <label class="relative inline-flex items-center cursor-pointer">
            <input type="checkbox" bind:checked={darkMode} class="sr-only" aria-hidden="false" />
            <div class="w-11 h-6 bg-gray-200 rounded-full" aria-hidden="true"></div>
        </label>
    </div>

    <div class="mb-4">
        <label for="language-select" class="block text-sm font-medium mb-1">Langue</label>
        <select id="language-select" bind:value={language} class="w-full border rounded p-2 text-sm">
            <option>Français</option>
            <option>English</option>
        </select>
    </div> -->
</div>
