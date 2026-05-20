import { ref, type Ref } from "vue";

// useBranches holds the branch picker's list state for a single repository
// root. Callers invoke load(root) when the dropdown opens (or on demand);
// errors stay surfaced via the error ref so the host UI can render them
// inline without callback indirection.
export interface UseBranches {
  items: Ref<string[]>;
  loading: Ref<boolean>;
  error: Ref<string>;
  load: (root: string) => Promise<void>;
}

export function useBranches(): UseBranches {
  const items = ref<string[]>([]);
  const loading = ref(false);
  const error = ref("");

  async function load(root: string): Promise<void> {
    if (!root) return;

    loading.value = true;
    error.value = "";

    try {
      const result = await window.diffApp.listBranches(root);
      items.value = result.branches;
    } catch (cause) {
      error.value = cause instanceof Error ? cause.message : String(cause);
    } finally {
      loading.value = false;
    }
  }

  return { items, loading, error, load };
}
