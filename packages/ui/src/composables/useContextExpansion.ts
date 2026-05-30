import { reactive } from "vue";

import type { ExpandedContext, HunkInfo } from "@git-diff/domain/diff";

export const CONTEXT_EXPANSION_LINE_COUNT = 100;

export interface ExpansionRequest {
    sectionId: string;
    repoRoot: string;
    filePath: string;
    ref?: string;
}

interface SectionState {
    expansions: ExpandedContext[];
    downwardEof: boolean;
    inflight: Set<string>;
}

async function fetchRange(
    req: ExpansionRequest,
    startLine: number,
    endLine: number,
): Promise<{ lines: string[]; eof: boolean }> {
    if (startLine > endLine) {
        return { lines: [], eof: false };
    }

    const result = await window.diffApp.readRepositoryFileRange({
        root: req.repoRoot,
        path: req.filePath,
        ref: req.ref,
        startLine,
        endLine,
    });

    return { lines: result.lines, eof: result.eof };
}

/** Per-instance store — two DiffBody instances no longer corrupt each
 *  other's expansion state. */
export function useContextExpansion() {
    const sections = reactive(new Map<string, SectionState>());

    function ensureState(sectionId: string): SectionState {
        let state = sections.get(sectionId);

        if (!state) {
            state = reactive({
                expansions: [],
                downwardEof: false,
                inflight: new Set<string>(),
            });
            sections.set(sectionId, state);
        }

        return state;
    }

    function getExpansions(sectionId: string): ExpandedContext[] {
        return sections.get(sectionId)?.expansions ?? [];
    }

    function isInflight(sectionId: string, direction: "up" | "down", hunkMetaId: string): boolean {
        return sections.get(sectionId)?.inflight.has(`${direction}:${hunkMetaId}`) ?? false;
    }

    function isDownwardEof(sectionId: string): boolean {
        return sections.get(sectionId)?.downwardEof ?? false;
    }

    async function expandUp(req: ExpansionRequest, hunk: HunkInfo): Promise<void> {
        const state = ensureState(req.sectionId);
        const key = `up:${hunk.metaId}`;

        if (state.inflight.has(key)) {
            return;
        }

        const lowestOldLine = computeLowestExpandedInGap(state.expansions, hunk);
        const top = hunk.prevOldEnd + 1;
        const endLine = lowestOldLine - 1;

        if (endLine < top) {
            return;
        }

        const startLine = Math.max(top, endLine - CONTEXT_EXPANSION_LINE_COUNT + 1);

        state.inflight.add(key);

        try {
            const { lines } = await fetchRange(req, startLine, endLine);
            const delta = hunk.prevNewEnd - hunk.prevOldEnd;
            const additions = mapToExpansions(lines, startLine, delta);

            state.expansions = [...state.expansions, ...additions];
        } finally {
            state.inflight.delete(key);
        }
    }

    async function expandDown(
        req: ExpansionRequest,
        hunk: HunkInfo,
        nextHunkOldStart: number | null,
    ): Promise<void> {
        const state = ensureState(req.sectionId);
        const key = `down:${hunk.metaId}`;

        if (state.inflight.has(key)) {
            return;
        }

        if (hunk.isLast && state.downwardEof) {
            return;
        }

        const highestOldLine = computeHighestExpandedInGap(
            state.expansions,
            hunk,
            nextHunkOldStart,
        );
        const startLine = highestOldLine + 1;
        let endLine = startLine + CONTEXT_EXPANSION_LINE_COUNT - 1;

        if (nextHunkOldStart != null && endLine >= nextHunkOldStart) {
            endLine = nextHunkOldStart - 1;
        }

        if (endLine < startLine) {
            return;
        }

        state.inflight.add(key);

        try {
            const { lines, eof } = await fetchRange(req, startLine, endLine);
            const delta = hunk.newEnd - hunk.oldEnd;
            const additions = mapToExpansions(lines, startLine, delta);

            state.expansions = [...state.expansions, ...additions];

            if (hunk.isLast && eof) {
                state.downwardEof = true;
            }
        } finally {
            state.inflight.delete(key);
        }
    }

    function reset(sectionId?: string): void {
        if (sectionId) {
            sections.delete(sectionId);

            return;
        }

        sections.clear();
    }

    return { getExpansions, isInflight, isDownwardEof, expandUp, expandDown, reset };
}

function computeLowestExpandedInGap(expansions: ExpandedContext[], hunk: HunkInfo): number {
    let lowest = hunk.oldStart;

    for (const exp of expansions) {
        if (exp.oldLine > hunk.prevOldEnd && exp.oldLine < hunk.oldStart && exp.oldLine < lowest) {
            lowest = exp.oldLine;
        }
    }

    return lowest;
}

function computeHighestExpandedInGap(
    expansions: ExpandedContext[],
    hunk: HunkInfo,
    nextHunkOldStart: number | null,
): number {
    let highest = hunk.oldEnd;
    const upperBound = nextHunkOldStart ?? Number.POSITIVE_INFINITY;

    for (const exp of expansions) {
        if (exp.oldLine > hunk.oldEnd && exp.oldLine < upperBound && exp.oldLine > highest) {
            highest = exp.oldLine;
        }
    }

    return highest;
}

function mapToExpansions(lines: string[], startLine: number, delta: number): ExpandedContext[] {
    return lines.map((text, index) => ({
        oldLine: startLine + index,
        newLine: startLine + index + delta,
        text,
    }));
}
