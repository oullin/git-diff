import { builtinModules } from 'node:module';
import { resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import { defineConfig } from 'vite';

const projectDir = fileURLToPath(new URL('.', import.meta.url));
const target = process.env.ELECTRON_BUILD_TARGET === 'main' ? 'main' : 'preload';
const external = ['electron', ...builtinModules, ...builtinModules.map((name) => `node:${name}`)];

const lib =
	target === 'main'
		? { entry: resolve(projectDir, 'electron/main.ts'), formats: ['es'] as const, fileName: () => 'main.js' }
		: { entry: resolve(projectDir, 'electron/preload.ts'), formats: ['cjs'] as const, fileName: () => 'preload.cjs' };

export default defineConfig({
	cacheDir: resolve(projectDir, `../../.turbo/vite/ui-electron-${target}`),
	resolve: {
		alias: [
			{
				find: /^#electron\/(.+)\.js$/u,
				replacement: resolve(projectDir, 'electron/$1.ts'),
			},
		],
	},
	build: {
		emptyOutDir: false,
		lib,
		outDir: 'dist-electron',
		rollupOptions: {
			external,
		},
	},
});
