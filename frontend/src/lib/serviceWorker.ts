export function registerSW(onNeedRefresh: () => void) {
	if ('serviceWorker' in navigator) {
		navigator.serviceWorker.register('/service-worker.js', { type: 'module' })
			.then((reg) => {
				let refreshing = false;

				// Ce listener s’assure qu’on recharge UNE SEULE FOIS
				navigator.serviceWorker.addEventListener('controllerchange', () => {
					if (refreshing) return;
					refreshing = true;
					window.location.reload();
				});

				reg.onupdatefound = () => {
					const newSW = reg.installing;
					if (newSW) {
						newSW.onstatechange = () => {
							if (
								newSW.state === 'installed' &&
								navigator.serviceWorker.controller
							) {
								// Nouveau SW installé mais pas encore contrôleur actif
								onNeedRefresh(); // on affiche la confirmation
							}
						};
					}
				};
			})
			.catch(console.error);
	}
}
