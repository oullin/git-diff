<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from "vue";
import { GitPullRequest, FolderOpen, Plus } from "lucide-vue-next";
import AuthSetup from "@entry/components/AuthSetup.vue";
import AuthLogin from "@entry/components/AuthLogin.vue";
import FileContentViewer from "@entry/components/FileContentViewer.vue";
import TitleBar from "@entry/components/diff/TitleBar.vue";
import TopBar from "@entry/components/diff/TopBar.vue";
import Sidebar from "@entry/components/diff/Sidebar.vue";
import FileHeader from "@entry/components/diff/FileHeader.vue";
import DiffBody from "@entry/components/diff/DiffBody.vue";
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
  Repository,
  RepositoryFile,
  RepositoryState,
  ReviewComment,
  ReviewDetail,
  ReviewSession,
} from "@api";
import { PREF_KEYS, viewedPrefKey } from "@api";
import { ensureLanguage, languageFor } from "@lib/highlight";
import { ACCENTS, applyAccent, diffBgs, resolveAccent } from "@lib/accent";
import type { PatchLine } from "@lib/patch";
import { TWEAK_DEFAULTS, tweakPrefPatch, useTweaks, type Tweaks } from "@composables/useTweaks";
import { setTheme } from "@composables/useTheme";

type AuthMode = "loading" | "setup" | "login" | "ready";

const commentFeatures: RichTextFeatures = {
  checklist: false,
  images: false,
  markdownShortcuts: true,
  mentions: false,
  slashMenu: false,
  tables: false,
};

const state = ref<RepositoryState | null>(null);
const prefValues = ref<Record<string, string>>({});
const authMode = ref<AuthMode>("loading");
const authOSUsername = ref("");
const currentUser = ref<AuthUser | null>(null);

const prefs = computed(() => prefValues.value);
const tweaks = useTweaks(prefs);

const accent = computed(() => resolveAccent(prefValues.value[PREF_KEYS.uiAccent]));
const diffViewMode = computed<DiffViewMode>(() => tweaks.value.viewMode);
const hideWhitespace = computed(() => prefValues.value[PREF_KEYS.diffHideWhitespace] === "1");
const lastRepoRoot = computed(() => prefValues.value[PREF_KEYS.lastRepoRoot] ?? "");

const repositories = ref<Repository[]>([]);
const repositoriesLoading = ref(false);
const activeRepoPath = ref<string>("");
const reviews = ref<ReviewSession[]>([]);
const activeReview = ref<ReviewDetail | null>(null);
const selectedPath = ref("");
const searchQuery = ref("");
const loading = ref(false);
const error = ref("");
const collapsed = ref<Record<string, boolean>>({});
const splitRatios = ref<Record<string, number>>({});
const reviewPanelOpen = ref(false);
const commentDialogOpen = ref(false);
const commentTarget = ref<{
  file: ChangedFile;
  section: DiffSection;
  line: number;
  side: string;
} | null>(null);
const commentDraft = ref("");
const summaryDraft = ref("");

const selectedRepoFile = ref<RepositoryFile | null>(null);
const selectedFileLoading = ref(false);
const selectedFileError = ref("");
const repoScope = ref<"changed" | "all">("changed");

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

watch(
  accent,
  (current) => {
    applyAccent(current);
  },
  { immediate: true },
);

watch(
  () => tweaks.value.theme,
  (choice) => {
    setTheme(choice);
  },
  { immediate: true },
);

watch(
  () => tweaks.value.diffStyle,
  (style) => {
    const colors = diffBgs(style);
    document.documentElement.style.setProperty("--gd-word-add-bg", colors.addStrong);
    document.documentElement.style.setProperty("--gd-word-rem-bg", colors.remStrong);
  },
  { immediate: true },
);

onMounted(async () => {
  await bootstrapAuth();
  document.addEventListener("keydown", onKeydown);
});

onUnmounted(() => {
  document.removeEventListener("keydown", onKeydown);
});

