import type { SelectedLineRange, SelectionSide } from '@pierre/diffs';
import type { PatchLine } from '@git-diff/domain/diff';
import type { ReviewComment } from '@git-diff/domain';
import type { LineSelectionRange, LineSide } from '@composables/useLineSelection';

// Translates between the app's comment anchor vocabulary ('left'|'right') and
// @pierre/diffs' selection vocabulary ('deletions'|'additions'), and turns a
// library selection into the (line, range?) pair that
// useReviewSession.openCommentForLine() expects. Pure + unit-tested.

/** library side → app side. Unified rows report undefined → treat as new/right. */
export function sideToLineSide(side: SelectionSide | undefined): LineSide {
	return side === 'deletions' ? 'left' : 'right';
}

/** app side → library annotation/selection side (used for getLineIndex lookups). */
export function lineSideToSelectionSide(side: LineSide | string): SelectionSide {
	return side === 'left' ? 'deletions' : 'additions';
}

function patchLineFor(side: LineSide, lineNumber: number, sectionId: string): PatchLine {
	return {
		id: `${sectionId}:${side}:${lineNumber}`,
		type: side === 'left' ? 'del' : 'add',
		text: '',
		...(side === 'left' ? { oldLine: lineNumber } : { newLine: lineNumber }),
	};
}

/**
 * A library SelectedLineRange anchors a comment at its END; a single-line
 * selection (start === end, same side) produces no range. The returned `line`
 * carries the end anchor so openCommentForLine derives the right side/number.
 */
export function selectionToCommentTarget(range: SelectedLineRange, sectionId: string): { line: PatchLine; range?: LineSelectionRange } {
	const startSide = sideToLineSide(range.side);
	const endSide = sideToLineSide(range.endSide ?? range.side);
	const line = patchLineFor(endSide, range.end, sectionId);

	if (range.start === range.end && startSide === endSide) {
		return { line };
	}

	return {
		line,
		range: { sectionId, startSide, startLine: range.start, endSide, endLine: range.end },
	};
}

/** Comments anchored to a section, sorted by side then line for stable display. */
export function commentsForSection(comments: ReviewComment[], filePath: string, sectionKind: string): ReviewComment[] {
	return comments
		.filter((c) => c.filePath === filePath && c.diffSection === sectionKind)
		.slice()
		.sort((a, b) => (a.side === b.side ? a.lineNumber - b.lineNumber : a.side === 'left' ? -1 : 1));
}
