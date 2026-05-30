<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from "vue";
import AuthGate from "@entry/components/auth/AuthGate.vue";
import TitleBar from "@entry/components/diff/TitleBar.vue";
import TopBar from "@entry/components/diff/TopBar.vue";
import Sidebar from "@entry/components/diff/Sidebar.vue";
import RepoToolbar from "@entry/components/diff/RepoToolbar.vue";
import SearchBar from "@entry/components/diff/SearchBar.vue";
import { ToastViewport } from "@ui/toast";
import DiffList from "@entry/components/diff/DiffList.vue";
import RepoEmptyStates from "@entry/components/diff/RepoEmptyStates.vue";
import WalkthroughPanel from "@entry/components/diff/WalkthroughPanel.vue";
import JumpNav from "@entry/components/diff/JumpNav.vue";
import StatusBar from "@entry/components/diff/StatusBar.vue";
import ReviewPanel from "@entry/components/diff/ReviewPanel.vue";
import AddCommentDialog from "@entry/components/diff/AddCommentDialog.vue";
import { sanitizeHtml } from "@ui/safe-html";
import type { RichTextFeatures } from "@ui/rich-text-editor";
import type {
    AuthLoginResponse,
    AuthUser,
    ChangedFile,
    DiffSection,
    DiffViewMode,
    FileSearchResult,
    PullRequestSummary,
    Repository,
    RepositoryFile,
    RepositoryState,
    ReviewComment,
    ReviewDetail,
    ReviewSession,
} from "@git-diff/contracts";
import { PREF_KEYS, viewedPrefKey } from "@git-diff/contracts";
import { ensureLanguage, languageFor } from "@lib/highlight";
import type { LineSelectionRange } from "@composables/useLineSelection";
import { ACCENTS, resolveAccent } from "@lib/accent";
import type { PatchLine } from "@lib/patch";
import { TWEAK_DEFAULTS, tweakPrefPatch, useTweaks, type Tweaks } from "@composables/useTweaks";
import { useToasts } from "@composables/useToasts";
import { useStyleWatchers } from "@composables/useStyleWatchers";
import { useDiffNavigation } from "@composables/useDiffNavigation";
import { resetLazyRender } from "@composables/useLazyRender";
import { useKeyboardShortcuts } from "@composables/useKeyboardShortcuts";
import { useWalkthrough } from "@composables/useWalkthrough";
import { useCommits } from "@composables/useCommits";
import { usePullRequests } from "@composables/usePullRequests";
import { useRepositoryList } from "@composables/useRepositoryList";
import { useCommentDraft } from "@composables/useCommentDraft";
import { usePendingComments } from "@composables/usePendingComments";
import { useSelectedFile } from "@composables/useSelectedFile";
import { usePreferences } from "@composables/usePreferences";
import { useDiffLayout } from "@composables/useDiffLayout";
import { useReviewMarkdownCopy } from "@composables/useReviewMarkdownCopy";
import { useCommandRegistry } from "@composables/useCommandRegistry";
import CommandPalette from "@entry/components/CommandPalette.vue";
import { storeToRefs } from "pinia";
import { useAuthStore } from "@/stores/auth.store";
import { useRepoStore } from "@/stores/repo.store";
import { useReviewsStore } from "@/stores/reviews.store";
import { parseBridgeError } from "@lib/bridgeError";

const commentFeatures: RichTextFeatures = {
    checklist: false,
    images: false,
    markdownShortcuts: true,
    mentions: false,
    slashMenu: false,
    tables: false,
};

const authStore = useAuthStore();
const { mode: authMode, osUsername: authOSUsername, currentUser } = storeToRefs(authStore);

const repoStore = useRepoStore();
const { state, activeRepoPath, loading, error } = storeToRefs(repoStore);

const reviewsStore = useReviewsStore();
const { items: reviews, active: activeReview, summaryDraft } = storeToRefs(reviewsStore);

const {
    values: prefValues,
    load: loadPreferences,
    save: savePreferences,
    reset: resetPreferences,
} = usePreferences({
    onSaveError: (cause) => {
        error.value = cause instanceof Error ? cause.message : String(cause);
    },
});

const prefs = computed(() => prefValues.value);
const tweaks = useTweaks(prefs);

