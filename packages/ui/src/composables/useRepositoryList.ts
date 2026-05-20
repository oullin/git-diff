import { ref, type Ref } from "vue";
import type { Repository } from "@git-diff/contracts";

// useRepositoryList owns the registry of repositories the renderer knows
// about (whatever the backend's /v1/repositories returns). Adding a repo is
// driven by App.vue because it stitches together chooseRepository +
// openRepo, both of which want broader scope. Removal stays here because
// the operation is bounded to the list itself.
export interface UseRepositoryList {
  items: Ref<Repository[]>;
  loading: Ref<boolean>;
  refresh: () => Promise<void>;
  remove: (path: string) => Promise<void>;
}

export function useRepositoryList(): UseRepositoryList {
  const items = ref<Repository[]>([]);
  const loading = ref(false);

  async function refresh(): Promise<void> {
    loading.value = true;

    try {
      items.value = await window.diffApp.listRepositories();
    } finally {
      loading.value = false;
    }
  }

  async function remove(path: string): Promise<void> {
    await window.diffApp.removeRepository(path);
    await refresh();
  }

  return { items, loading, refresh, remove };
}
