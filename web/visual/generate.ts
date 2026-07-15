// make visual: renders every SvelteKit page from hand-authored fixtures (no
// live backend, no database) and writes deterministic static HTML into
// docs/ui-snapshots/ so a `git diff` on that directory shows UI changes.
// See docs/prds/006-visual-snapshots.md.
import { chromium } from 'playwright';
import { execSync } from 'node:child_process';
import { createServer, type Server } from 'node:http';
import { readFile } from 'node:fs/promises';
import { existsSync, mkdirSync, readFileSync, readdirSync, writeFileSync } from 'node:fs';
import { dirname, extname, join } from 'node:path';

// Mirrors the pattern in vite.config.ts that proxies Connect procedure calls
// to the Go backend. Here it identifies which requests to intercept with
// fixture data instead of letting them reach a (non-existent) backend.
const CONNECT_PATH_PATTERN = /^\/(?:[a-z][a-z0-9_]*\.)+v\d+\.[A-Za-z0-9]+Service\//;

const WEB_DIR = join(import.meta.dir, '..');
const ROUTES_DIR = join(WEB_DIR, 'src/routes');
const DIST_DIR = join(WEB_DIR, 'dist');
const FIXTURES_DIR = join(import.meta.dir, 'fixtures');
const OUTPUT_DIR = join(WEB_DIR, '../docs/ui-snapshots');
const PORT = 4173;

interface Fixture {
	localStorage?: Record<string, string>;
	rpc?: Record<string, unknown>;
	// Some pages only reach their real state through client-side navigation
	// (e.g. /login/otp needs an email carried over in memory from /login — a
	// fresh deep link to it just redirects back). When set, the snapshot
	// starts at `visit`, fills `fill` (selector -> value), clicks `submit`,
	// and waits for the app's own navigation to land on the route's URL,
	// instead of loading the route's URL directly.
	visit?: string;
	fill?: Record<string, string>;
	submit?: string;
}

interface Route {
	urlPath: string;
	fixtureFile: string;
	outputFile: string;
}

function toRoute(urlPath: string): Route {
	const relative = urlPath === '/' ? 'index' : urlPath.slice(1);
	return {
		urlPath,
		fixtureFile: join(FIXTURES_DIR, `${relative}.json`),
		outputFile: join(OUTPUT_DIR, `${relative}.html`)
	};
}

// Walks src/routes for +page.svelte files, turning the directory tree into
// URL paths. Route groups `(name)` contribute a directory but not a URL
// segment. Dynamic segments `[param]` have no fixture data to render with,
// so they are skipped rather than guessed at (see PRD Out of Scope).
function discoverRoutes(dir: string, urlSegments: string[]): Route[] {
	const entries = readdirSync(dir, { withFileTypes: true });
	const routes: Route[] = [];

	if (entries.some((entry) => entry.isFile() && entry.name === '+page.svelte')) {
		const urlPath = urlSegments.length === 0 ? '/' : `/${urlSegments.join('/')}`;
		routes.push(toRoute(urlPath));
	}

	for (const entry of entries) {
		if (!entry.isDirectory()) continue;
		if (entry.name.includes('[')) {
			console.warn(`make visual: skipping dynamic route segment "${entry.name}" (out of scope)`);
			continue;
		}
		const isGroup = entry.name.startsWith('(') && entry.name.endsWith(')');
		const nextSegments = isGroup ? urlSegments : [...urlSegments, entry.name];
		routes.push(...discoverRoutes(join(dir, entry.name), nextSegments));
	}

	return routes;
}

const MIME_TYPES: Record<string, string> = {
	'.html': 'text/html',
	'.js': 'text/javascript',
	'.css': 'text/css',
	'.svg': 'image/svg+xml',
	'.json': 'application/json',
	'.png': 'image/png',
	'.ico': 'image/x-icon',
	'.woff2': 'font/woff2'
};

// Serves the production build with SvelteKit's SPA fallback semantics:
// every path that isn't a real asset file resolves to index.html.
function serveDist(root: string, port: number): Promise<Server> {
	const server = createServer((req, res) => {
		void (async () => {
			const url = new URL(req.url ?? '/', 'http://localhost');
			const requested = join(root, decodeURIComponent(url.pathname));
			const candidates = requested.endsWith('/') ? [join(requested, 'index.html')] : [requested];
			for (const candidate of candidates) {
				try {
					const body = await readFile(candidate);
					res.writeHead(200, {
						'Content-Type': MIME_TYPES[extname(candidate)] ?? 'application/octet-stream'
					});
					res.end(body);
					return;
				} catch {
					// fall through to the next candidate, then the SPA fallback
				}
			}
			const body = await readFile(join(root, 'index.html'));
			res.writeHead(200, { 'Content-Type': 'text/html' });
			res.end(body);
		})();
	});
	return new Promise((resolve) => server.listen(port, () => resolve(server)));
}

