import { createHighlighterCore, type HighlighterCore } from "shiki/core";
import { createOnigurumaEngine } from "shiki/engine/oniguruma";
import type { BundledLanguage } from "shiki";
import { computed, shallowRef, watch } from "vue";

import lichtTheme from "../themes/licht.json" with { type: "json" };
import dunkelTheme from "../themes/dunkel.json" with { type: "json" };
import { resolvedTheme } from "@composables/useTheme";

type ThemeId = "Licht" | "Dunkel";

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

let highlighter: HighlighterCore | null = null;
const loadedLangs = new Set<string>();

/** Increments whenever a new language loads OR the color scheme flips, so Vue re-renders. */
export const highlighterRev = shallowRef(0);

const currentTheme = computed<ThemeId>(() => (resolvedTheme.value === "dark" ? "Dunkel" : "Licht"));

watch(currentTheme, () => {
    highlighterRev.value += 1;
});

async function getHighlighter(): Promise<HighlighterCore> {
    if (highlighter) {
        return highlighter;
    }

    highlighter = await createHighlighterCore({
        themes: [lichtTheme as never, dunkelTheme as never],
        langs: [],
        engine: createOnigurumaEngine(import("shiki/wasm")),
    });

    return highlighter;
}

export async function ensureLanguage(lang: BundledLanguage | null): Promise<void> {
    if (!lang || loadedLangs.has(lang)) {
        return;
    }

    const h = await getHighlighter();
    const mod = await import(/* @vite-ignore */ `shiki/langs/${lang}.mjs`);

    await h.loadLanguage(mod.default ?? mod);
    loadedLangs.add(lang);
    highlighterRev.value += 1;
}

export function highlightLine(text: string, lang: BundledLanguage | null): string {
    if (!lang || !loadedLangs.has(lang) || !highlighter) {
        return escapeHtml(text);
    }

    try {
        const tokens = highlighter.codeToTokensBase(text, {
            lang,
            theme: currentTheme.value,
            includeExplanation: false,
        });
        const row = tokens[0] ?? [];
        let html = "";

        for (const token of row) {
            const safe = escapeHtml(token.content);

            html += token.color ? `<span style="color:${token.color}">${safe}</span>` : safe;
        }

        return html;
    } catch {
        return escapeHtml(text);
    }
}

function escapeHtml(str: string): string {
    return str
        .replace(/&/g, "&amp;")
        .replace(/</g, "&lt;")
        .replace(/>/g, "&gt;")
        .replace(/"/g, "&quot;")
        .replace(/'/g, "&#39;");
}
