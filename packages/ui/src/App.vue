<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from "vue";
import { GitPullRequest, FolderOpen, Plus } from "lucide-vue-next";
import AuthGate from "@entry/components/auth/AuthGate.vue";
import FileContentViewer from "@entry/components/FileContentViewer.vue";
import TitleBar from "@entry/components/diff/TitleBar.vue";
import TopBar from "@entry/components/diff/TopBar.vue";
import Sidebar from "@entry/components/diff/Sidebar.vue";
import CommitPicker from "@entry/components/commits/CommitPicker.vue";
import PullRequestPicker from "@entry/components/commits/PullRequestPicker.vue";
import SearchBar from "@entry/components/diff/SearchBar.vue";
import { ToastViewport } from "@ui/toast";
import FileHeader from "@entry/components/diff/FileHeader.vue";
import DiffBody from "@entry/components/diff/DiffBody.vue";
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
import { ACCENTS, resolveAccent } from "@lib/accent";
import type { PatchLine } from "@lib/patch";
import { formatReviewAsMarkdown } from "@lib/reviewMarkdown";
import { TWEAK_DEFAULTS, tweakPrefPatch, useTweaks, type Tweaks } from "@composables/useTweaks";
import { useToasts } from "@composables/useToasts";
import { useStyleWatchers } from "@composables/useStyleWatchers";
import {
  useDiffNavigation,
  useKeyboardShortcuts,
} from "@composables/useDiffNavigation";
import { useWalkthrough } from "@composables/useWalkthrough";
import { useCommits } from "@composables/useCommits";
import { usePullRequests } from "@composables/usePullRequests";
import { useRepositoryList } from "@composables/useRepositoryList";
import { useCommentDraft } from "@composables/useCommentDraft";
import { usePendingComments } from "@composables/usePendingComments";
import { useSelectedFile } from "@composables/useSelectedFile";
import { usePreferences } from "@composables/usePreferences";
import { useDiffLayout } from "@composables/useDiffLayout";
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
const {
  mode: authMode,
  osUsername: authOSUsername,
  currentUser,
} = storeToRefs(authStore);

const repoStore = useRepoStore();
const {
  state,
  activeRepoPath,
  loading,
  error,
} = storeToRefs(repoStore);

