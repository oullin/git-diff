// @vitest-environment happy-dom
import { describe, expect, test } from 'vitest';

import { useSplitCellRender } from '@composables/useSplitCellRender';
import type { SplitRow } from '@git-diff/domain/diff';
import type { Range } from '@git-diff/domain/diff';

function pairRow(left: { text: string; oldLine: number } | null, right: { text: string; newLine: number } | null): Extract<SplitRow, { kind: 'pair' }> {
	return { kind: 'pair', id: 'r1', left, right } as Extract<SplitRow, { kind: 'pair' }>;
}

const deps = {
	highlightHtml: (t: string) => `<h>${t}</h>`,
	withRanges: (t: string, _r: Range[], cls: 'wh-add' | 'wh-rem') => `<${cls}>${t}</${cls}>`,
	computeWordHi: () => ({ hiL: [[0, 1]] as Range[], hiR: [[0, 1]] as Range[] }),
};

describe('useSplitCellRender.renderPair', () => {
	test('plain highlight when word highlight is off', () => {
		const { renderPair } = useSplitCellRender({ wordHighlight: () => false, ...deps });
		const out = renderPair(pairRow({ text: 'a', oldLine: 1 }, { text: 'b', newLine: 2 }));

		expect(out.left).toEqual({ kind: 'rem', num: 1, side: 'left', html: '<h>a</h>' });
		expect(out.right).toEqual({ kind: 'add', num: 2, side: 'right', html: '<h>b</h>' });
	});

	test('word ranges applied when both sides present and enabled', () => {
		const { renderPair } = useSplitCellRender({ wordHighlight: () => true, ...deps });
		const out = renderPair(pairRow({ text: 'a', oldLine: 1 }, { text: 'b', newLine: 2 }));

		expect(out.left.html).toBe('<wh-rem>a</wh-rem>');
		expect(out.right.html).toBe('<wh-add>b</wh-add>');
	});

	test('missing side renders as empty with blank number', () => {
		const { renderPair } = useSplitCellRender({ wordHighlight: () => true, ...deps });
		const out = renderPair(pairRow(null, { text: 'b', newLine: 2 }));

		expect(out.left).toEqual({ kind: 'empty', num: '', side: 'left', html: '' });
		expect(out.right.kind).toBe('add');
	});
});
