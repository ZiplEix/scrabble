export function registerSW(onNeedRefresh: () => void) {
	if ('serviceWorker' in navigator) {
		navigator.serviceWorker.register('/service-worker.js').then(reg => {
			reg.onupdatefound = () => {
				const newSW = reg.installing;
				if (newSW) {
					newSW.onstatechange = () => {
						if (newSW.state === 'installed' && navigator.serviceWorker.controller) {
							// Un nouveau SW est prêt, on peut proposer de recharger
							onNeedRefresh();
						}
					};
				}
			};
		});
	}
}
