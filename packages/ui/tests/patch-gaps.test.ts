// @vitest-environment happy-dom
import { describe, expect, test } from "vitest";

import {
  applyExpansions,
  getHunkInfos,
  parseHunkHeader,
  parsePatch,
  splitPatchLines,
} from "@git-diff/domain/diff";

const SECTION = {
  id: "sec1",
  kind: "unstaged" as const,
  binary: false,
  patch: [
    "@@ -10,3 +10,3 @@ first hunk",
    " context line a",
    "-old line b",
    "+new line b",
    " context line c",
    "@@ -50,2 +50,2 @@ second hunk",
    "-old line 50",
    "+new line 50",
    " context line 51",
  ].join("\n"),
};

describe("parseHunkHeader", () => {
  test("parses standard @@ headers", () => {
    expect(parseHunkHeader("@@ -10,3 +12,4 @@ trailer")).toEqual({
      oldStart: 10,
      oldCount: 3,
      newStart: 12,
      newCount: 4,
    });
  });

  test("defaults missing counts to 1", () => {
    expect(parseHunkHeader("@@ -10 +12 @@")).toEqual({
      oldStart: 10,
      oldCount: 1,
      newStart: 12,
      newCount: 1,
    });
  });

  test("returns null for non-headers", () => {
    expect(parseHunkHeader(" not a header")).toBeNull();
    expect(parseHunkHeader("@@ malformed")).toBeNull();
  });
});

describe("getHunkInfos", () => {
  test("derives boundaries and gap info for every hunk", () => {
    const lines = parsePatch(SECTION, false);
    const infos = getHunkInfos(lines);

    expect(infos).toHaveLength(2);

    expect(infos[0]).toMatchObject({
      oldStart: 10,
      newStart: 10,
      oldEnd: 12,
      newEnd: 12,
      prevOldEnd: 0,
      prevNewEnd: 0,
      isLast: false,
    });

    expect(infos[1]).toMatchObject({
      oldStart: 50,
      newStart: 50,
      oldEnd: 51,
      newEnd: 51,
      // The gap between hunks: 13..49 in old space.
      prevOldEnd: 12,
      prevNewEnd: 12,
      isLast: true,
    });
  });
});

describe("applyExpansions", () => {
  test("splices expansion rows before the next hunk header in oldLine order", () => {
    const base = parsePatch(SECTION, false);

    const merged = applyExpansions(base, [
      { oldLine: 14, newLine: 14, text: "gap line 14" },
      { oldLine: 13, newLine: 13, text: "gap line 13" },
    ]);

    const ctxLines = merged.filter((l) => l.type === "context").map((l) => l.text);

    // Originals: "context line a", "context line c", "context line 51".
    // Spliced rows ("gap line 13", "gap line 14") should appear between
    // "context line c" and the second hunk header.
    expect(ctxLines).toEqual([
      "context line a",
      "context line c",
      "gap line 13",
      "gap line 14",
      "context line 51",
    ]);
  });

  test("emits trailing expansions after the last hunk", () => {
    const base = parsePatch(SECTION, false);

    const merged = applyExpansions(base, [
      { oldLine: 52, newLine: 52, text: "below line 52" },
      { oldLine: 53, newLine: 53, text: "below line 53" },
    ]);

    const last = merged.slice(-2);

    expect(last.map((l) => l.text)).toEqual(["below line 52", "below line 53"]);
    expect(last.every((l) => l.type === "context")).toBe(true);
  });

  test("returns a copy of the input when there are no expansions", () => {
    const base = parsePatch(SECTION, false);
    const merged = applyExpansions(base, []);

    expect(merged).not.toBe(base);
    expect(merged).toEqual(base);
  });
});

describe("splitPatchLines", () => {
  test("treats expansion-derived context lines as single-cell context rows", () => {
    const base = parsePatch(SECTION, false);
    const merged = applyExpansions(base, [{ oldLine: 13, newLine: 13, text: "gap line 13" }]);
    const rows = splitPatchLines(merged);

    const contextRows = rows
      .filter((r) => r.kind === "context")
      .map((r) => (r as { kind: "context"; line: { text: string } }).line.text);

    expect(contextRows).toContain("gap line 13");
  });
});
