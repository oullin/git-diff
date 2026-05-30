import type { DiffSection } from "../repo/index.js";
export type PatchLineType = "context" | "add" | "del" | "meta";
export interface PatchLine {
    id: string;
    type: PatchLineType;
    text: string;
    oldLine?: number;
    newLine?: number;
}
export type SplitRow = {
    id: string;
    kind: "meta";
    line: PatchLine;
} | {
    id: string;
    kind: "context";
    line: PatchLine;
} | {
    id: string;
    kind: "pair";
    left?: PatchLine;
    right?: PatchLine;
};
export interface HunkInfo {
    metaId: string;
    oldStart: number;
    newStart: number;
    oldEnd: number;
    newEnd: number;
    /** Last old line of the previous hunk (or 0 before the first hunk). */
    prevOldEnd: number;
    prevNewEnd: number;
    /** True for the trailing hunk; only this one can expand downward past EOF. */
    isLast: boolean;
}
export declare function parseHunkHeader(text: string): {
    oldStart: number;
    oldCount: number;
    newStart: number;
    newCount: number;
} | null;
export declare function getHunkInfos(lines: ReadonlyArray<PatchLine>): HunkInfo[];
export declare function parsePatch(section: DiffSection, hideWhitespace: boolean): PatchLine[];
export declare function splitPatch(section: DiffSection, hideWhitespace: boolean): SplitRow[];
export declare function splitPatchLines(lines: ReadonlyArray<PatchLine>): SplitRow[];
//# sourceMappingURL=patch-parser.d.ts.map