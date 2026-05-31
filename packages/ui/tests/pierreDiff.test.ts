// Pins the @pierre/diffs adapter behaviour. The parser inside
// patch-parser.ts tries pierre first and falls back to a hand-rolled
// line scanner; if pierre breaks or the export shape changes (it's a
// beta library), this test catches it loudly instead of silently
// dropping into the fallback.

import { describe, expect, test } from 'vitest';
import { parsePatchFiles } from '@pierre/diffs';

import { getHunkInfos, parsePatch } from '@git-diff/domain/diff';

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
