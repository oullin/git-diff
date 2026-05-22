import { computed, type Ref } from "vue";
import type { DiffHunkStyle } from "@git-diff/contracts";
import { diffBgs, type DiffStyleColors } from "@lib/accent";

/** Kind label for a rendered cell in a split-view diff row. */
export type DiffCellKind = "ctx" | "add" | "rem" | "empty";

/**
 * Per-instance diff-cell style derivations. Tracks the active accent
 * palette and exposes the background, line-number background, side-bar
 * colour, and sign character for a given cell kind. Splits these
 * concerns out of DiffBody so the component template binds against a
 * small typed surface.
 */
export function useDiffStyles(diffStyle: Ref<DiffHunkStyle>) {
    const colors = computed<DiffStyleColors>(() => diffBgs(diffStyle.value));

    function bgFor(kind: DiffCellKind): string {
        if (kind === "add") {return colors.value.addBg;}

        if (kind === "rem") {return colors.value.remBg;}

        return "transparent";
    }

    function numBgFor(kind: DiffCellKind): string {
        if (kind === "add") {return colors.value.addNum;}

        if (kind === "rem") {return colors.value.remNum;}

        return "transparent";
    }

    function barFor(kind: DiffCellKind): string {
        if (kind === "add") {return colors.value.addBar;}

        if (kind === "rem") {return colors.value.remBar;}

        return "transparent";
    }

    function sign(kind: DiffCellKind): string {
        if (kind === "add") {return "+";}

        if (kind === "rem") {return "−";}

        return " ";
    }

    return { colors, bgFor, numBgFor, barFor, sign };
}
