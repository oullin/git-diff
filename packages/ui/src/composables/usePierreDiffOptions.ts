import type { BaseDiffOptions, ThemeTypes } from '@pierre/diffs';
import type { DiffHunkStyle, DiffViewMode } from '@git-diff/domain';
import { PIERRE_THEMES } from '@lib/pierreTheme';

// Pure mapping from the app's user-facing Tweaks/props onto the declarative
// subset of @pierre/diffs FileDiffOptions. Interaction/annotation callbacks
// (comments, gutter utility, expansion) are layered on by the renderer
// component — this module owns only the display knobs so they can be
// unit-tested without mounting the renderer.
//
// Naming note: the app's `viewMode` ('split'|'unified') maps to the library's
// `diffStyle`, while the app's `diffStyle` ('soft'|'punchy'|'bar') maps to the
// library's `diffIndicators` + `disableBackground`.

/** Matches the current useContextExpansion CONTEXT_EXPANSION_LINE_COUNT feel. */
export const PIERRE_EXPANSION_LINE_COUNT = 100;

export interface PierreDisplayInput {
	viewMode: DiffViewMode;
	diffStyle: DiffHunkStyle;
	wordHighlight: boolean;
	wrapLongLines: boolean;
	hideWhitespace: boolean;
	themeType: ThemeTypes;
}

/** The display-only slice of FileDiffOptions this app controls. */
export type PierreDisplayOptions = Required<
	Pick<BaseDiffOptions, 'diffStyle' | 'diffIndicators' | 'disableBackground' | 'lineDiffType' | 'overflow' | 'disableLineNumbers' | 'themeType' | 'expansionLineCount' | 'parseDiffOptions'>
> & {
	theme: typeof PIERRE_THEMES;
};

export function toPierreDisplayOptions(input: PierreDisplayInput): PierreDisplayOptions {
	const bar = input.diffStyle === 'bar';

	return {
		// app viewMode → library layout
		diffStyle: input.viewMode === 'split' ? 'split' : 'unified',

		// app diffStyle → change indicators. 'bar' = edge stripe, no fill;
		// 'soft'/'punchy' = classic +/- with tinted fill (punchy bumps the fill
		// strength via a CSS class in the renderer, no library knob for it).
		diffIndicators: bar ? 'bars' : 'classic',
		disableBackground: bar,

		lineDiffType: input.wordHighlight ? 'word' : 'none',
		overflow: input.wrapLongLines ? 'wrap' : 'scroll',
		disableLineNumbers: false,

		theme: PIERRE_THEMES,
		themeType: input.themeType,

		expansionLineCount: PIERRE_EXPANSION_LINE_COUNT,
		parseDiffOptions: { ignoreWhitespace: input.hideWhitespace },
	};
}

/** Wrapper classes that drive the CSS bridge in style.css (.pierre-diff). */
export function pierreSurfaceClasses(input: Pick<PierreDisplayInput, 'diffStyle'> & { density: 'comfortable' | 'compact' }): string[] {
	const classes = ['pierre-diff'];

	if (input.density === 'compact') {
		classes.push('density-compact');
	}

	if (input.diffStyle === 'punchy') {
		classes.push('pierre-diff--punchy');
	}

	return classes;
}
