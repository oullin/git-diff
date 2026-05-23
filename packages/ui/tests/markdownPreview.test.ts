// @vitest-environment happy-dom
import { describe, expect, test } from "vitest";

import type { ChangedFile } from "@git-diff/contracts";

import {
    getAddedLineNumbers,
    isMarkdownPath,
    renderMarkdownWithLineAnchors,
} from "@lib/markdownPreview";

function makeFile(
    status: ChangedFile["status"],
    patch: string,
    overrides: Partial<ChangedFile> = {},
): ChangedFile {
    return {
        path: "README.md",
        status,
        additions: 1,
        deletions: 0,
        binary: false,
        fingerprint: "fp",
        sections: [
            {
                id: "sec1",
                kind: "unstaged",
                binary: false,
                patch,
            },
        ],
        ...overrides,
    };
}

describe("isMarkdownPath", () => {
    test("matches .md and .markdown case-insensitively", () => {
        expect(isMarkdownPath("README.md")).toBe(true);
        expect(isMarkdownPath("docs/PLAN.MD")).toBe(true);
        expect(isMarkdownPath("notes.markdown")).toBe(true);
    });

    test("rejects non-markdown paths", () => {
        expect(isMarkdownPath("src/app.ts")).toBe(false);
        expect(isMarkdownPath("README")).toBe(false);
    });
});

describe("getAddedLineNumbers", () => {
    test("returns the set of newLine numbers for added lines", () => {
        const patch = [
            "@@ -1,3 +1,5 @@",
            " context 1",
            "+new 2",
            "+new 3",
            " context 4",
            " context 5",
        ].join("\n");
        const file = makeFile("modified", patch);

        expect([...getAddedLineNumbers(file)].sort()).toEqual([2, 3]);
    });

    test("returns empty for added files (every line is implicitly an addition)", () => {
        const patch = ["@@ -0,0 +1,2 @@", "+new 1", "+new 2"].join("\n");
        const file = makeFile("added", patch);

        expect(getAddedLineNumbers(file).size).toBe(0);
    });
});

describe("renderMarkdownWithLineAnchors", () => {
    test("wraps each source line in a data-line element", () => {
        const html = renderMarkdownWithLineAnchors(
            "first\nsecond",
            new Set<number>(),
            (line) => `<strong>${line}</strong>`,
            (x) => x,
        );

        expect(html).toContain('data-line="1"');
        expect(html).toContain('data-line="2"');
        expect(html).toContain("<strong>first</strong>");
        expect(html).toContain("<strong>second</strong>");
    });

    test("adds the md-line-added class for highlighted line numbers", () => {
        const html = renderMarkdownWithLineAnchors(
            "first\nsecond",
            new Set<number>([2]),
            (line) => line,
            (x) => x,
        );

        expect(html).toContain('class="md-line" data-line="1"');
        expect(html).toContain('class="md-line md-line-added" data-line="2"');
    });
});
