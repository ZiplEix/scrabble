import { env } from "$env/dynamic/public";
import { api } from "./api";

const VAPID_PUBLIC_KEY = env.PUBLIC_VAPID_PUBLIC_KEY

export async function subscribeToPush() {
    if (!('serviceWorker' in navigator) || !('PushManager' in window)) return;

    const permission = await Notification.requestPermission();
    if (permission !== 'granted') return;

    const reg = await navigator.serviceWorker.ready;
    const subscription = await reg.pushManager.subscribe({
        userVisibleOnly: true,
        applicationServerKey: urlBase64ToUint8Array(VAPID_PUBLIC_KEY)
    });

    await api.post('/notifications/push-subscribe', JSON.stringify(subscription));
}

function urlBase64ToUint8Array(base64String: string): Uint8Array {
   const padding = '='.repeat((4 - (base64String.length % 4)) % 4);
	const base64 = (base64String + padding).replace(/-/g, '+').replace(/_/g, '/');
	const rawData = atob(base64);
	return Uint8Array.from([...rawData].map(char => char.charCodeAt(0)));
}