function onKeydown(event: KeyboardEvent) {
  if (authMode.value !== "ready") return;
  const target = event.target as HTMLElement | null;
  if (
    target &&
    (target.tagName === "INPUT" || target.tagName === "TEXTAREA" || target.isContentEditable)
  ) {
    return;
  }
  if (event.key === "j" || event.key === "ArrowDown") {
    if (event.metaKey || event.ctrlKey || event.altKey) return;
    event.preventDefault();
    selectAdjacent(1);
  } else if (event.key === "k" || event.key === "ArrowUp") {
    if (event.metaKey || event.ctrlKey || event.altKey) return;
    event.preventDefault();
    selectAdjacent(-1);
  } else if (event.key === "v") {
    const file = selectedFile.value;
    if (file && changedByPath.value.has(file.path)) {
      event.preventDefault();
      void toggleViewed(file);
    }
  } else if (event.key === "Enter" && (event.metaKey || event.ctrlKey)) {
    event.preventDefault();
    void startReview();
  }
}

function selectAdjacent(delta: number) {
  if (files.value.length === 0) return;
  const next = Math.min(Math.max(changedIndex.value + delta, 0), files.value.length - 1);
  const file = files.value[next];
  if (file) selectFile(file.path);
}

async function bootstrapAuth() {
  authMode.value = "loading";
  try {
    const result = await window.diffApp.authBootstrap();
    authOSUsername.value = result.state.osUsername;
    if (result.user) {
      currentUser.value = result.user;
      await enterApp();
      return;
    }
    authMode.value = result.state.needsSetup ? "setup" : "login";
  } catch (cause) {
    error.value = cause instanceof Error ? cause.message : String(cause);
    authMode.value = "login";
  }
}

async function enterApp() {
  authMode.value = "ready";
  prefValues.value = (await window.diffApp.getUIPreferences()).values;
  await refreshRepositoryList();
  const lastRoot = lastRepoRoot.value;
  const initial =
    repositories.value.find((repo) => repo.path === lastRoot)?.path ??
    repositories.value[0]?.path ??
    "";
  if (initial) {
    await openRepo(initial);
  }
}

async function handleAuthCompleted(response: AuthLoginResponse) {
  currentUser.value = response.user;
  await enterApp();
}

async function handleAuthWiped() {
  currentUser.value = null;
  authMode.value = "setup";
}

async function logOut() {
  try {
    await window.diffApp.authLogout();
  } catch {
    // ignore — we'll still reset locally
  }
  currentUser.value = null;
  prefValues.value = {};
  state.value = null;
  repositories.value = [];
  activeRepoPath.value = "";
  reviews.value = [];
  activeReview.value = null;
  selectedPath.value = "";
  selectedRepoFile.value = null;
  await bootstrapAuth();
}

