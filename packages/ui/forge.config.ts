/**
 * Electron Forge configuration (the supported packaging path).
 *
 * Builds the unsigned macOS arm64 .dmg + .zip and embeds the Go API binary as
 * an extraResource. Signing/notarization stay off until the Apple secrets are
 * provided (see docs/distribution.md); when present, osxSign/osxNotarize wire
 * themselves in automatically.
 *
 * Loaded by Forge via jiti, so native ESM + TypeScript (and `import.meta`) work
 * without a separate loader.
 */

import { execFileSync } from 'node:child_process';
import { join, resolve } from 'node:path';

const here = import.meta.dirname;

export default {
	packagerConfig: {
		name: 'Git Diff Review',
		appBundleId: 'io.gocanto.git-diff',
		asar: true,
		icon: resolve(here, 'build/icon'),
		extraResource: [resolve(here, '../api/dist/api')],
		// Electron downloads cache; honored by the actions/cache step in CI.
		download: {
			cacheRoot: process.env.ELECTRON_CACHE || join(here, '.cache/electron'),
		},
		// The main + preload bundles inline their full dependency graph (see
		// vite.electron.config.ts), so the packaged app needs no node_modules.
		// Disabling prune skips Forge's dependency walker, which can't resolve
		// pnpm's workspace layout; we then ignore everything except the built
		// renderer (dist/) + electron bundles (dist-electron/) and package.json.
		prune: false,
		ignore: [
			/^\/node_modules(?:$|\/)/,
			/^\/\.cache(?:$|\/)/,
			/^\/\.turbo(?:$|\/)/,
			/^\/build(?:$|\/)/,
			/^\/coverage(?:$|\/)/,
			/^\/electron(?:$|\/)/,
			/^\/out(?:$|\/)/,
			/^\/public(?:$|\/)/,
			/^\/release(?:$|\/)/,
			/^\/scripts(?:$|\/)/,
			/^\/src(?:$|\/)/,
			/^\/tests(?:$|\/)/,
			/^\/forge\.config\.ts$/,
			/^\/index\.html$/,
			/^\/tsconfig.*\.json$/,
			/^\/vite\..*config\..*$/,
			/\.DS_Store$/,
		],
		osxSign: process.env.APPLE_SIGNING_IDENTITY ? { identity: process.env.APPLE_SIGNING_IDENTITY } : undefined,
		osxNotarize:
			process.env.APPLE_API_KEY && process.env.APPLE_API_KEY_ID && process.env.APPLE_API_ISSUER
				? {
						appleApiKey: process.env.APPLE_API_KEY,
						appleApiKeyId: process.env.APPLE_API_KEY_ID,
						appleApiIssuer: process.env.APPLE_API_ISSUER,
					}
				: undefined,
	},
	rebuildConfig: {},
	makers: [
		{ name: '@electron-forge/maker-dmg', config: { format: 'ULFO' }, platforms: ['darwin'] },
		{ name: '@electron-forge/maker-zip', platforms: ['darwin'] },
	],
	publishers: [
		{
			name: '@electron-forge/publisher-github',
			config: {
				repository: { owner: 'oullin', name: 'git-diff' },
				prerelease: false,
				draft: false,
			},
		},
	],
	hooks: {
		// The old dist:mac:* scripts generated the app icon before packaging;
		// keep that guarantee by regenerating it at the start of every build.
		generateAssets: async () => {
			execFileSync('bash', ['scripts/generate-app-icon.sh'], {
				cwd: here,
				stdio: 'inherit',
			});
		},
	},
};
