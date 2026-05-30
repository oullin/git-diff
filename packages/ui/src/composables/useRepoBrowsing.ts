import { nextTick, ref, type ComputedRef, type Ref } from "vue";
import {
    PREF_KEYS,
    type ChangedFile,
    type FileSearchResult,
    type PullRequestSummary,
    type RepositoryState,
    type ReviewDetail,
    type ReviewSession,
} from "@git-diff/domain";
import { parseBridgeError } from "@lib/bridgeError";
import type { ToastItem } from "@ui/toast";

/**
 * Owns repository navigation: opening repos/commits/pull-requests, refresh,
 * branch switch/create, repo add/remove, and file selection. Reseeds the
 * selected path after each context change. Depends on injected store refs and
 * the review/selected-file/pull-request collaborators so the stores stay the
 * single source of truth.
 */
export interface RepoBrowsingOptions {
    state: Ref<RepositoryState | null>;
    activeRepoPath: Ref<string>;
    reviews: Ref<ReviewSession[]>;
    activeReview: Ref<ReviewDetail | null>;
    selectedPath: Ref<string>;
    pullRequests: Ref<PullRequestSummary[]>;
    activePullRequest: Ref<PullRequestSummary | null>;
    changedByPath: ComputedRef<Map<string, ChangedFile>>;
    reviewPanelOpen: Ref<boolean>;
    lastRepoRoot: ComputedRef<string>;
    withBusy: (fn: () => Promise<void>) => Promise<void>;
    savePreferences: (patch: Record<string, string>) => Promise<void>;
    refreshRepositoryList: () => Promise<void>;
    removeRepositoryFromList: (path: string) => Promise<void>;
    loadPendingComments: () => Promise<void>;
    resetSelectedFile: () => void;
    loadSelectedFile: (root: string, path: string) => Promise<void>;
    showToast: (toast: Omit<ToastItem, "id">, ttlMs?: number) => string;
    onError: (message: string) => void;
}

