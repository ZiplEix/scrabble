<script lang="ts">
    import { api } from "$lib/api";
  import { subscribeToPush } from "$lib/push";
    import type { UserInfos } from "$lib/types/user_infos";
  import { onMount } from "svelte";
    import Checkbox from "./Checkbox.svelte";

    let { userInfos = $bindable() }: { userInfos: UserInfos } = $props();

    let notificationsEnable = $state(userInfos?.notifications_enabled ?? false);
    let turn = $state(userInfos?.turn_notifications_enabled ?? false);
    let messages = $state(userInfos?.messages_notifications_enabled ?? false);

    let permission = $state<'granted' | 'denied' | 'default'>( 'default' );

    onMount(() => {
        if (Notification.permission === "granted") {
            permission = "granted";
        } else if (Notification.permission === "denied") {
            permission = "denied";
        } else {
            permission = "default";
        }
    })

    $effect(() => {
        if (userInfos) {
            // notificationsEnable = userInfos.notifications_enabled ?? notificationsEnable;
            turn = userInfos.turn_notifications_enabled ?? turn;
            messages = userInfos.messages_notifications_enabled ?? messages;
        }
    });

    async function savePrefs() {
        try {
            await api.put("/me/prefs", {
                turn,
                messages
            });
            userInfos = { ...userInfos,
                // notifications_enabled: notificationsEnable,
                turn_notifications_enabled: turn,
                messages_notifications_enabled: messages
            };
        } catch (err: any) {
            const error = err?.response?.data?.message || "Échec de la mise à jour des préférences";
            console.error(error);
            alert(error);
        }
    }

    async function askNotificationPermission() {
		const perm = await Notification.requestPermission();

		if (perm === "granted") {
			await subscribeToPush();
            userInfos = { ...userInfos, notifications_enabled: true };
		}
	}

    async function toggleNotification() {
        notificationsEnable = permission === 'granted';

        if (permission != 'granted') {
            console.log("Demande d'autorisation de notification");
            await askNotificationPermission();
        } else {
            userInfos = { ...userInfos, notifications_enabled: false };
            await api.delete('/notifications/push-subscribe');
        }
    }

    function toggleTurn() {
        turn = !turn;
        savePrefs();
    }

    function toggleMessages() {
        messages = !messages;
        savePrefs();
    }
</script>

<div class="rounded-2xl bg-white/90 backdrop-blur-md ring-1 ring-black/5 shadow-lg p-4">
    <h2 class="text-lg font-medium mb-3">Options</h2>
    <div class="mb-2">
        <!-- turn off/on all the types of notifications -->
        <div class="flex items-center justify-between">
            <div>
                <p class="text-sm font-medium">Notifications</p>
                <p class="text-xs text-gray-500">Autoriser les notifications sur cet appareil</p>
            </div>
            <div>
                <Checkbox
                    checked={permission === 'granted'}
                    onToggle={async () => {
                        await toggleNotification();
                    }} />
            </div>
        </div>
        <!-- Notifications type -->
        <div class="pl-4 pt-4 gap-2 flex flex-col">
            <div class="flex items-center justify-between">
                <div>
                    <p class="text-sm font-medium">A votre tour de jouer</p>
                    <p class="text-xs text-gray-500 pr-2">Activer ou désactiver les notifications lorsque c'est à votre tour de jouer dans une partie</p>
                </div>
                <Checkbox checked={turn} onToggle={toggleTurn} />
            </div>
            <div class="flex items-center justify-between">
                <div>
                    <p class="text-sm font-medium">Messages</p>
                    <p class="text-xs text-gray-500 pr-2">Activer ou désactiver les notifications lorsqu'un message est envoyé</p>
                </div>
                <Checkbox checked={messages} onToggle={toggleMessages} />
            </div>
        </div>
    </div>
</div>
