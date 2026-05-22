import { computed, type Ref } from "vue";
import type { DiffSection } from "@git-diff/contracts";
import { getHunkInfos, parsePatch, type HunkInfo } from "@lib/patch";

/**
 * Per-instance hunk-info cache. `parsePatch` is deterministic given
 * (section, hideWhitespace); caching per (section.id, whitespace
 * toggle) avoids rewalking the patch on every render.
 *
 * Replaces the hand-rolled `Map` that lived inside DiffBody.vue with a
 * reactive computed that invalidates automatically when
 * `hideWhitespace` flips, so we no longer have a manual cache key
 * encoding the toggle.
 */
export function useHunkInfo(hideWhitespace: Ref<boolean>) {
    const cache = computed<Map<string, HunkInfo[]>>(() => {
        // Force re-derivation on whitespace toggle.
        void hideWhitespace.value;

        return new Map();
    });

    function hunksFor(section: DiffSection): HunkInfo[] {
        const map = cache.value;
        const cached = map.get(section.id);

        if (cached) {
            return cached;
        }

        const infos = getHunkInfos(parsePatch(section, hideWhitespace.value));

        map.set(section.id, infos);

        return infos;
    }

    function findHunk(section: DiffSection, metaId: string): HunkInfo | null {
        return hunksFor(section).find((h) => h.metaId === metaId) ?? null;
    }

    function nextHunkOldStart(section: DiffSection, hunk: HunkInfo): number | null {
        const infos = hunksFor(section);
        const index = infos.findIndex((h) => h.metaId === hunk.metaId);

        if (index < 0 || index === infos.length - 1) {
            return null;
        }

        return infos[index + 1]!.oldStart;
    }

    return { hunksFor, findHunk, nextHunkOldStart };
}
