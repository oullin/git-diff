# packages/ui — agent rules

## Framework rule

This package is **Vue 3 only**. Do not add React, react-dom, Preact, Solid, or any framework-interop layer (Veaury, web-component wrappers around `react-dom/client.createRoot`, etc.). If a library only ships a React component, we either build a Vue equivalent ourselves or use a non-React entry point if the library exposes one (e.g. `@pierre/trees/web-components`).

## Diff parsing

Parsing helpers — `parsePatchFiles`, `parseDiffFromFile`, `processFile`, `processPatch`, and the `FileDiffMetadata`/`ParsedPatch` types — come from [`@pierre/diffs`](https://www.npmjs.com/package/@pierre/diffs). They are framework-neutral and stay in [src/lib/patch.ts](src/lib/patch.ts). Don't hand-roll a parser to replace them; do keep our small fallback path for malformed input.

## Diff rendering

Diffs are rendered by the **`@pierre/diffs`** (diffs.com) vanilla `FileDiff` engine — it owns syntax highlighting (Shiki), split/unified layout, change indicators, line wrapping, word/char intra-line highlighting, and native context expansion. We wrap it; we do not re-implement rendering.

- [src/components/diff/PierreDiffBody.vue](src/components/diff/PierreDiffBody.vue) — Vue wrapper: one `FileDiff` per diff section, options, theme, add-comment affordance, outdated detection
- [src/composables/usePierreFileDiff.ts](src/composables/usePierreFileDiff.ts) — resolves a `FileDiffMetadata` per section (full-file mode via `parseDiffFromFile`, patch fallback via `processFile`)
- [src/composables/usePierreDiffOptions.ts](src/composables/usePierreDiffOptions.ts) — Tweaks/props → `FileDiffOptions`
- [src/composables/usePierreComments.ts](src/composables/usePierreComments.ts) — comment side translation + selection→target
- [src/lib/pierreTheme.ts](src/lib/pierreTheme.ts) — registers the Primer Licht/Dunkel Shiki themes
- [src/components/diff/CommentThread.vue](src/components/diff/CommentThread.vue) — comment UI, rendered in light DOM beneath each section (FileDiff renders into a Shadow DOM, so rich Vue comment UI can't be embedded inline)

`FileDiff` renders into a **Shadow DOM**; theme it via the `--diffs-*` override vars in the `.pierre-diff` scope in `src/style.css`. `@pierre/diffs` also exports React components under `@pierre/diffs/react` and a `<diffs-container>` web component — **do not import either**; use the vanilla `FileDiff` class directly.

## File tree

The current tree is [src/components/RepoFileTree.vue](src/components/RepoFileTree.vue). If we adopt `@pierre/trees` later, use the `@pierre/trees/web-components` entry, not `@pierre/trees/react`.
