/**
 * Patch processing — barrel module.
 *
 * Pure parsing of unified diff text lives in patch-parser.ts; splicing
 * of fetched context expansions into a parsed patch lives in
 * patch-expander.ts. Importing from this file keeps existing call
 * sites working while making the SRP split visible at the module
 * boundary.
 */
export type { HunkInfo, PatchLine, PatchLineType, SplitRow } from "@lib/patch-parser";
export {
    getHunkInfos,
    parseHunkHeader,
    parsePatch,
    splitPatch,
    splitPatchLines,
} from "@lib/patch-parser";
export type { ExpandedContext } from "@lib/patch-expander";
export { applyExpansions } from "@lib/patch-expander";