async function openRepo(path: string) {
  loading.value = true;
  error.value = "";
  activeRepoPath.value = path;
  try {
    state.value = await window.diffApp.repositoryState(path);
    selectedPath.value = state.value.files[0]?.path ?? "";
    selectedRepoFile.value = null;
    selectedFileError.value = "";
    const reviewResponse = await window.diffApp.listReviews(25);
    reviews.value = reviewResponse.reviews.filter(
      (review) => review.repoRoot === state.value?.root,
    );
    activeReview.value = reviews.value[0]
      ? await window.diffApp.reviewDetail(reviews.value[0].id)
      : null;
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
    state.value = await window.diffApp.refreshRepository(state.value.root);
    if (!state.value.files.some((file) => file.path === selectedPath.value)) {
      selectedPath.value = state.value.files[0]?.path ?? "";
    }
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
    error.value = cause instanceof Error ? cause.message : String(cause);
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

async function refreshRepositoryList() {
  repositoriesLoading.value = true;
  try {
    repositories.value = await window.diffApp.listRepositories();
  } finally {
    repositoriesLoading.value = false;
  }
}

async function removeRepository(path: string) {
  await window.diffApp.removeRepository(path);
  await refreshRepositoryList();
  if (activeRepoPath.value === path) {
    activeRepoPath.value = "";
    state.value = null;
    activeReview.value = null;
    reviews.value = [];
    selectedPath.value = "";
    selectedRepoFile.value = null;
    selectedFileError.value = "";
    reviewPanelOpen.value = false;
  }
}

async function startReview() {
  if (!state.value) return;
  const review = await window.diffApp.createReview({
    repoRoot: state.value.root,
    branch: state.value.branch,
    headSha: state.value.headSha,
    title: `Review ${state.value.branch || state.value.headSha || "local changes"}`,
    summary: sanitizeHtml(summaryDraft.value),
    filesChanged: state.value.files.length,
    additions: state.value.additions,
    deletions: state.value.deletions,
  });
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

function toggleCollapsed(path: string) {
  collapsed.value[path] = !collapsed.value[path];
}

function selectFile(path: string) {
  selectedPath.value = path;
  if (changedByPath.value.has(path)) {
    selectedRepoFile.value = null;
    selectedFileError.value = "";
    nextTick(() =>
      document.getElementById(fileElementID(path))?.scrollIntoView({ block: "start" }),
    );
    return;
  }
  void loadRepoFile(path);
}

async function loadRepoFile(path: string) {
  if (!state.value || !path) {
    selectedRepoFile.value = null;
    return;
  }
  const root = state.value.root;
  selectedFileLoading.value = true;
  selectedFileError.value = "";
  selectedRepoFile.value = null;
  try {
    const file = await window.diffApp.readRepositoryFile(root, path);
    if (selectedPath.value === path) {
      selectedRepoFile.value = file;
    }
  } catch (cause) {
    if (selectedPath.value === path) {
      selectedFileError.value = cause instanceof Error ? cause.message : String(cause);
    }
  } finally {
    if (selectedPath.value === path) {
      selectedFileLoading.value = false;
    }
  }
}

function openCommentForLine(file: ChangedFile, section: DiffSection, line: PatchLine) {
  const lineNumber = line.newLine ?? line.oldLine;
  if (!lineNumber || !activeReview.value) {
    reviewPanelOpen.value = true;
    return;
  }
  commentTarget.value = {
    file,
    section,
    line: lineNumber,
    side: line.newLine ? "right" : "left",
  };
  commentDraft.value = "";
  commentDialogOpen.value = true;
}

async function saveComment() {
  if (!commentTarget.value || !activeReview.value || !commentDraft.value.trim()) {
    return;
  }
  const target = commentTarget.value;
  await window.diffApp.createReviewComment({
    reviewId: activeReview.value.review.id,
    filePath: target.file.path,
    diffSection: target.section.kind,
    side: target.side,
    lineNumber: target.line,
    authorLabel: currentUser.value?.displayName || currentUser.value?.osUsername || "You",
    bodyHtml: sanitizeHtml(commentDraft.value),
  });
  activeReview.value = await window.diffApp.reviewDetail(activeReview.value.review.id);
  commentTarget.value = null;
  commentDraft.value = "";
  commentDialogOpen.value = false;
}

function cancelComment() {
  commentTarget.value = null;
  commentDraft.value = "";
  commentDialogOpen.value = false;
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

async function savePreferences(patch: Record<string, string>) {
  const next = { ...prefValues.value, ...patch };
  for (const [key, value] of Object.entries(patch)) {
    if (value === "") {
      delete next[key];
    } else {
      next[key] = value;
    }
  }
  prefValues.value = next;
  try {
    const response = await window.diffApp.saveUIPreferences(patch);
    prefValues.value = response.values;
  } catch (cause) {
    error.value = cause instanceof Error ? cause.message : String(cause);
  }
}

function fileElementID(path: string): string {
  return `file-${path.replace(/[^a-z0-9_-]/gi, "-")}`;
}

function copyPath(path: string) {
  if (path) void navigator.clipboard.writeText(path);
}

void TWEAK_DEFAULTS;
void ACCENTS;
</script>

<template>
  <div
    v-if="authMode === 'loading'"
    class="grid h-screen place-items-center bg-background text-sm text-muted-foreground"
  >
    Loading…
  </div>
  <AuthSetup
    v-else-if="authMode === 'setup'"
    :os-username="authOSUsername"
    @completed="handleAuthCompleted"
  />
  <AuthLogin
    v-else-if="authMode === 'login'"
    :os-username="authOSUsername"
    @logged-in="handleAuthCompleted"
    @wiped="handleAuthWiped"
  />
  <div
    v-else
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
                  @update:split-ratio="(value: number) => (splitRatios[file.path] = value)"
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
  </div>
</template>
