// @vitest-environment happy-dom
import { afterEach, describe, expect, test } from "vitest";

import { commentRangeLabel, useLineSelection } from "../src/composables/useLineSelection";

afterEach(() => {
    useLineSelection().clear();
});

describe("commentRangeLabel", () => {
    test("returns single-line label when start matches end", () => {
        expect(commentRangeLabel({ side: "right", lineNumber: 12 })).toBe("New line 12");
        expect(
            commentRangeLabel({
                side: "right",
                lineNumber: 12,
                startLineNumber: 12,
                startSide: "right",
            }),
        ).toBe("New line 12");
    });

    test("uses same-side range form when both lines stay on one side", () => {
        expect(
            commentRangeLabel({
                side: "right",
                lineNumber: 18,
                startLineNumber: 12,
                startSide: "right",
            }),
        ).toBe("New lines 12–18");
    });

    test("uses cross-side arrow form when start and end sides differ", () => {
        expect(
            commentRangeLabel({
                side: "right",
                lineNumber: 18,
                startLineNumber: 12,
                startSide: "left",
            }),
        ).toBe("Old line 12 → New line 18");
    });
});

describe("useLineSelection.rangeTo", () => {
    test("returns null when no anchor is set", () => {
        const sel = useLineSelection();

        expect(sel.rangeTo({ sectionId: "s", side: "right", lineNumber: 10 })).toBeNull();
    });

    test("orders same-side ranges top-to-bottom regardless of click order", () => {
        const sel = useLineSelection();

        sel.setAnchor({ sectionId: "s", side: "right", lineNumber: 20 });
        const range = sel.rangeTo({ sectionId: "s", side: "right", lineNumber: 5 });

        expect(range).toEqual({
            sectionId: "s",
            startSide: "right",
            startLine: 5,
            endSide: "right",
            endLine: 20,
        });
    });

    test("places left side as start for cross-side ranges, regardless of click order", () => {
        const sel = useLineSelection();

        sel.setAnchor({ sectionId: "s", side: "right", lineNumber: 18 });
        const range = sel.rangeTo({ sectionId: "s", side: "left", lineNumber: 12 });

        expect(range).toEqual({
            sectionId: "s",
            startSide: "left",
            startLine: 12,
            endSide: "right",
            endLine: 18,
        });
    });

    test("ignores anchors from a different section", () => {
        const sel = useLineSelection();

        sel.setAnchor({ sectionId: "section-a", side: "right", lineNumber: 10 });
        expect(sel.rangeTo({ sectionId: "section-b", side: "right", lineNumber: 20 })).toBeNull();
    });
});
