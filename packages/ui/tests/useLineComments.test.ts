// @vitest-environment happy-dom
import { describe, expect, test } from "vitest";

import { lineSideAndNumber, useLineComments } from "@composables/useLineComments";
import type { PatchLine } from "@git-diff/domain/diff";
import type { DiffSection, ReviewComment } from "@git-diff/domain";

function line(partial: Partial<PatchLine>): PatchLine {
    return {
        id: "l",
        type: "context",
        text: "",
        oldLine: undefined,
        newLine: undefined,
        ...partial,
    } as PatchLine;
}

const SECTION = { id: "s1", kind: "unstaged" } as DiffSection;

describe("lineSideAndNumber", () => {
    test("added line anchors right on its new line", () => {
        expect(lineSideAndNumber(line({ type: "add", newLine: 12 }))).toEqual({
            side: "right",
            lineNumber: 12,
        });
    });

    test("deleted line anchors left on its old line", () => {
        expect(lineSideAndNumber(line({ type: "del", oldLine: 7 }))).toEqual({
            side: "left",
            lineNumber: 7,
        });
    });

    test("context line prefers the right (new) side", () => {
        expect(lineSideAndNumber(line({ type: "context", oldLine: 3, newLine: 4 }))).toEqual({
            side: "right",
            lineNumber: 4,
        });
    });

    test("falls back to left when only an old line exists", () => {
        expect(
            lineSideAndNumber(line({ type: "context", oldLine: 3, newLine: undefined })),
        ).toEqual({
            side: "left",
            lineNumber: 3,
        });
    });

    test("returns null when no line numbers exist", () => {
        expect(lineSideAndNumber(line({ oldLine: undefined, newLine: undefined }))).toBeNull();
    });
});

describe("useLineComments.commentsForLine", () => {
    const comment = (partial: Partial<ReviewComment>): ReviewComment =>
        ({
            id: 1,
            filePath: "a.ts",
            diffSection: "unstaged",
            side: "right",
            lineNumber: 4,
            ...partial,
        }) as ReviewComment;

    const { commentsForLine } = useLineComments({
        comments: () => [
            comment({ id: 1, side: "right", lineNumber: 4 }),
            comment({ id: 2, side: "left", lineNumber: 3 }),
            comment({ id: 3, filePath: "other.ts", side: "right", lineNumber: 4 }),
        ],
        filePath: () => "a.ts",
    });

    test("matches a comment on the right side by new line", () => {
        const got = commentsForLine(SECTION, line({ oldLine: 88, newLine: 4 }));

        expect(got.map((c) => c.id)).toEqual([1]);
    });

    test("matches a left-side comment by old line", () => {
        const got = commentsForLine(SECTION, line({ oldLine: 3, newLine: 99 }));

        expect(got.map((c) => c.id)).toEqual([2]);
    });

    test("ignores comments on other files", () => {
        const got = commentsForLine(SECTION, line({ oldLine: 3, newLine: 4 }));

        expect(got.some((c) => c.id === 3)).toBe(false);
    });

    test("returns empty for an undefined line", () => {
        expect(commentsForLine(SECTION, undefined)).toEqual([]);
    });
});
