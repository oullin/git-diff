// @vitest-environment happy-dom
import { describe, expect, test, vi } from "vitest";

import { useContextExpansionControls } from "@composables/useContextExpansionControls";
import type { ExpandedContext, HunkInfo } from "@git-diff/domain/diff";
import type { ExpansionRequest } from "@composables/useContextExpansion";
import type { DiffSection } from "@git-diff/domain";

type ExpandUpFn = (req: ExpansionRequest, hunk: HunkInfo) => Promise<void>;

type ExpandDownFn = (req: ExpansionRequest, hunk: HunkInfo, next: number | null) => Promise<void>;

const SECTION = { id: "s1", kind: "unstaged" } as DiffSection;

function hunk(partial: Partial<HunkInfo>): HunkInfo {
  return {
    metaId: "m1",
    oldStart: 10,
    oldEnd: 20,
    newEnd: 20,
    prevOldEnd: 0,
    prevNewEnd: 0,
    isLast: false,
    ...partial,
  } as HunkInfo;
}

function controls(overrides: {
  expansions?: ExpandedContext[];
  nextStart?: number | null;
  downwardEof?: boolean;
  expandUp?: ReturnType<typeof vi.fn>;
  expandDown?: ReturnType<typeof vi.fn>;
  repoRoot?: string;
}) {
  return useContextExpansionControls({
    getExpansions: () => overrides.expansions ?? [],
    isDownwardEof: () => overrides.downwardEof ?? false,
    nextHunkOldStart: () => overrides.nextStart ?? null,
    expandUp: (overrides.expandUp ?? vi.fn().mockResolvedValue(undefined)) as unknown as ExpandUpFn,
    expandDown: (overrides.expandDown ??
      vi.fn().mockResolvedValue(undefined)) as unknown as ExpandDownFn,
    repoRoot: () => overrides.repoRoot ?? "/repo",
    filePath: () => "a.ts",
    commitRef: () => undefined,
  });
}

describe("canExpandUp", () => {
  test("true when there is a visible gap above the hunk", () => {
    const c = controls({});

    expect(c.canExpandUp(SECTION, hunk({ oldStart: 10, prevOldEnd: 0 }))).toBe(true);
  });

  test("false when the hunk butts against the previous one", () => {
    const c = controls({});

    expect(c.canExpandUp(SECTION, hunk({ oldStart: 10, prevOldEnd: 9 }))).toBe(false);
  });

  test("false for a null hunk", () => {
    expect(controls({}).canExpandUp(SECTION, null)).toBe(false);
  });
});

describe("canExpandDown", () => {
  test("uses next hunk start as the bound", () => {
    const c = controls({ nextStart: 30 });

    expect(c.canExpandDown(SECTION, hunk({ oldEnd: 20 }))).toBe(true);
  });

  test("false when adjacent to the next hunk", () => {
    const c = controls({ nextStart: 21 });

    expect(c.canExpandDown(SECTION, hunk({ oldEnd: 20 }))).toBe(false);
  });

  test("last hunk depends on downward EOF flag", () => {
    expect(controls({ downwardEof: false }).canExpandDown(SECTION, hunk({ isLast: true }))).toBe(
      true,
    );
    expect(controls({ downwardEof: true }).canExpandDown(SECTION, hunk({ isLast: true }))).toBe(
      false,
    );
  });
});

describe("dispatch", () => {
  test("onExpandUp invokes expandUp with the request", () => {
    const expandUp = vi.fn().mockResolvedValue(undefined);
    const c = controls({ expandUp });

    c.onExpandUp(SECTION, hunk({}));

    expect(expandUp).toHaveBeenCalledOnce();
    expect(expandUp.mock.calls[0][0]).toMatchObject({ sectionId: "s1", filePath: "a.ts" });
  });

  test("onExpandUp is a no-op without a repo root", () => {
    const expandUp = vi.fn();
    const c = controls({ expandUp, repoRoot: "" });

    c.onExpandUp(SECTION, hunk({}));

    expect(expandUp).not.toHaveBeenCalled();
  });

  test("onExpandDown forwards the next hunk start", () => {
    const expandDown = vi.fn().mockResolvedValue(undefined);
    const c = controls({ expandDown, nextStart: 30 });

    c.onExpandDown(SECTION, hunk({}));

    expect(expandDown).toHaveBeenCalledOnce();
    expect(expandDown.mock.calls[0][2]).toBe(30);
  });
});
