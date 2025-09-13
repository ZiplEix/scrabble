import { build, files, version } from '$service-worker';

// ---- CACHE & OFFLINE ----
const CACHE = `cache-${version}`;
// 'build' = assets générés par Vite, 'files' = tout ce qui vient de /static
const ASSETS = [...build, ...files, '/offline.html']; // ajoute un offline.html dans /static

self.addEventListener('install', (event) => {
	event.waitUntil((async () => {
		const cache = await caches.open(CACHE);
		await cache.addAll(ASSETS);
	})());
});

// self.addEventListener('activate', (event) => {
// 	event.waitUntil(self.clients.claim());
// });
self.addEventListener('activate', (event) => {
	event.waitUntil((async () => {
		const keys = await caches.keys();
		await Promise.all(keys.map((k) => (k !== CACHE ? caches.delete(k) : Promise.resolve())));
		await (self).clients.claim();
	})());
});

self.addEventListener('fetch', (event) => {
	const req = event.request;

	// Stratégie "network falling back to cache" pour la navigation
	if (req.mode === 'navigate') {
		event.respondWith((async () => {
		try {
			const fresh = await fetch(req);
			return fresh;
		} catch {
			const cache = await caches.open(CACHE);
			return (await cache.match('/offline.html')) || Response.error();
		}
		})());
		return;
	}

	// Cache-first pour les assets connus
	event.respondWith((async () => {
		const cached = await caches.match(req);
		return cached || fetch(req);
	})());
});

// ---- PUSH NOTIFICATIONS ----
self.addEventListener('push', (event) => {
	const data = (event.data && event.data.json()) || {};
	const title = data.title || 'Nouvelle notification';
	const options = {
		body: data.body || '',
		icon: '/icons/icon-192.png',
		badge: '/icons/icon-192.png',
		data: data.url || '/'
	};
	event.waitUntil((self).registration.showNotification(title, options));
});

self.addEventListener('notificationclick', (event) => {
	event.notification.close();
	event.waitUntil((self).clients.openWindow(event.notification.data));
});

// self.addEventListener('push', (event) => {
// 	const data = event.data?.json() || {};
// 	const title = data.title || 'Nouvelle notification';
// 	const options = {
// 		body: data.body || '',
// 		icon: '/icons/icon-192.png',
// 		badge: '/icons/icon-192.png',
// 		data: data.url || '/',
// 	};

// 	event.waitUntil(self.registration.showNotification(title, options));
// });

// self.addEventListener('notificationclick', (event) => {
// 	event.notification.close();
// 	event.waitUntil(clients.openWindow(event.notification.data));
// });