const accent = computed(() => resolveAccent(prefValues.value[PREF_KEYS.uiAccent]));
const diffViewMode = computed<DiffViewMode>(() => tweaks.value.viewMode);
const hideWhitespace = computed(() => prefValues.value[PREF_KEYS.diffHideWhitespace] === "1");
const lastRepoRoot = computed(() => prefValues.value[PREF_KEYS.lastRepoRoot] ?? "");

const {
    items: repositories,
    loading: repositoriesLoading,
    refresh: refreshRepositoryList,
    remove: removeRepositoryFromList,
} = useRepositoryList();
const selectedPath = ref("");
const searchQuery = ref("");
const { collapsed, splitRatios, previewing, toggleCollapsed, setSplitRatio, togglePreview } =
    useDiffLayout();
const reviewPanelOpen = ref(false);
const {
    target: commentTarget,
    draft: commentDraft,
    open: commentDialogOpen,
    begin: beginCommentDraft,
    cancel: cancelCommentDraft,
    close: closeCommentDraft,
} = useCommentDraft();
const {
    file: selectedRepoFile,
    loading: selectedFileLoading,
    error: selectedFileError,
    load: loadSelectedFile,
    reset: resetSelectedFile,
} = useSelectedFile({ selectedPath });
const repoScope = ref<"changed" | "all">("changed");
const { items: commits, loading: commitsLoading, load: loadCommits } = useCommits(state);
const repoMode = computed<"working" | "commit">(() => state.value?.mode ?? "working");
const searchOpen = ref(false);

const {
    items: pullRequests,
    loading: pullRequestsLoading,
    active: activePullRequest,
    load: loadPullRequests,
} = usePullRequests({
    state,
    onLoadError: (cause) => {
        showToast({
            tone: "error",
            title: "Could not load pull requests",
            description: cause instanceof Error ? cause.message : String(cause),
        });
    },
});

const { toasts, show: showToast, dismiss: dismissToast } = useToasts();

const {
    record: walkthrough,
    loading: walkthroughLoading,
    error: walkthroughError,
    generate: generateWalkthrough,
} = useWalkthrough(state);

const files = computed(() => state.value?.files ?? []);

// Reset viewport-deferred render bookkeeping whenever the diff context changes
// so a previous search/navigation "render all" doesn't defeat lazy loading.
watch(
    () => [state.value?.root, state.value?.commitSha, state.value?.mode],
    () => resetLazyRender(),
);

const changedByPath = computed(() => {
    const map = new Map<string, ChangedFile>();

    for (const file of files.value) {
        map.set(file.path, file);
    }

    return map;
});
const changedPathsSet = computed(() => new Set(changedByPath.value.keys()));
const trackedFiles = computed(() => state.value?.trackedFiles ?? []);
const repoPaths = computed(() => {
    const set = new Set<string>(trackedFiles.value);

    for (const file of files.value) {
        set.add(file.path);
    }

    return Array.from(set).sort();
});
const selectedFile = computed<ChangedFile | null>(
    () => changedByPath.value.get(selectedPath.value) ?? files.value[0] ?? null,
);
const selectedIsChanged = computed(
    () => !!selectedPath.value && changedByPath.value.has(selectedPath.value),
);
const reviewComments = computed<ReviewComment[]>(() => activeReview.value?.comments ?? []);
const threadsByPath = computed(() => {
    const counts = new Map<string, number>();

    for (const comment of reviewComments.value) {
        counts.set(comment.filePath, (counts.get(comment.filePath) ?? 0) + 1);
    }

    return counts;
});

const userInitials = computed(() => {
    const name = currentUser.value?.displayName ?? currentUser.value?.osUsername ?? "GO";

    return (
        name
            .split(/[\s_-]+/)
            .map((part) => part[0]?.toUpperCase() ?? "")
            .slice(0, 2)
            .join("") || "GO"
    );
});

const changedIndex = computed(() => files.value.findIndex((f) => f.path === selectedPath.value));

watch(
    files,
    (list) => {
        const seen = new Set<string>();

        for (const file of list) {
            const lang = languageFor(file.path);

            if (lang && !seen.has(lang)) {
                seen.add(lang);
                void ensureLanguage(lang);
            }
        }
    },
    { immediate: true },
);

useStyleWatchers(accent, tweaks);

const { selectAdjacent, jumpToHunk } = useDiffNavigation({
    files,
    changedIndex,
    onSelect: (path) => selectFile(path),
});

