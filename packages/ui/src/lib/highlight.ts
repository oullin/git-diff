import { createHighlighterCore, type HighlighterCore } from "shiki/core";
import { createOnigurumaEngine } from "shiki/engine/oniguruma";
import type { BundledLanguage } from "shiki";
import { computed, type ComputedRef, shallowRef, watch } from "vue";
import lichtTheme from "@themes/licht.json" with { type: "json" };
import dunkelTheme from "@themes/dunkel.json" with { type: "json" };
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

function escapeHtml(str: string): string {
  return str
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#39;");
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

/** `rev` bumps on async-token arrival and theme flip; consume it in a
 *  computed to opt into re-render on cache fills. */
export class Highlighter {
  /** Tokenized-line HTML cache, keyed by `${theme}:${lang}:${text}`. LRU bounded. */
  private readonly cache = new Map<string, string>();
  private readonly inflight = new Set<string>();
  private readonly cacheLimit: number;
  private fallbackHighlighter: HighlighterCore | null = null;
  private fallbackInit: Promise<HighlighterCore> | null = null;
  private readonly fallbackLoaded = new Set<string>();
  readonly rev = shallowRef(0);
  private readonly theme: ComputedRef<ThemeId>;

  constructor(cacheLimit = 5000) {
    this.cacheLimit = cacheLimit;
    this.theme = computed<ThemeId>(() => (resolvedTheme.value === "dark" ? "Dunkel" : "Licht"));

    watch(this.theme, () => {
      // Theme change invalidates every cached token string.
      this.cache.clear();
      this.inflight.clear();
      this.rev.value += 1;
    });
  }

  /** Async path: returns plaintext-escaped HTML now; the caller
   *  observes `rev` to re-render once the cache fills. */
  highlightLine(text: string, lang: BundledLanguage | null): string {
    if (!lang) {
      return escapeHtml(text);
    }

    if (text.length > MAX_LINE_TOKENIZE_LENGTH) {
      return escapeHtml(text);
    }

    const theme = this.theme.value;
    const key = this.keyFor(text, lang, theme);
    const cached = this.cacheGet(key);

    if (cached !== undefined) {
      return cached;
    }

    this.scheduleAsync(key, text, lang, theme);

    return escapeHtml(text);
  }

  async ensureLanguage(lang: BundledLanguage | null): Promise<void> {
    if (!lang) {
      return;
    }

    if (isHighlightPoolAvailable()) {
      await prewarmLanguage(lang);

      this.rev.value += 1;

      return;
    }

    await this.fallbackEnsureLanguage(lang);

    this.rev.value += 1;
  }

  private keyFor(text: string, lang: BundledLanguage, theme: ThemeId): string {
    return `${theme}:${lang}:${text}`;
  }

  private cacheGet(key: string): string | undefined {
    const value = this.cache.get(key);

    if (value === undefined) {
      return undefined;
    }

    // LRU: refresh recency by re-inserting.
    this.cache.delete(key);
    this.cache.set(key, value);

    return value;
  }

  private cacheSet(key: string, value: string): void {
    if (this.cache.has(key)) {
      this.cache.delete(key);
    }

    this.cache.set(key, value);

    if (this.cache.size > this.cacheLimit) {
      const oldest = this.cache.keys().next().value;

      if (oldest !== undefined) {
        this.cache.delete(oldest);
      }
    }
  }

  private scheduleAsync(key: string, text: string, lang: BundledLanguage, theme: ThemeId): void {
    if (this.inflight.has(key)) {
      return;
    }

    this.inflight.add(key);

    const tokenize = isHighlightPoolAvailable()
      ? requestHighlight(text, lang, theme)
      : this.fallbackEnsureLanguage(lang).then((h) =>
          tokenizeWithHighlighter(h, text, lang, theme),
        );

    tokenize
      .then((html) => {
        this.cacheSet(key, html);
      })
      .catch(() => {
        // Leave the cache empty so subsequent renders retry,
        // but plaintext stays visible meanwhile.
      })
      .finally(() => {
        this.inflight.delete(key);
        this.rev.value += 1;
      });
  }

  private getFallbackHighlighter(): Promise<HighlighterCore> {
    if (this.fallbackHighlighter) {
      return Promise.resolve(this.fallbackHighlighter);
    }

    if (!this.fallbackInit) {
      this.fallbackInit = createHighlighterCore({
        themes: [lichtTheme as never, dunkelTheme as never],
        langs: [],
        engine: createOnigurumaEngine(import("shiki/wasm")),
      }).then((h) => {
        this.fallbackHighlighter = h;

        return h;
      });
    }

    return this.fallbackInit;
  }

  private async fallbackEnsureLanguage(lang: BundledLanguage): Promise<HighlighterCore> {
    const h = await this.getFallbackHighlighter();

    if (this.fallbackLoaded.has(lang)) {
      return h;
    }

    const mod = await import(/* @vite-ignore */ `shiki/langs/${lang}.mjs`);

    await h.loadLanguage(mod.default ?? mod);

    this.fallbackLoaded.add(lang);

    return h;
  }
}

// Singleton so a single cache covers the whole app; tests can construct
// their own Highlighter to assert behaviour in isolation.
const sharedHighlighter = new Highlighter();

export const highlighterRev = sharedHighlighter.rev;

export function highlightLine(text: string, lang: BundledLanguage | null): string {
  return sharedHighlighter.highlightLine(text, lang);
}

export function ensureLanguage(lang: BundledLanguage | null): Promise<void> {
  return sharedHighlighter.ensureLanguage(lang);
}
