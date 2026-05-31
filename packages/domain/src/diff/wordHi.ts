// Compute char-range highlights for a paired removed/added line.
// Strategy: trim a common prefix and suffix; whatever remains in the middle is
// the changed span. Cheap, deterministic, no diff library needed.

export type Range = [number, number];

// Past this length the prefix/suffix walk starts to dominate render cost on
// pathological lines (minified bundles, generated code). Skipping the highlight
// is preferable to blocking the main thread.
export const MAX_LINE_WORD_DIFF_LENGTH = 2000;

export function computeWordHi(left: string, right: string): { hiL: Range[]; hiR: Range[] } {
	if (left === right || !left || !right) {
		return { hiL: [], hiR: [] };
	}

	if (left.length > MAX_LINE_WORD_DIFF_LENGTH || right.length > MAX_LINE_WORD_DIFF_LENGTH) {
		return { hiL: [], hiR: [] };
	}

	const max = Math.min(left.length, right.length);

	let prefix = 0;

	while (prefix < max && left[prefix] === right[prefix]) {
		prefix++;
	}

	let suffix = 0;

	while (suffix < max - prefix && left[left.length - 1 - suffix] === right[right.length - 1 - suffix]) {
		suffix++;
	}

	const hiL: Range[] = [[prefix, left.length - suffix]];
	const hiR: Range[] = [[prefix, right.length - suffix]];

	// If either side has an empty range, both sides are pure insert/delete — no
	// useful highlight to show.
	if (hiL[0][1] <= hiL[0][0] && hiR[0][1] <= hiR[0][0]) {
		return { hiL: [], hiR: [] };
	}

	return { hiL, hiR };
}
