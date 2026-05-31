import { ref, type Ref } from "vue";
import type { RepositoryFile } from "@git-diff/domain";

// useSelectedFile manages the file-viewer panel for paths that are NOT in
// the diff (the FileContentViewer fallback). load() races against the latest
// selectedPath so a slow read does not overwrite content for a file the user
// has since navigated away from. reset() drops the cached content without
// touching the loading flag.
export interface UseSelectedFile {
  file: Ref<RepositoryFile | null>;
  loading: Ref<boolean>;
  error: Ref<string>;
  load: (root: string, path: string) => Promise<void>;
  reset: () => void;
}

export interface UseSelectedFileOptions {
  // Reactive pointer to the current selection. The composable only commits
  // load() results when this still matches the path it was given.
  selectedPath: Ref<string>;
}

export function useSelectedFile(opts: UseSelectedFileOptions): UseSelectedFile {
  const file = ref<RepositoryFile | null>(null);

  const loading = ref(false);

  const error = ref("");

  async function load(root: string, path: string): Promise<void> {
    if (!root || !path) {
      file.value = null;

      return;
    }

    loading.value = true;
    error.value = "";
    file.value = null;

    try {
      const loaded = await window.diffApp.readRepositoryFile(root, path);

      if (opts.selectedPath.value === path) {
        file.value = loaded;
      }
    } catch (cause) {
      if (opts.selectedPath.value === path) {
        error.value = cause instanceof Error ? cause.message : String(cause);
      }
    } finally {
      if (opts.selectedPath.value === path) {
        loading.value = false;
      }
    }
  }

  function reset(): void {
    file.value = null;
    error.value = "";
  }

  return { file, loading, error, load, reset };
}