{
    const palette = useCommandRegistry();

    palette.register({
        id: "diff.toggle-whitespace",
        title: "Toggle whitespace-only changes",
        section: "Diff",
        keymapId: "toggle_whitespace",
        run: () =>
            savePreferences({ [PREF_KEYS.diffHideWhitespace]: hideWhitespace.value ? "0" : "1" }),
    });

    palette.register({
        id: "review.copy-markdown",
        title: "Copy active review as Markdown",
        section: "Review",
        run: () => copyReviewAsMarkdown(),
    });

    palette.register({
        id: "walkthrough.generate",
        title: "Generate AI walkthrough",
        section: "Review",
        run: async () => {
            await generateWalkthrough();
        },
    });

    palette.register({
        id: "file.copy-path",
        title: "Copy current file path",
        section: "File",
        run: () => {
            const file = selectedFile.value;

            if (file) {
                copyPath(file.path);
            }
        },
    });
}

const shortcutsEnabled = computed(() => authMode.value === "ready");
let unsubscribeShortcuts: (() => void) | null = null;
let unsubscribeLaunchIntent: (() => void) | null = null;

onMounted(async () => {
    await bootstrapAuth();
    unsubscribeShortcuts = useKeyboardShortcuts({
        enabled: shortcutsEnabled,
        onSelectAdjacent: selectAdjacent,
        onJumpToHunk: jumpToHunk,
        onToggleViewed: () => {
            const file = selectedFile.value;

            if (file && changedByPath.value.has(file.path)) {
                void toggleViewed(file);
            }
        },
        onStartReview: () => void startReview(),
        onOpenSearch: () => {
            searchOpen.value = true;
        },
    });
    unsubscribeLaunchIntent = window.diffApp.onLaunchIntent(async (intent) => {
        if (intent.kind === "help" || !intent.repoPath) {
            return;
        }

        await openRepo(intent.repoPath);
        const prNumber = intent.pullRequestNumber ?? intent.prNumber;
        const commitRef = intent.commitRef ?? intent.sha;

        if (intent.kind === "pull-request" && prNumber && state.value) {
            await openPullRequest(prNumber);
        } else if (commitRef && state.value) {
            await openCommit(commitRef);
        }
    });
});

onUnmounted(() => {
    unsubscribeShortcuts?.();
    unsubscribeLaunchIntent?.();
});

async function bootstrapAuth() {
    const { entered } = await authStore.bootstrap();

    if (authStore.bootstrapError) {
        error.value = authStore.bootstrapError;
    }

    if (entered) {
        await enterApp();
    }
}

async function enterApp() {
    await loadPreferences();
    await refreshRepositoryList();
    await applyLaunchIntent();
}

async function applyLaunchIntent() {
    let intent = null;

    try {
        intent = await window.diffApp.takeLaunchIntent();
    } catch {
        // No CLI in this build (browser fallback) — fall through to last-repo path.
    }

    const initialPath =
        intent?.repoPath ??
        repositories.value.find((repo) => repo.path === lastRepoRoot.value)?.path ??
        repositories.value[0]?.path ??
        "";

    if (!initialPath) {
        return;
    }

    await openRepo(initialPath);

    const prNumber = intent?.pullRequestNumber ?? intent?.prNumber;
    const commitRef = intent?.commitRef ?? intent?.sha;

    if (intent?.kind === "pull-request" && prNumber && state.value) {
        await openPullRequest(prNumber);
    } else if (commitRef && state.value) {
        await openCommit(commitRef);
    }

    if (intent?.walkthrough) {
        await generateWalkthrough();
    }
}

async function handleAuthCompleted(response: AuthLoginResponse) {
    authStore.complete(response);
    await enterApp();
}

async function handleAuthWiped() {
    authStore.markWiped();
}

async function logOut() {
    await authStore.logout();
    resetPreferences();
    state.value = null;
    repositories.value = [];
    activeRepoPath.value = "";
    reviews.value = [];
    activeReview.value = null;
    selectedPath.value = "";
    resetSelectedFile();
    await bootstrapAuth();
}

