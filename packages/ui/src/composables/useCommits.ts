import { ref, type Ref } from "vue";
import type { CommitSummary, RepositoryState } from "@git-diff/contracts";

// useCommits owns the commit-picker list state for the currently open
// repository. load(limit) reaches into the bridge and fills `items`; opening
// a specific commit still belongs to App.vue (it mutates RepositoryState).
export interface UseCommits {
  items: Ref<CommitSummary[]>;
  loading: Ref<boolean>;
  error: Ref<string>;
  load: (limit?: number) => Promise<void>;
  reset: () => void;
}

export function useCommits(state: Ref<RepositoryState | null>): UseCommits {
  const items = ref<CommitSummary[]>([]);
  const loading = ref(false);
  const error = ref("");

  async function load(limit = 100): Promise<void> {
    const current = state.value;

    if (!current) return;

    loading.value = true;
    error.value = "";

    try {
      const response = await window.diffApp.listCommits(current.root, limit);
      items.value = response.commits;
    } catch (cause) {
      error.value = cause instanceof Error ? cause.message : String(cause);
    } finally {
      loading.value = false;
    }
  }

  function reset(): void {
    items.value = [];
    error.value = "";
  }

  return { items, loading, error, load, reset };
}
