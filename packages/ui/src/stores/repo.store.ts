import { defineStore } from "pinia";
import { ref } from "vue";
import type { RepositoryState } from "@git-diff/domain";

// useRepoStore is the canonical home for the active RepositoryState, the
// path of the repo the user picked (which can lag behind state.root when a
// load is in flight), and the loading/error pair that the busy UI watches.
// Cross-domain orchestration (e.g. openRepo seeding selectedPath, reviews,
// preferences) still belongs to App.vue because it spans several stores.
export const useRepoStore = defineStore("repo", () => {
    const state = ref<RepositoryState | null>(null);
    const activeRepoPath = ref("");
    const loading = ref(false);
    const error = ref("");

    function setState(next: RepositoryState | null): void {
        state.value = next;
    }

    function setActive(path: string): void {
        activeRepoPath.value = path;
    }

    function setLoading(value: boolean): void {
        loading.value = value;
    }

    function setError(value: string): void {
        error.value = value;
    }

    function clear(): void {
        state.value = null;
        activeRepoPath.value = "";
        error.value = "";
    }

    // withBusy runs an async action with the loading flag pinned to true and
    // resets it on completion regardless of outcome. Errors propagate so
    // callers can decide whether to surface them or just record them on
    // `error.value`.
    async function withBusy<T>(action: () => Promise<T>): Promise<T> {
        loading.value = true;
        error.value = "";

        try {
            return await action();
        } finally {
            loading.value = false;
        }
    }

    return {
        state,
        activeRepoPath,
        loading,
        error,
        setState,
        setActive,
        setLoading,
        setError,
        clear,
        withBusy,
    };
});
