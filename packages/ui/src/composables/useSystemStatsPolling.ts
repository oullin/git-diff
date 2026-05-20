import { onBeforeUnmount, onMounted, ref, type Ref } from "vue";
import type { SystemStats } from "@git-diff/contracts";

// useSystemStatsPolling refreshes the systemStats ref on a fixed interval and
// pauses polling when the document is hidden so background tabs do not keep
// the timer alive. Errors are intentionally swallowed: the panel keeps the
// last good sample on transient failures.
export interface UseSystemStatsPolling {
  stats: Ref<SystemStats | null>;
}

export function useSystemStatsPolling(intervalMs = 5000): UseSystemStatsPolling {
  const stats = ref<SystemStats | null>(null);
  let timer: ReturnType<typeof setInterval> | null = null;

  async function refresh(): Promise<void> {
    try {
      stats.value = await window.diffApp.getSystemStats();
    } catch {
      // Non-critical — leave the previous value in place.
    }
  }

  function start(): void {
    if (timer) return;

    refresh();
    timer = setInterval(refresh, intervalMs);
  }

  function stop(): void {
    if (!timer) return;

    clearInterval(timer);
    timer = null;
  }

  function handleVisibilityChange(): void {
    if (document.hidden) {
      stop();
    } else {
      start();
    }
  }

  onMounted(() => {
    if (!document.hidden) start();
    document.addEventListener("visibilitychange", handleVisibilityChange);
  });

  onBeforeUnmount(() => {
    document.removeEventListener("visibilitychange", handleVisibilityChange);
    stop();
  });

  return { stats };
}
