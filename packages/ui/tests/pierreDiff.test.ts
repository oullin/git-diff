// Pins the @pierre/diffs adapter behaviour. The parser inside
// patch-parser.ts tries pierre first and falls back to a hand-rolled
// line scanner; if pierre breaks or the export shape changes (it's a
// beta library), this test catches it loudly instead of silently
// dropping into the fallback.

import { describe, expect, test } from 'vitest';
import { parsePatchFiles } from '@pierre/diffs';

import { getHunkInfos, parsePatch } from '@git-diff/domain/diff';
import { pierreSurfaceClasses, toPierreDisplayOptions, type PierreDisplayInput } from '@composables/usePierreDiffOptions';
import { resolveSectionRefs, type DiffSourceContext } from '@composables/usePierreFileDiff';
import { commentsForSection, lineSideToSelectionSide, selectionToCommentTarget, sideToLineSide } from '@composables/usePierreComments';
import type { ReviewComment } from '@git-diff/domain';

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

		expect(opts.theme).toEqual({ light: 'Licht', dark: 'Dunkel' });
		expect(opts.themeType).toBe('dark');
		expect(opts.disableLineNumbers).toBe(false);
		expect(opts.expansionLineCount).toBe(100);
	});
});

describe('resolveSectionRefs', () => {
	const working: DiffSourceContext = { repoRoot: '/repo', ignoreWhitespace: false };

	test('working tree sections map to HEAD/index/working-tree refs', () => {
		expect(resolveSectionRefs('staged', working)).toEqual({ oldRef: 'HEAD', newRef: ':0' });
		expect(resolveSectionRefs('unstaged', working)).toEqual({ oldRef: ':0', newRef: '' });
		expect(resolveSectionRefs('untracked', working)).toEqual({ oldRef: null, newRef: '' });
	});

	test('plain commit diffs against its first parent', () => {
		const ctx: DiffSourceContext = { repoRoot: '/repo', commitRef: 'abc123', ignoreWhitespace: false };

		expect(resolveSectionRefs('commit', ctx)).toEqual({ oldRef: 'abc123^', newRef: 'abc123' });
	});

	test('pull request diffs head against the base ref', () => {
		const ctx: DiffSourceContext = { repoRoot: '/repo', commitRef: 'head456', baseRef: 'main', ignoreWhitespace: false };

		expect(resolveSectionRefs('commit', ctx)).toEqual({ oldRef: 'main', newRef: 'head456' });
	});
});

describe('comment side translation', () => {
	test('library side ↔ app side', () => {
		expect(sideToLineSide('deletions')).toBe('left');
		expect(sideToLineSide('additions')).toBe('right');
		expect(sideToLineSide(undefined)).toBe('right');
		expect(lineSideToSelectionSide('left')).toBe('deletions');
		expect(lineSideToSelectionSide('right')).toBe('additions');
	});

	test('single-line selection → line anchor, no range', () => {
		const { line, range } = selectionToCommentTarget({ start: 42, side: 'additions', end: 42, endSide: 'additions' }, 'sec');

		expect(range).toBeUndefined();
		expect(line.newLine).toBe(42);
		expect(line.oldLine).toBeUndefined();
		expect(line.type).toBe('add');
	});

	test('deletion-side single line anchors on the old line', () => {
		const { line } = selectionToCommentTarget({ start: 7, side: 'deletions', end: 7, endSide: 'deletions' }, 'sec');

		expect(line.oldLine).toBe(7);
		expect(line.newLine).toBeUndefined();
		expect(line.type).toBe('del');
	});

	test('multi-line selection → range anchored at the end', () => {
		const { line, range } = selectionToCommentTarget({ start: 10, side: 'additions', end: 15, endSide: 'additions' }, 'sec');

		expect(line.newLine).toBe(15);
		expect(range).toEqual({ sectionId: 'sec', startSide: 'right', startLine: 10, endSide: 'right', endLine: 15 });
	});
});

describe('commentsForSection', () => {
	const make = (id: number, side: 'left' | 'right', lineNumber: number, diffSection = 'unstaged'): ReviewComment =>
		({ id, filePath: 'a.ts', diffSection, side, lineNumber, resolved: false, bodyHtml: '', authorLabel: 'x', createdAt: '' }) as unknown as ReviewComment;

	test('filters by file + section and sorts left-before-right then by line', () => {
		const comments = [make(1, 'right', 20), make(2, 'left', 5), make(3, 'right', 8), make(4, 'left', 30, 'staged'), { ...make(5, 'right', 9), filePath: 'b.ts' } as ReviewComment];

		const result = commentsForSection(comments, 'a.ts', 'unstaged');

		expect(result.map((c) => c.id)).toEqual([2, 3, 1]);
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
