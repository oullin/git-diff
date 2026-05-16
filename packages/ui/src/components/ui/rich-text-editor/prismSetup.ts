// @lexical/code uses prismjs and expects `Prism` to be reachable on the global
// scope before its language component files (prism-clike, prism-javascript,
// etc.) execute. Those component files reference a bare `Prism` symbol at
// module top-level; under Vite's ESM bundling that lookup falls through to
// `globalThis.Prism`, so we seed it here before @lexical/code loads.
// @ts-expect-error -- prismjs ships no types; default export is the Prism object.
import Prism from "prismjs";

(globalThis as unknown as { Prism: unknown }).Prism = Prism;
