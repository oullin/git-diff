// @vitest-environment happy-dom
import { afterEach, describe, expect, test } from "vitest";
import { computed, effectScope, nextTick, ref } from "vue";
import { estimatedDiffHeight, useLazyDiffFile } from "@composables/useLazyDiffFile";

import {
  forceRenderAllDiffFiles,
  forceRenderDiffFile,
  isDiffFileForced,
  resetLazyRender,
} from "@composables/useLazyRender";

afterEach(() => {
  resetLazyRender();
});

describe("estimatedDiffHeight", () => {
  test("clamps tiny diffs to a minimum", () => {
    expect(estimatedDiffHeight(1)).toBe(120);
  });

  test("clamps huge diffs to a maximum", () => {
    expect(estimatedDiffHeight(100000)).toBe(4000);
  });

  test("scales linearly between the bounds", () => {
    expect(estimatedDiffHeight(100)).toBe(2048);
  });
});

describe("useLazyRender registry", () => {
  test("forceRenderDiffFile marks a single path", () => {
    expect(isDiffFileForced("a.ts")).toBe(false);
    forceRenderDiffFile("a.ts");
    expect(isDiffFileForced("a.ts")).toBe(true);
    expect(isDiffFileForced("b.ts")).toBe(false);
  });

  test("forceRenderAllDiffFiles marks every path", () => {
    forceRenderAllDiffFiles();
    expect(isDiffFileForced("anything")).toBe(true);
  });

  test("resetLazyRender clears state", () => {
    forceRenderAllDiffFiles();
    forceRenderDiffFile("a.ts");
    resetLazyRender();
    expect(isDiffFileForced("a.ts")).toBe(false);
    expect(isDiffFileForced("anything")).toBe(false);
  });
});

describe("useLazyDiffFile", () => {
  test("renders immediately when deferral is disabled", () => {
    const scope = effectScope();

    scope.run(() => {
      const { rendered } = useLazyDiffFile(ref(null), {
        path: computed(() => "a.ts"),
        shouldDefer: ref(false),
      });

      expect(rendered.value).toBe(true);
    });

    scope.stop();
  });

  test("defers until force-rendered", async () => {
    const scope = effectScope();

    scope.run(() => {
      const { rendered } = useLazyDiffFile(ref(null), {
        path: computed(() => "big.ts"),
        shouldDefer: ref(true),
      });

      expect(rendered.value).toBe(false);
      forceRenderDiffFile("big.ts");
      expect(rendered.value).toBe(true);
    });

    await nextTick();

    scope.stop();
  });
});
