// @vitest-environment happy-dom
import { describe, expect, test } from "vitest";

import { highlightLine, languageFor } from "@lib/highlight";
import { MAX_LINE_TOKENIZE_LENGTH } from "@lib/highlightPool";

describe("languageFor", () => {
    test("maps file extensions to Shiki languages", () => {
        expect(languageFor("src/lib/foo.ts")).toBe("ts");
        expect(languageFor("README.md")).toBe("md");
        expect(languageFor("Dockerfile")).toBe("docker");
    });

    test("returns null for unknown extensions and extensionless files", () => {
        expect(languageFor("LICENSE")).toBeNull();
        expect(languageFor("src/foo.unknownext")).toBeNull();
    });
});

describe("highlightLine", () => {
    test("returns escaped plaintext when no language is detected", () => {
        expect(highlightLine("<script>alert(1)</script>", null)).toBe(
            "&lt;script&gt;alert(1)&lt;/script&gt;",
        );
    });

    test("returns escaped plaintext for over-long lines without trying to tokenize", () => {
        const longLine = "a".repeat(MAX_LINE_TOKENIZE_LENGTH + 1);

        expect(highlightLine(longLine, "ts")).toBe(longLine);
    });

    test("returns escaped plaintext on cache miss (async tokens land later)", () => {
        // Synchronous behaviour: first call returns plaintext even when a
        // language is provided. The actual tokenized HTML arrives via the
        // worker / fallback path and triggers a re-render through
        // highlighterRev — covered by integration testing, not here.
        expect(highlightLine("const x = 1;", "ts")).toBe("const x = 1;");
    });
});
