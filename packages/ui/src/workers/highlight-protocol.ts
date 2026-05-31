import type { BundledLanguage } from 'shiki';

// Both the worker and highlightPool import these types so wire-shape
// changes break both sides at compile time.

export type HighlightThemeId = 'Licht' | 'Dunkel';

export interface HighlightRequest {
	kind: 'highlight';
	id: number;
	text: string;
	lang: BundledLanguage;
	theme: HighlightThemeId;
}

export interface PrewarmRequest {
	kind: 'prewarm';
	id: number;
	lang: BundledLanguage;
}

export type WorkerRequest = HighlightRequest | PrewarmRequest;

export interface HighlightResponse {
	kind: 'highlight';
	id: number;
	html: string;
}

export interface PrewarmResponse {
	kind: 'prewarm';
	id: number;
}

export interface ErrorResponse {
	kind: 'error';
	id: number;
	message: string;
}

export type WorkerResponse = HighlightResponse | PrewarmResponse | ErrorResponse;
