import type { DiffSection } from "@git-diff/domain";
import type { ExpandedContext, HunkInfo } from "@git-diff/domain/diff";
import type { ExpansionRequest } from "@composables/useContextExpansion";

/**
 * Visibility/affordance math layered on top of useContextExpansion: decides
 * whether the up/down expand controls are actionable and dispatches the
 * expansion requests. Pure functions of the injected getters — unit-testable
 * without mounting a component.
 */
export interface ExpansionControlsOptions {
    getExpansions: (sectionId: string) => ExpandedContext[];
    isDownwardEof: (sectionId: string) => boolean;
    nextHunkOldStart: (section: DiffSection, hunk: HunkInfo) => number | null;
    expandUp: (req: ExpansionRequest, hunk: HunkInfo) => Promise<void>;
    expandDown: (
        req: ExpansionRequest,
        hunk: HunkInfo,
        nextHunkOldStart: number | null,
    ) => Promise<void>;
    repoRoot: () => string;
    filePath: () => string;
    commitRef: () => string | undefined;
}

export function useContextExpansionControls(opts: ExpansionControlsOptions) {
    function expansionRequest(section: DiffSection): ExpansionRequest {
        return {
            sectionId: section.id,
            repoRoot: opts.repoRoot(),
            filePath: opts.filePath(),
            ref: opts.commitRef(),
        };
    }

    function lowestVisibleAbove(section: DiffSection, hunk: HunkInfo): number {
        let lowest = hunk.oldStart;

        for (const exp of opts.getExpansions(section.id)) {
            if (
                exp.oldLine > hunk.prevOldEnd &&
                exp.oldLine < hunk.oldStart &&
                exp.oldLine < lowest
            ) {
                lowest = exp.oldLine;
            }
        }

        return lowest;
    }

    function highestVisibleBelow(
        section: DiffSection,
        hunk: HunkInfo,
        nextStart: number | null,
    ): number {
        let highest = hunk.oldEnd;
        const upper = nextStart ?? Number.POSITIVE_INFINITY;

        for (const exp of opts.getExpansions(section.id)) {
            if (exp.oldLine > hunk.oldEnd && exp.oldLine < upper && exp.oldLine > highest) {
                highest = exp.oldLine;
            }
        }

        return highest;
    }

    function canExpandUp(section: DiffSection, hunk: HunkInfo | null): boolean {
        if (!hunk) {
            return false;
        }

        return lowestVisibleAbove(section, hunk) > hunk.prevOldEnd + 1;
    }

    function canExpandDown(section: DiffSection, hunk: HunkInfo | null): boolean {
        if (!hunk) {
            return false;
        }

        const next = opts.nextHunkOldStart(section, hunk);
        const highest = highestVisibleBelow(section, hunk, next);

        if (next != null) {
            return highest < next - 1;
        }

        return !opts.isDownwardEof(section.id);
    }

    function onExpandUp(section: DiffSection, hunk: HunkInfo | null): void {
        if (!hunk || !opts.repoRoot()) {
            return;
        }

        void opts.expandUp(expansionRequest(section), hunk);
    }

    function onExpandDown(section: DiffSection, hunk: HunkInfo | null): void {
        if (!hunk || !opts.repoRoot()) {
            return;
        }

        void opts.expandDown(expansionRequest(section), hunk, opts.nextHunkOldStart(section, hunk));
    }

    return { canExpandUp, canExpandDown, onExpandUp, onExpandDown };
}
