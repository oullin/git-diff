import { ref, type Ref } from "vue";
import type { PullRequestSummary, RepositoryState } from "@git-diff/contracts";

export interface UsePullRequests {
    items: Ref<PullRequestSummary[]>;
    loading: Ref<boolean>;
    active: Ref<PullRequestSummary | null>;
    error: Ref<string>;
    load: () => Promise<void>;
    setActive: (pr: PullRequestSummary | null) => void;
    reset: () => void;
}

export interface UsePullRequestsOptions {
    state: Ref<RepositoryState | null>;
    onLoadError?: (cause: unknown) => void;
}

export function usePullRequests(opts: UsePullRequestsOptions): UsePullRequests {
    const items = ref<PullRequestSummary[]>([]);
    const loading = ref(false);
    const active = ref<PullRequestSummary | null>(null);
    const error = ref("");

    async function load(): Promise<void> {
        const current = opts.state.value;

        if (!current) {
            return;
        }

        loading.value = true;
        error.value = "";

        try {
            const response = await window.diffApp.listPullRequests(current.root);

            items.value = response.pullRequests;
        } catch (cause) {
            items.value = [];
            error.value = cause instanceof Error ? cause.message : String(cause);
            opts.onLoadError?.(cause);
        } finally {
            loading.value = false;
        }
    }

    function setActive(pr: PullRequestSummary | null): void {
        active.value = pr;
    }

    function reset(): void {
        items.value = [];
        active.value = null;
        error.value = "";
    }

    return { items, loading, active, error, load, setActive, reset };
}
