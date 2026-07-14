import { defineConfig } from 'vitest/config';
import tailwindcss from '@tailwindcss/vite';
import adapter from '@sveltejs/adapter-static';
import { sveltekit } from '@sveltejs/kit/vite';

// Connect procedure paths use the generated `<package>.<Service>` shape,
// e.g. `/auth.v1.AuthService/RequestLogin`. The proxy matches that shape
// generically (regex key), so a new service never needs to be registered
// here — there must never be a per-service list to maintain.
const backendTarget = 'http://localhost:8080';
const connectPathPattern = '^/(?:[a-z][a-z0-9_]*\\.)+v\\d+\\.[A-Za-z0-9]+Service/';

export default defineConfig({
	plugins: [
		tailwindcss(),
		sveltekit({
			compilerOptions: {
				// Force runes mode for the project, except for libraries. Can be removed in svelte 6.
				runes: ({ filename }) =>
					filename.split(/[/\\]/).includes('node_modules') ? undefined : true
			},
			adapter: adapter({
				pages: 'dist',
				assets: 'dist',
				fallback: 'index.html',
				precompress: false,
				strict: true
			})
		})
	],
	server: {
		proxy: {
			[connectPathPattern]: backendTarget,
			'/health': backendTarget
		}
	},
	test: {
		expect: { requireAssertions: true },
		environment: 'node',
		include: ['src/**/*.{test,spec}.{js,ts}']
	}
});
