import type { DiffSection, ReviewComment } from '@git-diff/domain';
import type { PatchLine } from '@git-diff/domain/diff';
import type { LineSide } from '@composables/useLineSelection';

/**
 * Maps diff lines to the review comments anchored on them and resolves the
 * side/line-number a comment action should target. Pure logic, no component
 * state — unit-testable in isolation.
 */
export interface LineCommentsOptions {
	comments: () => ReviewComment[];
	filePath: () => string;
}

/** Resolve which side and line number a comment on this line should anchor to. */
export function lineSideAndNumber(line: PatchLine): { side: LineSide; lineNumber: number } | null {
	if (line.type === 'add' && line.newLine != null) {
		return { side: 'right', lineNumber: line.newLine };
	}

	if (line.type === 'del' && line.oldLine != null) {
		return { side: 'left', lineNumber: line.oldLine };
	}

	// Context lines render on both sides in split view; default to right so a
	// single-line comment lands on the new-file side.
	if (line.newLine != null) {
		return { side: 'right', lineNumber: line.newLine };
	}

	if (line.oldLine != null) {
		return { side: 'left', lineNumber: line.oldLine };
	}

	return null;
}

export function useLineComments(opts: LineCommentsOptions) {
	function commentsForLine(section: DiffSection, line: PatchLine | undefined): ReviewComment[] {
		if (!line) {
			return [];
		}

		return opts.comments().filter((c) => {
			if (c.filePath !== opts.filePath() || c.diffSection !== section.kind) {
				return false;
			}

			const lineNumberForSide = c.side === 'left' ? line.oldLine : line.newLine;

			return lineNumberForSide === c.lineNumber;
		});
	}

	return { lineSideAndNumber, commentsForLine };
}
