self.addEventListener('activate', (event) => {
	event.waitUntil(self.clients.claim());
});

self.addEventListener('push', (event) => {
	const data = event.data?.json() || {};
	const title = data.title || 'Nouvelle notification';
	const options = {
		body: data.body || '',
		icon: '/icons/icon-192.png',
		badge: '/icons/icon-192.png',
		data: data.url || '/',
	};

	event.waitUntil(self.registration.showNotification(title, options));
});

self.addEventListener('notificationclick', (event) => {
	event.notification.close();
	event.waitUntil(clients.openWindow(event.notification.data));
});
