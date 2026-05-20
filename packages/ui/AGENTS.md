# packages/ui — agent rules

## Framework rule

This package is **Vue 3 only**. Do not add React, react-dom, Preact, Solid, or any framework-interop layer (Veaury, web-component wrappers around `react-dom/client.createRoot`, etc.). If a library only ships a React component, we either build a Vue equivalent ourselves or use a non-React entry point if the library exposes one (e.g. `@pierre/trees/web-components`).

## Diff parsing

Parsing helpers — `parsePatchFiles`, `parseDiffFromFile`, `processFile`, `processPatch`, and the `FileDiffMetadata`/`ParsedPatch` types — come from [`@pierre/diffs`](https://www.npmjs.com/package/@pierre/diffs). They are framework-neutral and stay in [src/lib/patch.ts](src/lib/patch.ts). Don't hand-roll a parser to replace them; do keep our small fallback path for malformed input.

## Diff rendering

Rendering, syntax highlighting, comment overlays, split/unified layout, hide-whitespace, and word-level intra-line highlights all live in Vue components and Vue-flavoured TS:

- [src/components/diff/DiffBody.vue](src/components/diff/DiffBody.vue) — main view
- [src/lib/highlight.ts](src/lib/highlight.ts) — Shiki integration via Vue refs
- [src/lib/wordHi.ts](src/lib/wordHi.ts) — cheap intra-line highlight heuristic
- [src/components/diff/CommentThread.vue](src/components/diff/CommentThread.vue) — overlay

`@pierre/diffs` also exports a React `<CodeView>` component. **Do not import it.** If we ever need its virtualization or worker-based highlighting, build a Vue equivalent on top of the parser plus a Vue virtualizer (`@tanstack/vue-virtual`) and a worker-wrapped Shiki.

## File tree

The current tree is [src/components/RepoFileTree.vue](src/components/RepoFileTree.vue). If we adopt `@pierre/trees` later, use the `@pierre/trees/web-components` entry, not `@pierre/trees/react`.