export function useRepoBrowsing(opts: RepoBrowsingOptions) {
    const {
        state,
        activeRepoPath,
        reviews,
        activeReview,
        selectedPath,
        pullRequests,
        activePullRequest,
        changedByPath,
        reviewPanelOpen,
        lastRepoRoot,
        withBusy,
        savePreferences,
        refreshRepositoryList,
        removeRepositoryFromList,
        loadPendingComments,
        resetSelectedFile,
        loadSelectedFile,
        showToast,
        onError,
    } = opts;

    const creatingBranch = ref(false);
    const branchCreateError = ref("");

    function fileElementID(path: string): string {
        return `file-${path.replace(/[^a-z0-9_-]/gi, "-")}`;
    }

    async function openRepo(path: string) {
        activeRepoPath.value = path;

        await withBusy(async () => {
            try {
                const opened = await window.diffApp.repositoryState(path);

                state.value = opened;
                selectedPath.value = opened.files[0]?.path ?? "";
                resetSelectedFile();
                const reviewResponse = await window.diffApp.listReviews(25);

                reviews.value = reviewResponse.reviews.filter(
                    (review) => review.repoRoot === opened.root,
                );
                activeReview.value = reviews.value[0]
                    ? await window.diffApp.reviewDetail(reviews.value[0].id)
                    : null;
                await loadPendingComments();
                await savePreferences({ [PREF_KEYS.lastRepoRoot]: opened.root });
                await window.diffApp.upsertRepository({ path: opened.root });
                await refreshRepositoryList();
                activeRepoPath.value = opened.root;
            } catch (cause) {
                onError(cause instanceof Error ? cause.message : String(cause));
                state.value = null;
            }
        });
    }

    async function refresh() {
        if (!state.value) {
            if (activeRepoPath.value) {
                await openRepo(activeRepoPath.value);
            }

            return;
        }

        await withBusy(async () => {
            try {
                const current = state.value!;
                const next =
                    current.mode === "commit" && current.commitSha
                        ? await window.diffApp.readCommit(current.commitSha, current.root)
                        : await window.diffApp.refreshRepository(current.root);

                state.value = next;

                if (!next.files.some((file) => file.path === selectedPath.value)) {
                    selectedPath.value = next.files[0]?.path ?? "";
                }
            } catch (cause) {
                onError(cause instanceof Error ? cause.message : String(cause));
            }
        });
    }

    async function openCommit(sha: string) {
        if (!state.value) {
            return;
        }

        await withBusy(async () => {
            try {
                state.value = await window.diffApp.readCommit(sha, state.value!.root);
                selectedPath.value = state.value.files[0]?.path ?? "";
                activeReview.value = null;
                activePullRequest.value = null;
            } catch (cause) {
                onError(cause instanceof Error ? cause.message : String(cause));
            }
        });
    }

    async function returnToWorkingTree() {
        if (!state.value) {
            return;
        }

        activePullRequest.value = null;
        await openRepo(state.value.root);
    }

    async function openPullRequest(number: number) {
        if (!state.value) {
            return;
        }

        await withBusy(async () => {
            try {
                const opened = await window.diffApp.readPullRequest(number, state.value!.root);

                state.value = opened;
                selectedPath.value = opened.files[0]?.path ?? "";
                activeReview.value = null;
                activePullRequest.value = pullRequests.value.find((pr) => pr.number === number) ?? {
                    number,
                    title: "",
                    author: "",
                    state: "open",
                    baseRef: "",
                    headRef: opened.branch,
                    url: "",
                };
            } catch (cause) {
                onError(cause instanceof Error ? cause.message : String(cause));
            }
        });
    }

    async function switchBranch(branch: string) {
        if (!state.value) {
            return;
        }

        await withBusy(async () => {
            try {
                const next = await window.diffApp.checkoutBranch(state.value!.root, branch);

                state.value = next;

                if (!next.files.some((file) => file.path === selectedPath.value)) {
                    selectedPath.value = next.files[0]?.path ?? "";
                }
            } catch (cause) {
                const structured = parseBridgeError(cause);

                if (structured?.code === "working_tree_dirty") {
                    const files = structured.files ?? [];

                    showToast(
                        {
                            tone: "error",
                            wide: true,
                            title: "Commit or stash your changes before switching branches",
                            description:
                                files.length > 0
                                    ? `${files.length} file${files.length === 1 ? "" : "s"} would be overwritten by checkout: ${files.join(", ")}`
                                    : undefined,
                        },
                        0,
                    );

                    return;
                }

                onError(
                    structured?.message ?? (cause instanceof Error ? cause.message : String(cause)),
                );
            }
        });
    }

    async function createBranch(name: string) {
        if (!state.value) {
            return;
        }

        creatingBranch.value = true;
        branchCreateError.value = "";
        try {
            state.value = await window.diffApp.createBranch(state.value.root, name);
            if (!state.value.files.some((file) => file.path === selectedPath.value)) {
                selectedPath.value = state.value.files[0]?.path ?? "";
            }
        } catch (cause) {
            branchCreateError.value = cause instanceof Error ? cause.message : String(cause);
        } finally {
            creatingBranch.value = false;
        }
    }

    async function addRepository() {
        const chosen = await window.diffApp.chooseRepository(
            activeRepoPath.value || lastRepoRoot.value,
        );

        if (chosen) {
            await openRepo(chosen);
        }
    }

    async function removeRepository(path: string) {
        await removeRepositoryFromList(path);
        if (activeRepoPath.value === path) {
            activeRepoPath.value = "";
            state.value = null;
            activeReview.value = null;
            reviews.value = [];
            selectedPath.value = "";
            resetSelectedFile();
            reviewPanelOpen.value = false;
        }
    }

    function selectFile(path: string) {
        selectedPath.value = path;
        if (changedByPath.value.has(path)) {
            resetSelectedFile();
            nextTick(() =>
                document.getElementById(fileElementID(path))?.scrollIntoView({ block: "start" }),
            );

            return;
        }

        if (state.value) {
            void loadSelectedFile(state.value.root, path);
        } else {
            resetSelectedFile();
        }
    }

    async function openSearchResult(result: FileSearchResult) {
        if (result.repoPath && result.repoPath !== activeRepoPath.value) {
            await openRepo(result.repoPath);
        }

        selectFile(result.filePath);
    }

    return {
        creatingBranch,
        branchCreateError,
        fileElementID,
        openRepo,
        refresh,
        openCommit,
        returnToWorkingTree,
        openPullRequest,
        switchBranch,
        createBranch,
        addRepository,
        removeRepository,
        selectFile,
        openSearchResult,
    };
}
