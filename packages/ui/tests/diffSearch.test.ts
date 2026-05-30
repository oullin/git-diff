// @vitest-environment happy-dom
import { afterEach, describe, expect, test } from "vitest";
import { applyHighlights, clearHighlights, searchDiff, setActiveHighlight } from "@lib/diffSearch";

function fixture(lines: string[]): HTMLElement {
    const root = document.createElement("div");

    for (const line of lines) {
        const node = document.createElement("span");

        node.setAttribute("data-diff-line-text", "");
        node.textContent = line;
        root.appendChild(node);
    }

    document.body.appendChild(root);

    return root;
}

afterEach(() => {
    document.body.innerHTML = "";
});

describe("searchDiff", () => {
    test("returns empty result for empty query", () => {
        const root = fixture(["hello world"]);

        expect(searchDiff("", root)).toEqual({ matches: [], matchedLines: 0 });
    });

    test("case-insensitive by default", () => {
        const root = fixture(["Hello World", "say HELLO"]);
        const result = searchDiff("hello", root);

        expect(result.matches.length).toBe(2);
        expect(result.matchedLines).toBe(2);
    });

    test("case-sensitive when requested", () => {
        const root = fixture(["Hello World", "say HELLO"]);
        const result = searchDiff("Hello", root, { caseSensitive: true });

        expect(result.matches.length).toBe(1);
    });

    test("counts overlapping occurrences once per position", () => {
        const root = fixture(["abababab"]);
        const result = searchDiff("aba", root);

        expect(result.matches.length).toBe(2);
    });

    test("skips elements without the data attribute", () => {
        const root = document.createElement("div");
        const tagged = document.createElement("span");

        tagged.setAttribute("data-diff-line-text", "");
        tagged.textContent = "foo bar";
        root.appendChild(tagged);
        const untagged = document.createElement("span");

        untagged.textContent = "foo bar";
        root.appendChild(untagged);
        document.body.appendChild(root);
        expect(searchDiff("foo", root).matches.length).toBe(1);
    });
});

describe("applyHighlights / clearHighlights", () => {
    test("wraps matches in spans and unwraps them", () => {
        const root = fixture(["the quick brown fox"]);
        const result = searchDiff("quick", root);
        const wrappers = applyHighlights(result.matches);

        expect(wrappers.length).toBe(1);
        expect(root.querySelectorAll(".gd-search-hit").length).toBe(1);
        clearHighlights(root);
        expect(root.querySelectorAll(".gd-search-hit").length).toBe(0);
        expect(root.firstElementChild?.textContent).toBe("the quick brown fox");
    });

    test("setActiveHighlight toggles the active class on one wrapper at a time", () => {
        const root = fixture(["hit one", "hit two"]);
        const wrappers = applyHighlights(searchDiff("hit", root).matches);

        setActiveHighlight(wrappers, 0);
        expect(wrappers[0]!.classList.contains("gd-search-hit-active")).toBe(true);
        expect(wrappers[1]!.classList.contains("gd-search-hit-active")).toBe(false);
        setActiveHighlight(wrappers, 1);
        expect(wrappers[0]!.classList.contains("gd-search-hit-active")).toBe(false);
        expect(wrappers[1]!.classList.contains("gd-search-hit-active")).toBe(true);
    });
});