// page.content() captures the rendered DOM, but its <link rel="stylesheet">
// and <script>/modulepreload tags point at hashed paths served by the dev
// preview server — opened as a plain file, the snapshot would show unstyled
// text with dead links. Snapshots exist to be opened directly (git viewer,
// double-click, a future diff tool), so CSS is inlined and the now-useless
// hydration script/preloads are dropped.
function inlineAssets(html: string, distRoot: string): string {
	const withInlinedCss = html.replace(/<link\b[^>]*>/g, (tag) => {
		if (/rel="stylesheet"/.test(tag)) {
			const hrefMatch = /href="([^"]+)"/.exec(tag);
			if (hrefMatch) {
				const css = readFileSync(join(distRoot, hrefMatch[1]), 'utf8');
				return `<style>${css}</style>`;
			}
		}
		return /rel="modulepreload"/.test(tag) ? '' : tag;
	});
	return withInlinedCss.replace(/<script[^>]*>[\s\S]*?<\/script>/g, '');
}

// Two builds of identical source do not produce byte-identical output: Vite
// content-hashes chunk filenames, and SvelteKit inlines a randomly generated
// `__sveltekit_<id>` global per build (to avoid collisions when multiple
// SvelteKit apps share a page). Both end up in the captured HTML, so they are
// normalized away here — otherwise every snapshot would "change" on every
// run even without a source edit.
function normalize(html: string): string {
	const IMMUTABLE_DIR = '(_app\\/immutable\\/(?:entry|chunks|nodes|assets)\\/)';
	return (
		html
			.replace(/__sveltekit_[a-z0-9]+/g, '__sveltekit_app')
			// named-plus-hash files, e.g. entry/start.vcdgWkST.js -> entry/start.js
			.replace(
				new RegExp(`${IMMUTABLE_DIR}([^/".]+)\\.[A-Za-z0-9_-]{6,14}\\.(js|css)`, 'g'),
				'$1$2.$3'
			)
			// unnamed hash-only chunks, e.g. chunks/BnmnixSp.js -> chunks/HASH.js
			.replace(new RegExp(`${IMMUTABLE_DIR}[A-Za-z0-9_-]{6,14}\\.(js|css)`, 'g'), '$1HASH.$2')
	);
}

async function snapshotRoute(browser: import('playwright').Browser, route: Route): Promise<void> {
	const fixture: Fixture = JSON.parse(await readFile(route.fixtureFile, 'utf8'));
	const context = await browser.newContext();
	const page = await context.newPage();

	await page.route(
		(url) => CONNECT_PATH_PATTERN.test(url.pathname),
		async (routeHandle) => {
			const pathname = new URL(routeHandle.request().url()).pathname;
			const body = fixture.rpc?.[pathname];
			if (body === undefined) {
				console.warn(`make visual: no fixture rpc response for ${pathname} on ${route.urlPath}`);
				await routeHandle.fulfill({ status: 501, body: 'no fixture response' });
				return;
			}
			await routeHandle.fulfill({
				status: 200,
				contentType: 'application/json',
				body: JSON.stringify(body)
			});
		}
	);

	if (fixture.localStorage) {
		await context.addInitScript((seed: Record<string, string>) => {
			for (const [key, value] of Object.entries(seed)) {
				window.localStorage.setItem(key, value);
			}
		}, fixture.localStorage);
	}

	await page.goto(`http://localhost:${PORT}${fixture.visit ?? route.urlPath}`, {
		waitUntil: 'networkidle'
	});

	if (fixture.fill) {
		for (const [selector, value] of Object.entries(fixture.fill)) {
			await page.fill(selector, value);
		}
	}
	if (fixture.submit) {
		await page.click(fixture.submit);
		await page.waitForURL(`**${route.urlPath}`);
		await page.waitForLoadState('networkidle');
	}

	const html = normalize(inlineAssets(await page.content(), DIST_DIR));

	mkdirSync(dirname(route.outputFile), { recursive: true });
	writeFileSync(route.outputFile, `${html}\n`);

	await context.close();
}

async function main(): Promise<void> {
	const routes = discoverRoutes(ROUTES_DIR, []);
	const missing = routes.filter((route) => !existsSync(route.fixtureFile));
	if (missing.length > 0) {
		console.error('make visual: missing fixture file for route(s):');
		for (const route of missing) {
			console.error(`  ${route.urlPath} -> ${route.fixtureFile}`);
		}
		process.exit(1);
	}

	console.log('==> building web app');
	execSync('bun run build', { cwd: WEB_DIR, stdio: 'inherit' });

	const server = await serveDist(DIST_DIR, PORT);
	try {
		const browser = await chromium.launch();
		try {
			for (const route of routes) {
				await snapshotRoute(browser, route);
			}
		} finally {
			await browser.close();
		}
	} finally {
		server.close();
	}

	console.log(`==> wrote ${routes.length} snapshot(s) to docs/ui-snapshots/`);
}

main().catch((err) => {
	console.error(err);
	process.exit(1);
});
