/// <reference lib="webworker" />

import { createHighlighterCore, type HighlighterCore } from 'shiki/core';
import { createOnigurumaEngine } from 'shiki/engine/oniguruma';
import type { BundledLanguage } from 'shiki';
import lichtTheme from '@themes/licht.json' with { type: 'json' };
import dunkelTheme from '@themes/dunkel.json' with { type: 'json' };

import type { ErrorResponse, HighlightResponse, HighlightThemeId, PrewarmResponse, WorkerRequest } from '@workers/highlight-protocol';

export type { ErrorResponse, HighlightRequest, HighlightResponse, HighlightThemeId, PrewarmRequest, PrewarmResponse, WorkerRequest, WorkerResponse } from '@workers/highlight-protocol';

let highlighter: HighlighterCore | null = null;

const loadedLangs = new Set<string>();

let initPromise: Promise<HighlighterCore> | null = null;

function getHighlighter(): Promise<HighlighterCore> {
	if (highlighter) {
		return Promise.resolve(highlighter);
	}

	if (!initPromise) {
		initPromise = createHighlighterCore({
			themes: [lichtTheme as never, dunkelTheme as never],
			langs: [],
			engine: createOnigurumaEngine(import('shiki/wasm')),
		}).then((h) => {
			highlighter = h;

			return h;
		});
	}

	return initPromise;
}

async function ensureLanguage(lang: BundledLanguage): Promise<HighlighterCore> {
	const h = await getHighlighter();

	if (loadedLangs.has(lang)) {
		return h;
	}

	const mod = await import(/* @vite-ignore */ `shiki/langs/${lang}.mjs`);

	await h.loadLanguage(mod.default ?? mod);

	loadedLangs.add(lang);

	return h;
}

function escapeHtml(value: string): string {
	return value.replace(/[&<>"']/g, (ch) => {
		switch (ch) {
			case '&':
				return '&amp;';

			case '<':
				return '&lt;';

			case '>':
				return '&gt;';

			case '"':
				return '&quot;';

			default:
				return '&#39;';
		}
	});
}

async function tokenize(text: string, lang: BundledLanguage, theme: HighlightThemeId): Promise<string> {
	const h = await ensureLanguage(lang);

	const tokens = h.codeToTokensBase(text, {
		lang,
		theme,
		includeExplanation: false,
	});

	const row = tokens[0] ?? [];

	let html = '';

	for (const token of row) {
		const safe = escapeHtml(token.content);

		html += token.color ? `<span style="color:${token.color}">${safe}</span>` : safe;
	}

	return html;
}

const ctx = self as unknown as DedicatedWorkerGlobalScope;

ctx.addEventListener('message', async (event: MessageEvent<WorkerRequest>) => {
	const message = event.data;

	try {
		if (message.kind === 'prewarm') {
			await ensureLanguage(message.lang);

			ctx.postMessage({ kind: 'prewarm', id: message.id } satisfies PrewarmResponse);

			return;
		}

		const html = await tokenize(message.text, message.lang, message.theme);

		ctx.postMessage({ kind: 'highlight', id: message.id, html } satisfies HighlightResponse);
	} catch (err) {
		ctx.postMessage({
			kind: 'error',
			id: message.id,
			message: err instanceof Error ? err.message : String(err),
		} satisfies ErrorResponse);
	}
});
