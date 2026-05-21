import { ref, type Ref } from "vue";
import type { PendingComment, RepositoryState } from "@git-diff/contracts";

// usePendingComments holds the list of draft comments scoped to the active
// repository / context (working tree, commit sha, ...). reload() reaches into
// the bridge with the current RepositoryState; on failure the list is reset
// rather than left stale.
export interface UsePendingComments {
    items: Ref<PendingComment[]>;
    reload: () => Promise<void>;
    clear: () => void;
}

export function usePendingComments(state: Ref<RepositoryState | null>): UsePendingComments {
    const items = ref<PendingComment[]>([]);

    async function reload(): Promise<void> {
        const current = state.value;

        if (!current) {
            return;
        }

        try {
            const response = await window.diffApp.listPendingComments({
                path: current.root,
                kind: current.mode,
                sha: current.commitSha,
            });

            items.value = response.comments;
        } catch {
            items.value = [];
        }
    }

    function clear(): void {
        items.value = [];
    }

    return { items, reload, clear };
}
