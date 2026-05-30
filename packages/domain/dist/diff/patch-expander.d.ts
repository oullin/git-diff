import { type PatchLine } from "./patch-parser.js";
export interface ExpandedContext {
    oldLine: number;
    newLine: number;
    text: string;
}
export declare function applyExpansions(lines: ReadonlyArray<PatchLine>, expansions: ReadonlyArray<ExpandedContext>): PatchLine[];
//# sourceMappingURL=patch-expander.d.ts.map