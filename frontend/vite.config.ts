import tailwindcss from '@tailwindcss/vite';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';
import { type SvelteKitPWAOptions } from '@vite-pwa/sveltekit'

const pwaConfig: Partial<SvelteKitPWAOptions> = {
	registerType: 'autoUpdate',
	manifest: {
		name: 'Scrabble Online',
		short_name: 'Scrabble',
		start_url: '/',
		display: 'standalone',
		background_color: '#ffffff',
		theme_color: '#00a86b',
		icons: [
			{
			src: '/icons/icon-192.png',
			sizes: '192x192',
			type: 'image/png'
			},
			{
			src: '/icons/icon-512.png',
			sizes: '512x512',
			type: 'image/png'
			}
		]
	}
}

export default defineConfig({
	plugins: [tailwindcss(), sveltekit()],
	server: {
		host: "0.0.0.0",
		port: 3000,
	}
});