async function openRepo(path: string) {
    activeRepoPath.value = path;

    await repoStore.withBusy(async () => {
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
            error.value = cause instanceof Error ? cause.message : String(cause);
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

    await repoStore.withBusy(async () => {
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
            error.value = cause instanceof Error ? cause.message : String(cause);
        }
    });
}

async function openCommit(sha: string) {
    if (!state.value) {
        return;
    }

    await repoStore.withBusy(async () => {
        try {
            state.value = await window.diffApp.readCommit(sha, state.value!.root);
            selectedPath.value = state.value.files[0]?.path ?? "";
            activeReview.value = null;
            activePullRequest.value = null;
        } catch (cause) {
            error.value = cause instanceof Error ? cause.message : String(cause);
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

    await repoStore.withBusy(async () => {
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
            error.value = cause instanceof Error ? cause.message : String(cause);
        }
    });
}

const creatingBranch = ref(false);
const branchCreateError = ref("");

async function switchBranch(branch: string) {
    if (!state.value) {
        return;
    }

    await repoStore.withBusy(async () => {
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

            error.value =
                structured?.message ?? (cause instanceof Error ? cause.message : String(cause));
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

async function startReview() {
    if (!state.value) {
        return;
    }

    const ctx = state.value;
    const review = await window.diffApp.createReview({
        repoRoot: ctx.root,
        branch: ctx.branch,
        headSha: ctx.headSha,
        title:
            ctx.mode === "commit"
                ? `Review commit ${ctx.branch || ctx.commitSha?.slice(0, 7) || ctx.headSha}`
                : `Review ${ctx.branch || ctx.headSha || "local changes"}`,
        summary: sanitizeHtml(summaryDraft.value),
        filesChanged: ctx.files.length,
        additions: ctx.additions,
        deletions: ctx.deletions,
        contextKind: ctx.mode,
        contextSha: ctx.commitSha,
    });
    // Promote any drafts that were taken before the review session existed.
    try {
        await window.diffApp.promotePendingComments(review.id);
    } catch {
        // Promotion failure shouldn't abort review start.
    }

    clearPendingComments();

    activeReview.value = await window.diffApp.reviewDetail(review.id);
    reviews.value = [review, ...reviews.value.filter((item) => item.id !== review.id)];
    summaryDraft.value = "";
    reviewPanelOpen.value = false;
}

async function toggleViewed(file: ChangedFile) {
    if (!state.value) {
        return;
    }

    const key = viewedPrefKey(state.value.root, file.path);
    const currentlyViewed = isViewed(file);

    await savePreferences({ [key]: currentlyViewed ? "" : file.fingerprint });

    if (activeReview.value) {
        await window.diffApp.addReviewEvent({
            reviewId: activeReview.value.review.id,
            type: currentlyViewed ? "file_unviewed" : "file_viewed",
            filePath: file.path,
            message: currentlyViewed ? "Marked unviewed" : "Marked viewed",
        });
        activeReview.value = await window.diffApp.reviewDetail(activeReview.value.review.id);
    }
}

function isViewed(file: ChangedFile): boolean {
    if (!state.value) {
        return false;
    }

    return prefValues.value[viewedPrefKey(state.value.root, file.path)] === file.fingerprint;
}

function threadsForFile(path: string): number {
    return threadsByPath.value.get(path) ?? 0;
}

async function openSearchResult(result: FileSearchResult) {
    if (result.repoPath && result.repoPath !== activeRepoPath.value) {
        await openRepo(result.repoPath);
    }

    selectFile(result.filePath);
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

function openCommentForLine(
    file: ChangedFile,
    section: DiffSection,
    line: PatchLine,
    range?: LineSelectionRange,
) {
    const lineNumber = line.newLine ?? line.oldLine;

    if (!lineNumber || !activeReview.value) {
        reviewPanelOpen.value = true;

        return;
    }

    if (range) {
        beginCommentDraft({
            file,
            section,
            line: range.endLine,
            side: range.endSide,
            startLine: range.startLine,
            startSide: range.startSide,
        });

        return;
    }

    beginCommentDraft({
        file,
        section,
        line: lineNumber,
        side: line.newLine ? "right" : "left",
    });
}

async function saveComment() {
    if (!commentTarget.value || !commentDraft.value.trim() || !state.value) {
        return;
    }

    const target = commentTarget.value;
    const author = currentUser.value?.displayName || currentUser.value?.osUsername || "You";
    const body = sanitizeHtml(commentDraft.value);

    if (activeReview.value) {
        await window.diffApp.createReviewComment({
            reviewId: activeReview.value.review.id,
            filePath: target.file.path,
            diffSection: target.section.kind,
            side: target.side,
            lineNumber: target.line,
            startLineNumber: target.startLine,
            startSide: target.startSide,
            authorLabel: author,
            bodyHtml: body,
        });
        activeReview.value = await window.diffApp.reviewDetail(activeReview.value.review.id);
    } else {
        // No active review yet — store as a draft. On startReview, draft comments
        // for this (repo, context) are promoted into the new review session.
        await window.diffApp.createPendingComment({
            repoRoot: state.value.root,
            contextKind: state.value.mode,
            contextSha: state.value.commitSha,
            filePath: target.file.path,
            diffSection: target.section.kind,
            side: target.side,
            lineNumber: target.line,
            startLineNumber: target.startLine,
            startSide: target.startSide,
            authorLabel: author,
            bodyHtml: body,
        });
        await loadPendingComments();
    }

    closeCommentDraft();
}

const {
    items: pendingComments,
    reload: loadPendingComments,
    clear: clearPendingComments,
} = usePendingComments(state);

function cancelComment() {
    cancelCommentDraft();
}

async function deleteComment(comment: ReviewComment) {
    if (!activeReview.value) {
        return;
    }

    await window.diffApp.deleteReviewComment({
        reviewId: activeReview.value.review.id,
        commentId: comment.id,
    });
    activeReview.value = await window.diffApp.reviewDetail(activeReview.value.review.id);
}

async function resolveComment(comment: ReviewComment, resolved: boolean) {
    if (!activeReview.value) {
        return;
    }

    await window.diffApp.setReviewCommentResolved({
        reviewId: activeReview.value.review.id,
        commentId: comment.id,
        resolved,
    });
    activeReview.value = await window.diffApp.reviewDetail(activeReview.value.review.id);
}

async function replyToComment(parent: ReviewComment, bodyHtml: string) {
    if (!activeReview.value) {
        return;
    }

    await window.diffApp.createReviewComment({
        reviewId: activeReview.value.review.id,
        filePath: parent.filePath,
        diffSection: parent.diffSection,
        side: parent.side,
        lineNumber: parent.lineNumber,
        authorLabel: currentUser.value?.displayName || currentUser.value?.osUsername || "You",
        bodyHtml: sanitizeHtml(bodyHtml),
    });
    activeReview.value = await window.diffApp.reviewDetail(activeReview.value.review.id);
}

async function updateTweak<K extends keyof Tweaks>(key: K, value: Tweaks[K]) {
    await savePreferences(tweakPrefPatch(key, value));
}

function fileElementID(path: string): string {
    return `file-${path.replace(/[^a-z0-9_-]/gi, "-")}`;
}

function copyPath(path: string) {
    if (path) {
        void navigator.clipboard.writeText(path);
    }
}

const { state: copyReviewState, copy: copyActiveReviewAsMarkdown } = useReviewMarkdownCopy({
    onError: (cause) => {
        error.value = cause instanceof Error ? cause.message : String(cause);
    },
});

async function copyReviewAsMarkdown() {
    if (activeReview.value) {
        await copyActiveReviewAsMarkdown(activeReview.value);
    }
}

void TWEAK_DEFAULTS;
void ACCENTS;
</script>

<template>
    <AuthGate @entered="handleAuthCompleted" @wiped="handleAuthWiped">
        <div
            class="flex h-screen flex-col overflow-hidden"
            :style="{
                background: 'var(--gd-bloom), var(--gd-bg)',
                color: 'var(--gd-text)',
            }"
        >
            <TitleBar
                :state="state"
                :repositories="repositories"
                :repositories-loading="repositoriesLoading"
                :active-repo-path="activeRepoPath"
                @select-repo="openRepo"
                @add-repo="addRepository"
                @remove-repo="removeRepository"
                @refresh-repos="refreshRepositoryList"
            />
            <TopBar
                :state="state"
                :current-user="currentUser"
                :user-initials="userInitials"
                :tweaks="tweaks"
                :creating-branch="creatingBranch"
                :branch-create-error="branchCreateError"
                @update:tweak="updateTweak"
                @refresh="refresh"
                @log-out="logOut"
                @switch-branch="switchBranch"
                @create-branch="createBranch"
                @select-result="openSearchResult"
            />

            <RepoToolbar
                v-if="state"
                :state="state"
                :commits="commits"
                :commits-loading="commitsLoading"
                :repo-mode="repoMode"
                :pull-requests="pullRequests"
                :pull-requests-loading="pullRequestsLoading"
                :active-pull-request="activePullRequest"
                :walkthrough-loading="walkthroughLoading"
                :has-walkthrough="walkthrough != null"
                :has-active-review="activeReview != null"
                :copy-review-state="copyReviewState"
                @load-commits="loadCommits()"
                @open-commit="openCommit"
                @back-to-working="returnToWorkingTree"
                @load-pull-requests="loadPullRequests()"
                @open-pull-request="openPullRequest"
                @generate-walkthrough="generateWalkthrough(walkthrough != null)"
                @copy-review-as-markdown="copyReviewAsMarkdown"
            />
            <WalkthroughPanel :record="walkthrough" :error="walkthroughError" />

            <main class="flex flex-1 min-h-0">
                <template v-if="!activeRepoPath || loading || error || !state">
                    <RepoEmptyStates
                        :kind="!activeRepoPath ? 'hero' : loading ? 'loading' : 'error'"
                        :last-repo-root="lastRepoRoot"
                        :error="error"
                        @add-repository="addRepository"
                        @open-last-repo="openRepo(lastRepoRoot)"
                    />
                </template>
                <template v-else>
                    <Sidebar
                        :files="files"
                        :selected-path="selectedPath"
                        :search-query="searchQuery"
                        :scope="repoScope"
                        :is-viewed="isViewed"
                        :threads-for-file="threadsForFile"
                        :all-paths="repoPaths"
                        :changed-paths-set="changedPathsSet"
                        :comments="reviewComments"
                        @update:scope="(value) => (repoScope = value)"
                        @update:search-query="(value) => (searchQuery = value)"
                        @select="selectFile"
                        @toggle-viewed="toggleViewed"
                        @start-review="startReview"
                    />

                    <section class="flex flex-col flex-1 min-w-0 relative">
                        <div class="flex-1 overflow-auto min-h-0">
                            <DiffList
                                :files="files"
                                :selected-path="selectedPath"
                                :selected-is-changed="selectedIsChanged"
                                :selected-repo-file="selectedRepoFile"
                                :selected-file-loading="selectedFileLoading"
                                :selected-file-error="selectedFileError"
                                :collapsed="collapsed"
                                :split-ratios="splitRatios"
                                :tweaks="tweaks"
                                :diff-view-mode="diffViewMode"
                                :hide-whitespace="hideWhitespace"
                                :hide-resolved="tweaks.hideResolved"
                                :review-comments="reviewComments"
                                :comment-features="commentFeatures"
                                :repo-root="state?.root ?? ''"
                                :commit-ref="state?.commitSha"
                                :previewing="previewing"
                                :is-viewed-fn="isViewed"
                                :file-element-i-d="fileElementID"
                                @toggle-collapsed="toggleCollapsed"
                                @toggle-viewed="toggleViewed"
                                @toggle-preview="togglePreview"
                                @copy-path="copyPath"
                                @open-comment-for-line="openCommentForLine"
                                @delete-comment="deleteComment"
                                @reply-comment="replyToComment"
                                @resolve-comment="resolveComment"
                                @update:split-ratio="setSplitRatio"
                            />
                        </div>

                        <JumpNav
                            v-if="tweaks.showMinimap && files.length > 0"
                            :index="changedIndex"
                            :total="files.length"
                            @prev="selectAdjacent(-1)"
                            @next="selectAdjacent(1)"
                        />
                    </section>

                    <ReviewPanel
                        :open="reviewPanelOpen"
                        :summary-draft="summaryDraft"
                        :features="commentFeatures"
                        @close="reviewPanelOpen = false"
                        @update:summary-draft="(value) => (summaryDraft = value)"
                        @start-review="startReview"
                    />
                </template>
            </main>

            <StatusBar
                v-if="tweaks.showStatusBar"
                :file-path="selectedFile?.path ?? ''"
                :viewed-count="files.filter((f) => isViewed(f)).length"
                :total="files.length"
            />

            <AddCommentDialog
                :open="commentDialogOpen"
                :file-path="commentTarget?.file.path"
                :line-number="commentTarget?.line"
                :model-value="commentDraft"
                :features="commentFeatures"
                @update:model-value="(value) => (commentDraft = value)"
                @save="saveComment"
                @cancel="cancelComment"
            />

            <SearchBar :open="searchOpen" @close="searchOpen = false" />

            <CommandPalette />

            <ToastViewport :toasts="toasts" @dismiss="dismissToast" />
        </div>
    </AuthGate>
</template>
