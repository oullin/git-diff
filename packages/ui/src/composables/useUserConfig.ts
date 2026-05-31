import { ref, type Ref } from "vue";
import type { UserConfig } from "@git-diff/domain";

// Module-level singleton: the Go backend is the source of truth; we
// fetch once on first use and expose a manual refetch. SSE isn't wired
// because the IPC bridge isn't an EventSource transport.

const config = ref<UserConfig | null>(null);

const loading = ref(false);

let inflight: Promise<UserConfig | null> | null = null;

export interface UseUserConfig {
  config: Ref<UserConfig | null>;
  loading: Ref<boolean>;
  /** Force-refetch from the backend. */
  refresh(): Promise<UserConfig | null>;
}

export function useUserConfig(): UseUserConfig {
  if (config.value === null && !inflight) {
    void load();
  }

  return {
    config,
    loading,
    refresh: load,
  };
}

async function load(): Promise<UserConfig | null> {
  if (inflight) {
    return inflight;
  }

  loading.value = true;

  inflight = (async () => {
    try {
      const value = await window.diffApp.getUserConfig();

      config.value = value;

      return value;
    } catch {
      // Defaults are applied at the composable boundary on null.
      return null;
    } finally {
      loading.value = false;
      inflight = null;
    }
  })();

  return inflight;
}
