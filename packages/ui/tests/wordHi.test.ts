// @vitest-environment happy-dom
import { describe, expect, test } from 'vitest';

import { computeWordHi, MAX_LINE_WORD_DIFF_LENGTH } from '@git-diff/domain/diff';

describe('computeWordHi', () => {
	test('returns ranges for diverging middles', () => {
		const { hiL, hiR } = computeWordHi('foo bar baz', 'foo qux baz');

		expect(hiL).toEqual([[4, 7]]);
		expect(hiR).toEqual([[4, 7]]);
	});

	test('returns empty ranges for identical inputs', () => {
		expect(computeWordHi('foo', 'foo')).toEqual({ hiL: [], hiR: [] });
	});

	test('returns empty ranges when either side exceeds the long-line guard', () => {
		const long = 'a'.repeat(MAX_LINE_WORD_DIFF_LENGTH + 1);

		expect(computeWordHi(long, long.replace('a', 'b'))).toEqual({ hiL: [], hiR: [] });
		expect(computeWordHi('short', long)).toEqual({ hiL: [], hiR: [] });
	});
});
