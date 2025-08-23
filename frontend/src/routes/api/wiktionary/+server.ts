import type { RequestHandler } from '@sveltejs/kit';

// Proxy côté serveur pour contourner le CORS et éventuellement mettre en cache
export const GET: RequestHandler = async ({ url, fetch }) => {
	const title = url.searchParams.get('title');
	if (!title) {
		return new Response(JSON.stringify({ error: 'missing_title' }), { status: 400, headers: { 'content-type': 'application/json' } });
	}

	const params = new URLSearchParams({
		action: 'query',
		format: 'json',
		prop: 'extracts',
		// explaintext: '1',
		// redirects: '1',
		titles: title,
		origin: '*', // pas nécessaire côté serveur, mais inoffensif
	});
	const upstream = `https://fr.wiktionary.org/w/api.php?${params.toString()}`;

	try {
		const res = await fetch(upstream);
		if (!res.ok) {
			return new Response(
				JSON.stringify({ error: 'upstream_error', status: res.status }),
				{ status: 502, headers: { 'content-type': 'application/json' } }
			);
		}
		const data = await res.json();
		return new Response(JSON.stringify(data), {
			headers: {
				'content-type': 'application/json',
				'cache-control': 'public, max-age=3600'
			}
		});
	} catch (err) {
		return new Response(JSON.stringify({ error: 'fetch_failed' }), { status: 502, headers: { 'content-type': 'application/json' } });
	}
};
