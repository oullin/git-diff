import { computed, ref, type ComputedRef, type Ref } from "vue";
import type { FileSearchResult } from "@git-diff/contracts";

// useFileSearch owns the debounced search-bar lifecycle: query string,
// in-flight token (so a slow response can't overwrite results from a newer
// query), and the result list. The composable wires its own setTimeout-based
// debouncer; callers just forward the input event to onInput().
export interface UseFileSearch {
  query: Ref<string>;
  results: Ref<FileSearchResult[]>;
  loading: Ref<boolean>;
  error: Ref<string>;
  open: ComputedRef<boolean>;
  onInput: (value: string) => void;
  reset: () => void;
}

export interface UseFileSearchOptions {
  debounceMs?: number;
  limit?: number;
}

export function useFileSearch(opts: UseFileSearchOptions = {}): UseFileSearch {
  const debounceMs = opts.debounceMs ?? 150;
  const limit = opts.limit ?? 50;

  const query = ref("");
  const results = ref<FileSearchResult[]>([]);
  const loading = ref(false);
  const error = ref("");
  const open = computed(() => query.value.trim().length > 0);

  let debounce: ReturnType<typeof setTimeout> | null = null;
  let token = 0;

  async function run(input: string): Promise<void> {
    const trimmed = input.trim();

    if (!trimmed) {
      results.value = [];
      loading.value = false;
      error.value = "";

      return;
    }

    const current = ++token;

    loading.value = true;
    error.value = "";

    try {
      const found = await window.diffApp.searchRepositoryFiles(trimmed, limit);

      if (current !== token) return;

      results.value = found;
    } catch (cause) {
      if (current !== token) return;

      error.value = cause instanceof Error ? cause.message : String(cause);
      results.value = [];
    } finally {
      if (current === token) {
        loading.value = false;
      }
    }
  }

  function onInput(value: string): void {
    query.value = value;

    if (debounce) {
      clearTimeout(debounce);
    }

    if (!value.trim()) {
      token++;
      results.value = [];
      loading.value = false;
      error.value = "";

      return;
    }

    debounce = setTimeout(() => {
      void run(value);
    }, debounceMs);
  }

  function reset(): void {
    if (debounce) {
      clearTimeout(debounce);
      debounce = null;
    }

    token++;
    query.value = "";
    results.value = [];
    loading.value = false;
    error.value = "";
  }

  return { query, results, loading, error, open, onInput, reset };
}
