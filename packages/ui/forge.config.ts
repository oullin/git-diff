/**
 * Electron Forge configuration.
 *
 * This config lives alongside the existing electron-builder block in
 * package.json. Both target the same outputs (.dmg + .zip on macOS arm64)
 * so we can cut over once a signed/notarized Forge build has been verified
 * in CI. After the cutover, the "build" key in package.json + the
 * electron-builder devDep can be removed.
 *
 * See docs/distribution.md for the rollout plan.
 */

import { resolve } from 'node:path';

const here = import.meta.dirname;

export default {
	packagerConfig: {
		name: 'Git Diff Review',
		appBundleId: 'io.gocanto.git-diff',
		asar: true,
		icon: resolve(here, 'build/icon'),
		extraResource: [resolve(here, '../api/dist/api')],
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
};
