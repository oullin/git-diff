import type { ChangedFile } from "@git-diff/contracts";
import type { ComputedRef } from "vue";

export interface UseDiffNavigationOptions {
    files: ComputedRef<ChangedFile[]>;
    changedIndex: ComputedRef<number>;
    onSelect: (path: string) => void;
}

/**
 * Diff-list and hunk-anchor navigation primitives. `selectAdjacent`
 * walks the changed-file list by delta; `jumpToHunk` scrolls the
 * viewport to the next/previous `[data-hunk-anchor]` element relative
 * to the screen midpoint and flashes a highlight on the target.
 *
 * Keyboard binding lives in @composables/useKeyboardShortcuts; this
 * module owns the navigation actions only so callers can wire them up
 * outside of a keyboard-driven context (e.g. button clicks).
 */
export function useDiffNavigation({ files, changedIndex, onSelect }: UseDiffNavigationOptions) {
    function selectAdjacent(delta: number): void {
        if (files.value.length === 0) {
            return;
        }

        const next = Math.min(Math.max(changedIndex.value + delta, 0), files.value.length - 1);
        const file = files.value[next];

        if (file) {
            onSelect(file.path);
        }
    }

    function jumpToHunk(delta: number): void {
        const anchors = Array.from(document.querySelectorAll<HTMLElement>("[data-hunk-anchor]"));

        if (anchors.length === 0) {
            return;
        }

        const midpoint = window.innerHeight / 2;
        let current = 0;

        for (let i = 0; i < anchors.length; i++) {
            const rect = anchors[i]!.getBoundingClientRect();

            if (rect.top <= midpoint) {
                current = i;
            } else {
                break;
            }
        }

        const target = Math.max(0, Math.min(anchors.length - 1, current + delta));
        const element = anchors[target]!;

        element.scrollIntoView({ block: "center", behavior: "smooth" });
        element.classList.add("gd-hunk-flash");
        window.setTimeout(() => element.classList.remove("gd-hunk-flash"), 350);
    }

    return { selectAdjacent, jumpToHunk };
}
