import type { RequestHandler } from '@sveltejs/kit';

// Server-side proxy to bypass browser CORS limitations and fetch 1mot.net definitions safely
export const GET: RequestHandler = async ({ url, fetch }) => {
	const word = url.searchParams.get('word');
	if (!word) {
		return new Response(
			JSON.stringify({ error: 'missing_word' }),
			{ status: 400, headers: { 'content-type': 'application/json' } }
		);
	}

	const upstream = `https://1mot.net/${encodeURIComponent(word.toLowerCase())}`;

	try {
		const res = await fetch(upstream, {
			headers: {
				'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36',
				'Accept': 'text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8',
				'Accept-Language': 'fr,fr-FR;q=0.8,en-US;q=0.5,en;q=0.3'
			}
		});

		if (!res.ok) {
			if (res.status === 404) {
				return new Response(
					JSON.stringify({ error: 'not_found' }),
					{ status: 404, headers: { 'content-type': 'application/json' } }
				);
			}
			return new Response(
				JSON.stringify({ error: 'upstream_error', status: res.status }),
				{ status: 502, headers: { 'content-type': 'application/json' } }
			);
		}

		const html = await res.text();
		return new Response(
			JSON.stringify({ html, url: res.url }),
			{
				headers: {
					'content-type': 'application/json',
					'cache-control': 'public, max-age=86400' // Cache definitions for 1 day
				}
			}
		);
	} catch (err) {
		console.error(`1mot.net scrape failed for ${word}:`, err);
		return new Response(
			JSON.stringify({ error: 'fetch_failed' }),
			{ status: 502, headers: { 'content-type': 'application/json' } }
		);
	}
};
