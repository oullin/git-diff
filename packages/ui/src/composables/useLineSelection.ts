import { reactive } from "vue";

export type LineSide = "left" | "right";

export interface LineAnchor {
    sectionId: string;
    side: LineSide;
    lineNumber: number;
}

export interface LineSelectionRange {
    sectionId: string;
    startSide: LineSide;
    startLine: number;
    endSide: LineSide;
    endLine: number;
}

interface SelectionState {
    anchor: LineAnchor | null;
}

const state: SelectionState = reactive({ anchor: null });

export function useLineSelection() {
    function setAnchor(anchor: LineAnchor | null): void {
        state.anchor = anchor;
    }

    function getAnchor(): LineAnchor | null {
        return state.anchor;
    }

    function clear(): void {
        state.anchor = null;
    }

    /**
     * Given the current anchor and a target line, returns the normalised range.
     * If the target is on the same side as the anchor and earlier in the file,
     * the start/end swap so the range always reads top-to-bottom. Cross-side
     * ranges always run from `left` to `right` regardless of click order.
     */
    function rangeTo(target: LineAnchor): LineSelectionRange | null {
        const anchor = state.anchor;

        if (!anchor || anchor.sectionId !== target.sectionId) {
            return null;
        }

        if (anchor.side === target.side) {
            const [startLine, endLine] =
                anchor.lineNumber <= target.lineNumber
                    ? [anchor.lineNumber, target.lineNumber]
                    : [target.lineNumber, anchor.lineNumber];

            return {
                sectionId: anchor.sectionId,
                startSide: anchor.side,
                startLine,
                endSide: anchor.side,
                endLine,
            };
        }

        // Cross-side: deletions (left) before additions (right) by convention.
        if (anchor.side === "left") {
            return {
                sectionId: anchor.sectionId,
                startSide: "left",
                startLine: anchor.lineNumber,
                endSide: "right",
                endLine: target.lineNumber,
            };
        }

        return {
            sectionId: anchor.sectionId,
            startSide: "left",
            startLine: target.lineNumber,
            endSide: "right",
            endLine: anchor.lineNumber,
        };
    }

    return { setAnchor, getAnchor, clear, rangeTo };
}

/**
 * Renders a human-readable label for a comment range. Single-line and
 * single-side ranges collapse to "Old line 12" / "New line 18" forms; mixed
 * ranges expand to "Old line 12 → New line 18".
 */
export function commentRangeLabel(comment: {
    side: string;
    lineNumber: number;
    startLineNumber?: number;
    startSide?: string;
}): string {
    const startLine = comment.startLineNumber;
    const startSide = comment.startSide ?? comment.side;
    const sideLabel = (side: string): string => (side === "left" ? "Old" : "New");

    if (startLine == null || (startLine === comment.lineNumber && startSide === comment.side)) {
        return `${sideLabel(comment.side)} line ${comment.lineNumber}`;
    }

    if (startSide === comment.side) {
        return `${sideLabel(comment.side)} lines ${startLine}–${comment.lineNumber}`;
    }

    return `${sideLabel(startSide)} line ${startLine} → ${sideLabel(comment.side)} line ${comment.lineNumber}`;
}
