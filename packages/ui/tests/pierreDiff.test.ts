// Pins the @pierre/diffs adapter behaviour. The parser inside
// patch-parser.ts tries pierre first and falls back to a hand-rolled
// line scanner; if pierre breaks or the export shape changes (it's a
// beta library), this test catches it loudly instead of silently
// dropping into the fallback.

import { describe, expect, test } from 'vitest';
import { parsePatchFiles } from '@pierre/diffs';

import { getHunkInfos, parsePatch } from '@git-diff/domain/diff';
import { pierreSurfaceClasses, toPierreDisplayOptions, type PierreDisplayInput } from '@composables/usePierreDiffOptions';

const SAMPLE_PATCH = [
	'diff --git a/src/example.ts b/src/example.ts',
	'index abc1234..def5678 100644',
	'--- a/src/example.ts',
	'+++ b/src/example.ts',
	'@@ -1,4 +1,5 @@',
	' function add(a: number, b: number) {',
	'-    return a + b;',
	'+    const total = a + b;',
	'+    return total;',
	' }',
	'',
].join('\n');

describe('@pierre/diffs adapter', () => {
	test('parsePatchFiles returns at least one parsed file with hunks', () => {
		const parsed = parsePatchFiles(SAMPLE_PATCH, 'sample');

		expect(parsed).toBeDefined();
		expect(parsed.length).toBeGreaterThan(0);
		expect(parsed[0]!.files.length).toBeGreaterThan(0);
		expect(parsed[0]!.files[0]!.hunks.length).toBeGreaterThan(0);
	});

	test('parsePatch round-trips a unified diff into typed PatchLine[]', () => {
		const lines = parsePatch({ id: 'sample', kind: 'unstaged', binary: false, patch: SAMPLE_PATCH }, false);

		expect(lines[0]?.type).toBe('meta');

		expect(lines.some((l) => l.type === 'del' && l.text.includes('return a + b'))).toBe(true);
		expect(lines.some((l) => l.type === 'add' && l.text.includes('const total'))).toBe(true);
		expect(lines.some((l) => l.type === 'add' && l.text.includes('return total'))).toBe(true);

		const contextLine = lines.find((l) => l.type === 'context' && l.text.includes('function add'));

		expect(contextLine?.oldLine).toBe(1);
		expect(contextLine?.newLine).toBe(1);
	});

	test('getHunkInfos derives a single hunk for our sample', () => {
		const lines = parsePatch({ id: 'sample', kind: 'unstaged', binary: false, patch: SAMPLE_PATCH }, false);

		const infos = getHunkInfos(lines);

		expect(infos).toHaveLength(1);
		expect(infos[0]!.oldStart).toBe(1);
		expect(infos[0]!.newStart).toBe(1);
		expect(infos[0]!.isLast).toBe(true);
	});
});

describe('toPierreDisplayOptions', () => {
	const base: PierreDisplayInput = {
		viewMode: 'split',
		diffStyle: 'soft',
		wordHighlight: true,
		wrapLongLines: false,
		hideWhitespace: false,
		themeType: 'light',
	};

	test('viewMode maps to library diffStyle (layout)', () => {
		expect(toPierreDisplayOptions({ ...base, viewMode: 'split' }).diffStyle).toBe('split');
		expect(toPierreDisplayOptions({ ...base, viewMode: 'unified' }).diffStyle).toBe('unified');
	});

	test('diffStyle soft/punchy use classic indicators with a background', () => {
		for (const diffStyle of ['soft', 'punchy'] as const) {
			const opts = toPierreDisplayOptions({ ...base, diffStyle });

			expect(opts.diffIndicators).toBe('classic');
			expect(opts.disableBackground).toBe(false);
		}
	});

	test('diffStyle bar uses bar indicators and disables the background', () => {
		const opts = toPierreDisplayOptions({ ...base, diffStyle: 'bar' });

		expect(opts.diffIndicators).toBe('bars');
		expect(opts.disableBackground).toBe(true);
	});

	test('wordHighlight toggles lineDiffType between word and none', () => {
		expect(toPierreDisplayOptions({ ...base, wordHighlight: true }).lineDiffType).toBe('word');
		expect(toPierreDisplayOptions({ ...base, wordHighlight: false }).lineDiffType).toBe('none');
	});

	test('wrapLongLines toggles overflow between wrap and scroll', () => {
		expect(toPierreDisplayOptions({ ...base, wrapLongLines: true }).overflow).toBe('wrap');
		expect(toPierreDisplayOptions({ ...base, wrapLongLines: false }).overflow).toBe('scroll');
	});

	test('hideWhitespace forwards ignoreWhitespace to the diff algorithm', () => {
		expect(toPierreDisplayOptions({ ...base, hideWhitespace: true }).parseDiffOptions).toEqual({ ignoreWhitespace: true });
		expect(toPierreDisplayOptions({ ...base, hideWhitespace: false }).parseDiffOptions).toEqual({ ignoreWhitespace: false });
	});

	test('theme + themeType wire the registered Primer themes', () => {
		const opts = toPierreDisplayOptions({ ...base, themeType: 'dark' });

		expect(opts.theme).toEqual({ light: 'licht', dark: 'dunkel' });
		expect(opts.themeType).toBe('dark');
		expect(opts.disableLineNumbers).toBe(false);
		expect(opts.expansionLineCount).toBe(100);
	});
});

describe('pierreSurfaceClasses', () => {
	test('always includes the base scope; adds density + punchy modifiers', () => {
		expect(pierreSurfaceClasses({ diffStyle: 'soft', density: 'comfortable' })).toEqual(['pierre-diff']);
		expect(pierreSurfaceClasses({ diffStyle: 'soft', density: 'compact' })).toEqual(['pierre-diff', 'density-compact']);
		expect(pierreSurfaceClasses({ diffStyle: 'punchy', density: 'comfortable' })).toEqual(['pierre-diff', 'pierre-diff--punchy']);
		expect(pierreSurfaceClasses({ diffStyle: 'bar', density: 'compact' })).toEqual(['pierre-diff', 'density-compact']);
	});
});
