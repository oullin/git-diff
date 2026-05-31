import type { SplitRow } from '@git-diff/domain/diff';
import type { Range } from '@git-diff/domain/diff';

/** Rendered model for one cell (one side) of a split-view diff row. */
export interface CellRender {
	kind: 'ctx' | 'add' | 'rem' | 'empty';
	num: number | '';
	side: 'left' | 'right';
	html: string;
}

/**
 * Builds the left/right render model for a split-view pair row, applying
 * word-level highlighting when enabled. Pure mapping over the injected
 * highlighting functions — unit-testable without a component.
 */
export interface SplitCellRenderOptions {
	wordHighlight: () => boolean;
	highlightHtml: (text: string) => string;
	withRanges: (text: string, ranges: Range[], cls: 'wh-add' | 'wh-rem') => string;
	computeWordHi: (left: string, right: string) => { hiL: Range[]; hiR: Range[] };
}

export function useSplitCellRender(opts: SplitCellRenderOptions) {
	function renderPair(row: Extract<SplitRow, { kind: 'pair' }>): {
		left: CellRender;
		right: CellRender;
	} {
		const left = row.left;
		const right = row.right;

		let leftHtml = '';
		let rightHtml = '';

		const leftKind: CellRender['kind'] = left ? 'rem' : 'empty';
		const rightKind: CellRender['kind'] = right ? 'add' : 'empty';

		if (left && right && opts.wordHighlight()) {
			const { hiL, hiR } = opts.computeWordHi(left.text, right.text);

			leftHtml = opts.withRanges(left.text, hiL, 'wh-rem');
			rightHtml = opts.withRanges(right.text, hiR, 'wh-add');
		} else {
			leftHtml = left ? opts.highlightHtml(left.text) : '';
			rightHtml = right ? opts.highlightHtml(right.text) : '';
		}

		return {
			left: { kind: leftKind, num: left?.oldLine ?? '', side: 'left', html: leftHtml },
			right: { kind: rightKind, num: right?.newLine ?? '', side: 'right', html: rightHtml },
		};
	}

	return { renderPair };
}
