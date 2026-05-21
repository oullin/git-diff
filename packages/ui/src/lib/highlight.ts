import { createHighlighterCore, type HighlighterCore } from "shiki/core";
import { createOnigurumaEngine } from "shiki/engine/oniguruma";
import type { BundledLanguage } from "shiki";
import { computed, shallowRef, watch } from "vue";

import lichtTheme from "../themes/licht.json" with { type: "json" };
import dunkelTheme from "../themes/dunkel.json" with { type: "json" };
import { resolvedTheme } from "@composables/useTheme";
import {
    isHighlightPoolAvailable,
    MAX_LINE_TOKENIZE_LENGTH,
    prewarmLanguage,
    requestHighlight,
    type ThemeId,
} from "@lib/highlightPool";

const EXTENSION_LANG: Record<string, BundledLanguage> = {
    ts: "ts",
    tsx: "tsx",
    js: "js",
    jsx: "jsx",
    mjs: "js",
    cjs: "js",
    vue: "vue",
    json: "json",
    jsonc: "jsonc",
    md: "md",
    mdx: "mdx",
    css: "css",
    scss: "scss",
    html: "html",
    go: "go",
    py: "python",
    rb: "ruby",
    rs: "rust",
    java: "java",
    kt: "kotlin",
    swift: "swift",
    sh: "shell",
    bash: "shell",
    zsh: "shell",
    yml: "yaml",
    yaml: "yaml",
    toml: "toml",
    xml: "xml",
    sql: "sql",
    c: "c",
    h: "c",
    cpp: "cpp",
    hpp: "cpp",
    cs: "csharp",
    php: "php",
    lua: "lua",
    dockerfile: "docker",
};

export function languageFor(path: string): BundledLanguage | null {
    const base = path.toLowerCase().split("/").pop() ?? "";

    if (base === "dockerfile" || base.endsWith(".dockerfile")) {
        return "docker";
    }

    const dot = base.lastIndexOf(".");

    if (dot < 0) {
        return null;
    }

    return EXTENSION_LANG[base.slice(dot + 1)] ?? null;
}

/** Increments whenever async tokens arrive OR the color scheme flips, so Vue re-renders. */
export const highlighterRev = shallowRef(0);

const currentTheme = computed<ThemeId>(() => (resolvedTheme.value === "dark" ? "Dunkel" : "Licht"));

watch(currentTheme, () => {
    // Theme change invalidates every cached token string. Reset the cache to
    // free memory and force re-tokenization in the new theme.
    cache.clear();
    inflight.clear();
    highlighterRev.value += 1;
});

// LRU cache for tokenized line HTML. Keyed by `${theme}:${lang}:${text}`. A few
// thousand entries cover the visible viewport plus recent scrollback; older
// entries get evicted automatically when the file changes.
const CACHE_LIMIT = 5000;
const cache = new Map<string, string>();
const inflight = new Set<string>();

function cacheGet(key: string): string | undefined {
    const value = cache.get(key);

    if (value === undefined) {
        return undefined;
    }

    // LRU: refresh recency by re-inserting.
    cache.delete(key);
    cache.set(key, value);

    return value;
}

function cacheSet(key: string, value: string): void {
    if (cache.has(key)) {
        cache.delete(key);
    }

    cache.set(key, value);

    if (cache.size > CACHE_LIMIT) {
        const oldest = cache.keys().next().value;

        if (oldest !== undefined) {
            cache.delete(oldest);
        }
    }
}

function keyFor(text: string, lang: BundledLanguage, theme: ThemeId): string {
    return `${theme}:${lang}:${text}`;
}

function escapeHtml(str: string): string {
    return str
        .replace(/&/g, "&amp;")
        .replace(/</g, "&lt;")
        .replace(/>/g, "&gt;")
        .replace(/"/g, "&quot;")
        .replace(/'/g, "&#39;");
}

// Main-thread highlighter. Used only when the worker pool is unavailable
// (vitest happy-dom, Electron contexts without Web Worker support, etc.) so
// tests and degraded environments still render highlighted output.
let fallbackHighlighter: HighlighterCore | null = null;
let fallbackInitPromise: Promise<HighlighterCore> | null = null;
const fallbackLoadedLangs = new Set<string>();

function getFallbackHighlighter(): Promise<HighlighterCore> {
    if (fallbackHighlighter) {
        return Promise.resolve(fallbackHighlighter);
    }

    if (!fallbackInitPromise) {
        fallbackInitPromise = createHighlighterCore({
            themes: [lichtTheme as never, dunkelTheme as never],
            langs: [],
            engine: createOnigurumaEngine(import("shiki/wasm")),
        }).then((h) => {
            fallbackHighlighter = h;

            return h;
        });
    }

    return fallbackInitPromise;
}

async function fallbackEnsureLanguage(lang: BundledLanguage): Promise<HighlighterCore> {
    const h = await getFallbackHighlighter();

    if (fallbackLoadedLangs.has(lang)) {
        return h;
    }

    const mod = await import(/* @vite-ignore */ `shiki/langs/${lang}.mjs`);

    await h.loadLanguage(mod.default ?? mod);
    fallbackLoadedLangs.add(lang);

    return h;
}

function tokenizeWithHighlighter(
    h: HighlighterCore,
    text: string,
    lang: BundledLanguage,
    theme: ThemeId,
): string {
    const tokens = h.codeToTokensBase(text, {
        lang,
        theme,
        includeExplanation: false,
    });
    const row = tokens[0] ?? [];
    let html = "";

    for (const token of row) {
        const safe = escapeHtml(token.content);

        html += token.color ? `<span style="color:${token.color}">${safe}</span>` : safe;
    }

    return html;
}

export async function ensureLanguage(lang: BundledLanguage | null): Promise<void> {
    if (!lang) {
        return;
    }

    if (isHighlightPoolAvailable()) {
        await prewarmLanguage(lang);
        highlighterRev.value += 1;

        return;
    }

    await fallbackEnsureLanguage(lang);
    highlighterRev.value += 1;
}

function scheduleAsync(key: string, text: string, lang: BundledLanguage, theme: ThemeId): void {
    if (inflight.has(key)) {
        return;
    }

    inflight.add(key);

    const tokenize = isHighlightPoolAvailable()
        ? requestHighlight(text, lang, theme)
        : fallbackEnsureLanguage(lang).then((h) => tokenizeWithHighlighter(h, text, lang, theme));

    tokenize
        .then((html) => {
            cacheSet(key, html);
        })
        .catch(() => {
            // Leave the cache empty so subsequent renders try again. Plaintext
            // remains visible in the meantime — that's better than retrying in
            // a hot loop on a permanent failure.
        })
        .finally(() => {
            inflight.delete(key);
            highlighterRev.value += 1;
        });
}

export function highlightLine(text: string, lang: BundledLanguage | null): string {
    if (!lang) {
        return escapeHtml(text);
    }

    if (text.length > MAX_LINE_TOKENIZE_LENGTH) {
        return escapeHtml(text);
    }

    const theme = currentTheme.value;
    const key = keyFor(text, lang, theme);
    const cached = cacheGet(key);

    if (cached !== undefined) {
        return cached;
    }

    scheduleAsync(key, text, lang, theme);

    return escapeHtml(text);
}
