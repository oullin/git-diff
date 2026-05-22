// Barrel: parsing lives in patch-parser, context-splice in patch-expander.
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
