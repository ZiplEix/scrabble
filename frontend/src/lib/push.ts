import { env } from "$env/dynamic/public";
import { api } from "./api";

const VAPID_PUBLIC_KEY = env.PUBLIC_VAPID_PUBLIC_KEY

export async function subscribeToPush() {
    console.log('Attempting to subscribe to push notifications...');
    if (!('serviceWorker' in navigator) || !('PushManager' in window)) return;

    console.log('Service Worker and Push Manager are supported.');

    const permission = await Notification.requestPermission();

    console.log('Notification permission:', permission);

    if (permission !== 'granted') return;

    console.log('Notification permission granted.');

    const reg = await navigator.serviceWorker.ready;

    console.log('Service Worker ready:', reg);

    const subscription = await reg.pushManager.subscribe({
        userVisibleOnly: true,
        applicationServerKey: urlBase64ToUint8Array(VAPID_PUBLIC_KEY)
    });

    console.log('Push subscription:', subscription);

    await api.post('/notifications/push-subscribe', JSON.stringify(subscription));

    console.log('Push subscription saved to server.');
}

function urlBase64ToUint8Array(base64String: string): Uint8Array {
   const padding = '='.repeat((4 - (base64String.length % 4)) % 4);
	const base64 = (base64String + padding).replace(/-/g, '+').replace(/_/g, '/');
	const rawData = atob(base64);
	return Uint8Array.from([...rawData].map(char => char.charCodeAt(0)));
}
