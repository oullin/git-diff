import type { Ref } from "vue";
import type { BundledLanguage } from "shiki";
import { highlighterRev, highlightLine } from "@lib/highlight";
import { computeWordHi, type Range } from "@git-diff/domain/diff";

function escapeHtml(value: string): string {
    return value.replace(/[&<>"']/g, (ch) => {
        switch (ch) {
            case "&":
                return "&amp;";
            case "<":
                return "&lt;";
            case ">":
                return "&gt;";
            case '"':
                return "&quot;";
            default:
                return "&#39;";
        }
    });
}

export function useDiffHighlighting(lang: Ref<BundledLanguage | null>) {
    function highlightHtml(text: string): string {
        // Read the rev so Vue re-runs callers when a language finishes loading.
        void highlighterRev.value;

        return highlightLine(text || " ", lang.value);
    }

    function withRanges(text: string, ranges: Range[], cls: "wh-add" | "wh-rem"): string {
        if (!ranges.length) {
            return escapeHtml(text);
        }

        const out: string[] = [];
        let cursor = 0;

        for (const [start, end] of ranges) {
            if (start > cursor) {
                out.push(escapeHtml(text.slice(cursor, start)));
            }

            out.push(`<span class="${cls}">${escapeHtml(text.slice(start, end))}</span>`);
            cursor = end;
        }

        if (cursor < text.length) {
            out.push(escapeHtml(text.slice(cursor)));
        }

        return out.join("");
    }

    return { highlightHtml, withRanges, escapeHtml, computeWordHi };
}