const reviewsStore = useReviewsStore();
const {
  items: reviews,
  active: activeReview,
  summaryDraft,
} = storeToRefs(reviewsStore);

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
const { collapsed, splitRatios, toggleCollapsed, setSplitRatio } = useDiffLayout();
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
    if (intent.kind === "pull-request" && intent.prNumber && state.value) {
      await openPullRequest(intent.prNumber);
    } else if (intent.sha && state.value) {
      await openCommit(intent.sha);
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

  if (intent?.kind === "pull-request" && intent.prNumber && state.value) {
    await openPullRequest(intent.prNumber);
  } else if (intent?.sha && state.value) {
    await openCommit(intent.sha);
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
  loading.value = true;
  error.value = "";
  activeRepoPath.value = path;
  try {
    state.value = await window.diffApp.repositoryState(path);
    selectedPath.value = state.value.files[0]?.path ?? "";
    resetSelectedFile();
    const reviewResponse = await window.diffApp.listReviews(25);
    reviews.value = reviewResponse.reviews.filter(
      (review) => review.repoRoot === state.value?.root,
    );
    activeReview.value = reviews.value[0]
      ? await window.diffApp.reviewDetail(reviews.value[0].id)
      : null;
    await loadPendingComments();
    await savePreferences({ [PREF_KEYS.lastRepoRoot]: state.value.root });
    await window.diffApp.upsertRepository({ path: state.value.root });
    await refreshRepositoryList();
    activeRepoPath.value = state.value.root;
  } catch (cause) {
    error.value = cause instanceof Error ? cause.message : String(cause);
    state.value = null;
  } finally {
    loading.value = false;
  }
}

async function refresh() {
  if (!state.value) {
    if (activeRepoPath.value) {
      await openRepo(activeRepoPath.value);
    }
    return;
  }
  loading.value = true;
  try {
    const current = state.value;
    if (current.mode === "commit" && current.commitSha) {
      state.value = await window.diffApp.readCommit(current.commitSha, current.root);
    } else {
      state.value = await window.diffApp.refreshRepository(current.root);
    }
    if (!state.value.files.some((file) => file.path === selectedPath.value)) {
      selectedPath.value = state.value.files[0]?.path ?? "";
    }
  } catch (cause) {
    error.value = cause instanceof Error ? cause.message : String(cause);
  } finally {
    loading.value = false;
  }
}

async function openCommit(sha: string) {
  if (!state.value) return;
  loading.value = true;
  error.value = "";
  try {
    state.value = await window.diffApp.readCommit(sha, state.value.root);
    selectedPath.value = state.value.files[0]?.path ?? "";
    activeReview.value = null;
    activePullRequest.value = null;
  } catch (cause) {
    error.value = cause instanceof Error ? cause.message : String(cause);
  } finally {
    loading.value = false;
  }
}

async function returnToWorkingTree() {
  if (!state.value) return;
  activePullRequest.value = null;
  await openRepo(state.value.root);
}

async function openPullRequest(number: number) {
  if (!state.value) return;
  loading.value = true;
  error.value = "";
  try {
    state.value = await window.diffApp.readPullRequest(number, state.value.root);
    selectedPath.value = state.value.files[0]?.path ?? "";
    activeReview.value = null;
    activePullRequest.value = pullRequests.value.find((pr) => pr.number === number) ?? {
      number,
      title: "",
      author: "",
      state: "open",
      baseRef: "",
      headRef: state.value.branch,
      url: "",
    };
  } catch (cause) {
    error.value = cause instanceof Error ? cause.message : String(cause);
  } finally {
    loading.value = false;
  }
}

const creatingBranch = ref(false);
const branchCreateError = ref("");

async function switchBranch(branch: string) {
  if (!state.value) return;
  loading.value = true;
  error.value = "";
  try {
    state.value = await window.diffApp.checkoutBranch(state.value.root, branch);
    if (!state.value.files.some((file) => file.path === selectedPath.value)) {
      selectedPath.value = state.value.files[0]?.path ?? "";
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
    error.value = structured?.message ?? (cause instanceof Error ? cause.message : String(cause));
  } finally {
    loading.value = false;
  }
}

async function createBranch(name: string) {
  if (!state.value) return;
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
  const chosen = await window.diffApp.chooseRepository(activeRepoPath.value || lastRepoRoot.value);
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
  if (!state.value) return;
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
  if (!state.value) return;
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
  if (!state.value) return false;
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

function openCommentForLine(file: ChangedFile, section: DiffSection, line: PatchLine) {
  const lineNumber = line.newLine ?? line.oldLine;
  if (!lineNumber || !activeReview.value) {
    reviewPanelOpen.value = true;
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
  if (!activeReview.value) return;
  await window.diffApp.deleteReviewComment({
    reviewId: activeReview.value.review.id,
    commentId: comment.id,
  });
  activeReview.value = await window.diffApp.reviewDetail(activeReview.value.review.id);
}

async function replyToComment(parent: ReviewComment, bodyHtml: string) {
  if (!activeReview.value) return;
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
  if (path) void navigator.clipboard.writeText(path);
}

const copyReviewState = ref<"idle" | "copied" | "error">("idle");

async function copyReviewAsMarkdown() {
  if (!activeReview.value) return;
  try {
    const markdown = formatReviewAsMarkdown(activeReview.value);
    await navigator.clipboard.writeText(markdown);
    copyReviewState.value = "copied";
    setTimeout(() => {
      if (copyReviewState.value === "copied") copyReviewState.value = "idle";
    }, 1500);
  } catch (cause) {
    copyReviewState.value = "error";
    error.value = cause instanceof Error ? cause.message : String(cause);
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

    <div
      v-if="state"
      class="flex flex-wrap items-center gap-2 border-b border-border bg-background/70 px-4 py-1.5 text-xs"
    >
      <CommitPicker
        :commits="commits"
        :loading="commitsLoading"
        :active-sha="state.commitSha"
        :mode="repoMode"
        @open="loadCommits()"
        @select="openCommit"
        @back="returnToWorkingTree"
      />
      <PullRequestPicker
        :pull-requests="pullRequests"
        :loading="pullRequestsLoading"
        :active-number="activePullRequest?.number"
        @open="loadPullRequests()"
        @select="openPullRequest"
      />
      <span
        v-if="activePullRequest"
        class="flex items-center gap-1 font-mono text-muted-foreground"
      >
        PR #{{ activePullRequest.number }}
        <span v-if="activePullRequest.title" class="font-sans"
          >· {{ activePullRequest.title }}</span
        >
      </span>
      <span
        v-else-if="repoMode === 'commit' && state.commitSha"
        class="font-mono text-muted-foreground"
      >
        commit {{ state.branch || state.commitSha.slice(0, 7) }}
      </span>
      <span class="flex-1" />
      <button
        type="button"
        class="inline-flex items-center gap-1.5 rounded border border-border bg-background px-2.5 py-1 text-xs font-medium hover:bg-muted disabled:opacity-50"
        :disabled="walkthroughLoading"
        @click="generateWalkthrough(walkthrough != null)"
      >
        {{
          walkthroughLoading
            ? "Generating…"
            : walkthrough
              ? "Refresh walkthrough"
              : "Generate walkthrough"
        }}
      </button>
      <button
        v-if="activeReview"
        type="button"
        class="inline-flex items-center gap-1.5 rounded border border-border bg-background px-2.5 py-1 text-xs font-medium hover:bg-muted disabled:opacity-50"
        :disabled="copyReviewState === 'copied'"
        @click="copyReviewAsMarkdown"
      >
        {{ copyReviewState === "copied" ? "Copied!" : "Copy review as Markdown" }}
      </button>
    </div>
    <WalkthroughPanel :record="walkthrough" :error="walkthroughError" />

    <main class="flex flex-1 min-h-0">
      <template v-if="!activeRepoPath || loading || error || !state">
        <div
          v-if="!activeRepoPath"
          class="flex flex-1 flex-col overflow-auto"
          :style="{ background: 'var(--gd-bg)' }"
        >
          <div class="mx-auto w-full max-w-5xl px-8 pt-16 pb-10">
            <div class="hero">
              <div
                class="flex h-10 w-10 items-center justify-center rounded-md border"
                :style="{ borderColor: 'var(--gd-border)', background: 'var(--gd-panel)' }"
              >
                <GitPullRequest class="h-5 w-5" :style="{ color: 'var(--gd-text-3)' }" />
              </div>
              <h1 class="hero-title">A local git diff viewer</h1>
              <p class="hero-subtitle">
                Inspect changes across your repositories with a fast file tree, side-by-side diffs,
                and lightweight local reviews.
              </p>
              <div class="hero-cta-row">
                <button class="toolbar-btn" type="button" @click="addRepository">
                  <Plus class="h-4 w-4" />Add repository
                </button>
                <button
                  v-if="lastRepoRoot"
                  class="toolbar-btn"
                  type="button"
                  @click="openRepo(lastRepoRoot)"
                >
                  <FolderOpen class="h-4 w-4" />Open last repo
                </button>
              </div>
              <div class="hero-version">No repository selected.</div>
            </div>
          </div>
        </div>
        <div
          v-else-if="loading"
          class="grid flex-1 place-items-center text-sm"
          :style="{ color: 'var(--gd-text-3)' }"
        >
          Loading repository…
        </div>
        <div v-else class="grid flex-1 place-items-center p-8">
          <div
            class="max-w-xl rounded-md border p-5"
            :style="{ borderColor: 'var(--gd-border)', background: 'var(--gd-panel)' }"
          >
            <div class="font-semibold">Unable to read repository</div>
            <p class="mt-2 text-sm" :style="{ color: 'var(--gd-text-3)' }">{{ error }}</p>
            <button class="mt-4 toolbar-btn" type="button" @click="addRepository">
              Choose another repository
            </button>
          </div>
        </div>
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
          @open-review-panel="reviewPanelOpen = true"
        />

        <section class="flex flex-col flex-1 min-w-0 relative">
          <div class="flex-1 overflow-auto min-h-0" style="padding: 16px 18px 24px">
            <FileContentViewer
              v-if="selectedPath && !selectedIsChanged"
              :file="selectedRepoFile"
              :path="selectedPath"
              :loading="selectedFileLoading"
              :error="selectedFileError"
            />
            <template v-else-if="files.length === 0">
              <div
                class="grid h-full place-items-center text-sm"
                :style="{ color: 'var(--gd-text-3)' }"
              >
                No changes detected. Edit some files and refresh.
              </div>
            </template>
            <template v-else>
              <article
                v-for="file in files"
                :id="fileElementID(file.path)"
                :key="file.path"
                :style="{
                  background: 'var(--gd-panel)',
                  borderRadius: '12px',
                  boxShadow: 'var(--gd-shadow-card)',
                  overflow: 'clip',
                  marginBottom: '18px',
                }"
              >
                <FileHeader
                  :file="file"
                  :collapsed="!!collapsed[file.path]"
                  :viewed="isViewed(file)"
                  @toggle-collapsed="toggleCollapsed(file.path)"
                  @toggle-viewed="toggleViewed(file)"
                  @copy="copyPath"
                />
                <DiffBody
                  v-if="!collapsed[file.path]"
                  :file="file"
                  :view-mode="diffViewMode"
                  :diff-style="tweaks.diffStyle"
                  :density="tweaks.density"
                  :word-highlight="tweaks.wordHighlight"
                  :hide-whitespace="hideWhitespace"
                  :comments="reviewComments"
                  :reply-features="commentFeatures"
                  :split-ratio="splitRatios[file.path] ?? 0.5"
                  @add-comment="(section, line) => openCommentForLine(file, section, line)"
                  @delete-comment="deleteComment"
                  @reply-comment="replyToComment"
                  @update:split-ratio="(value: number) => setSplitRatio(file.path, value)"
                />
              </article>
            </template>
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

    <ToastViewport :toasts="toasts" @dismiss="dismissToast" />
    </div>
  </AuthGate>
</template>
