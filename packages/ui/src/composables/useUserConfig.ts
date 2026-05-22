import { ref, type Ref } from "vue";
import type { UserConfig } from "@git-diff/contracts";

// Module-level singleton: every component that calls useUserConfig() sees
// the same reactive ref. The Go backend is the source of truth — we fetch
// once on first use and expose a manual refetch.
//
// Why not SSE here: the renderer talks to the Go backend over the
// preload-injected `window.diffApp` IPC bridge, not over a network socket
// the browser's EventSource can open. Hot-reload from file edits will be
// piped through the Electron main process in a follow-up; the GET path is
// enough for theme, keymap, and walkthrough budgets on app start.

const config = ref<UserConfig | null>(null);
const loading = ref(false);
let inflight: Promise<UserConfig | null> | null = null;

export interface UseUserConfig {
    config: Ref<UserConfig | null>;
    loading: Ref<boolean>;
    /** Force-refetch from the backend. */
    refresh(): Promise<UserConfig | null>;
}

/**
 * Returns the shared reactive user config. The first caller triggers a
 * fetch; subsequent callers see the result reactively. Safe to call from
 * any component / composable.
 */
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
            // Failure is non-fatal — defaults applied at composable boundary.
            return null;
        } finally {
            loading.value = false;
            inflight = null;
        }
    })();

    return inflight;
}
