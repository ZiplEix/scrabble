module.exports = {
	content: ['./src/**/*.{html,js,svelte,ts}'],
	theme: {
		extend: {
			gridTemplateColumns: {
				'15': 'repeat(15, minmax(0, 1fr))',
			}
		}
	},
	plugins: [],
}
