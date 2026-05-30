// @vitest-environment happy-dom
import { describe, expect, test } from "vitest";
import { ref } from "vue";
import { useRepoSelectors } from "@composables/useRepoSelectors";
import type { AuthUser, ChangedFile, RepositoryState, ReviewDetail } from "@git-diff/contracts";

function changed(path: string): ChangedFile {
    return { path, fingerprint: `fp:${path}` } as ChangedFile;
}

function state(files: ChangedFile[], trackedFiles: string[] = []): RepositoryState {
    return { root: "/repo", files, trackedFiles } as RepositoryState;
}

function setup(overrides?: {
    state?: RepositoryState | null;
    selectedPath?: string;
    activeReview?: ReviewDetail | null;
    currentUser?: AuthUser | null;
}) {
    return useRepoSelectors({
        state: ref(overrides?.state ?? null),
        selectedPath: ref(overrides?.selectedPath ?? ""),
        activeReview: ref(overrides?.activeReview ?? null),
        currentUser: ref(overrides?.currentUser ?? null),
    });
}

describe("useRepoSelectors", () => {
    test("files falls back to an empty list without state", () => {
        expect(setup().files.value).toEqual([]);
    });

    test("changedByPath and changedPathsSet index the changed files", () => {
        const s = setup({ state: state([changed("a.ts"), changed("b.ts")]) });

        expect([...s.changedByPath.value.keys()]).toEqual(["a.ts", "b.ts"]);
        expect(s.changedPathsSet.value.has("a.ts")).toBe(true);
    });

    test("repoPaths unions tracked + changed files, sorted and deduped", () => {
        const s = setup({ state: state([changed("b.ts")], ["a.ts", "b.ts"]) });

        expect(s.repoPaths.value).toEqual(["a.ts", "b.ts"]);
    });

    test("selectedFile prefers the selected path, else the first file", () => {
        const files = [changed("a.ts"), changed("b.ts")];

        expect(setup({ state: state(files), selectedPath: "b.ts" }).selectedFile.value?.path).toBe(
            "b.ts",
        );
        expect(setup({ state: state(files), selectedPath: "" }).selectedFile.value?.path).toBe(
            "a.ts",
        );
    });

    test("selectedIsChanged reflects whether the path is a changed file", () => {
        const s = setup({ state: state([changed("a.ts")]), selectedPath: "a.ts" });

        expect(s.selectedIsChanged.value).toBe(true);
        expect(setup({ selectedPath: "x.ts" }).selectedIsChanged.value).toBe(false);
    });

    test("threadsByPath / threadsForFile count comments per file", () => {
        const review = {
            comments: [{ filePath: "a.ts" }, { filePath: "a.ts" }, { filePath: "b.ts" }],
        } as ReviewDetail;
        const s = setup({ activeReview: review });

        expect(s.threadsForFile("a.ts")).toBe(2);
        expect(s.threadsForFile("b.ts")).toBe(1);
        expect(s.threadsForFile("missing.ts")).toBe(0);
    });

    test("userInitials derives up to two uppercase initials", () => {
        expect(
            setup({ currentUser: { displayName: "Ada Lovelace" } as AuthUser }).userInitials.value,
        ).toBe("AL");
        // A null user falls back to the single-word "GO" placeholder -> "G".
        expect(setup({ currentUser: null }).userInitials.value).toBe("G");
    });

    test("changedIndex locates the selected file", () => {
        const s = setup({ state: state([changed("a.ts"), changed("b.ts")]), selectedPath: "b.ts" });

        expect(s.changedIndex.value).toBe(1);
    });
});
