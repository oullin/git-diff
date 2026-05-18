<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from "vue";
import { parsePatchFiles } from "@pierre/diffs";
import {
  Check,
  ChevronDown,
  ChevronRight,
  FolderOpen,
  GitPullRequest,
  LogOut,
  MessageSquare,
  Plus,
  RefreshCw,
  Search,
  Trash2,
  X,
} from "lucide-vue-next";
import { LazyRichTextEditor, type RichTextFeatures } from "@ui/rich-text-editor";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@ui/dropdown-menu";
import { ResizableHandle, ResizablePanel, ResizablePanelGroup } from "@ui/resizable";
import RepoFileTree from "@entry/components/RepoFileTree.vue";
import FileContentViewer from "@entry/components/FileContentViewer.vue";
import AuthSetup from "@entry/components/AuthSetup.vue";
import AuthLogin from "@entry/components/AuthLogin.vue";
import { SafeHtml, sanitizeHtml } from "@ui/safe-html";
import type {
  AuthLoginResponse,
  AuthUser,
  ChangedFile,
  DiffSection,
  DiffViewMode,
  Repository,
  RepositoryFile,
  RepositoryState,
  ReviewComment,
  ReviewDetail,
  ReviewSession,
} from "@api";
import { PREF_KEYS } from "@api";
import { cn } from "@lib/utils";
import { ensureLanguage, highlightLine, highlighterRev, languageFor } from "@lib/highlight";

type PatchLine = {
  id: string;
  type: "context" | "add" | "del" | "meta";
  text: string;
  oldLine?: number;
  newLine?: number;
};

type SplitRow =
  | { id: string; kind: "meta"; line: PatchLine }
  | { id: string; kind: "context"; line: PatchLine }
  | { id: string; kind: "pair"; left?: PatchLine; right?: PatchLine };

const commentFeatures: RichTextFeatures = {
  checklist: false,
  images: false,
  markdownShortcuts: true,
  mentions: false,
  slashMenu: false,
  tables: false,
};

const DEFAULT_PANEL_RIGHT_WIDTH = 25;
const DEFAULT_PANEL_FILE_TREE_WIDTH = 28;

type AuthMode = "loading" | "setup" | "login" | "ready";

const state = ref<RepositoryState | null>(null);
const prefValues = ref<Record<string, string>>({});
const authMode = ref<AuthMode>("loading");
const authOSUsername = ref("");
const currentUser = ref<AuthUser | null>(null);

const theme = computed(() => prefValues.value[PREF_KEYS.theme] ?? "dark");
const diffViewMode = computed<DiffViewMode>(
  () => (prefValues.value[PREF_KEYS.diffViewMode] as DiffViewMode) || "split",
);
const hideWhitespace = computed(() => prefValues.value[PREF_KEYS.diffHideWhitespace] === "1");
const lastRepoRoot = computed(() => prefValues.value[PREF_KEYS.lastRepoRoot] ?? "");
const panelFileTreeWidth = computed(() =>
  parsePercent(prefValues.value[PREF_KEYS.panelFileTreeWidth]),
);
const panelRightWidth = computed(() => parsePercent(prefValues.value[PREF_KEYS.panelRightWidth]));

function parsePercent(raw: string | undefined): number | null {
  if (!raw) {
    return null;
  }

  const value = Number(raw);

  if (!Number.isFinite(value) || value <= 0 || value >= 100) {
    return null;
  }

  return value;
}

const rightPanelDefault = computed(() => panelRightWidth.value ?? DEFAULT_PANEL_RIGHT_WIDTH);
const fileTreeDefault = computed(() => panelFileTreeWidth.value ?? DEFAULT_PANEL_FILE_TREE_WIDTH);
const diffPanelDefault = computed(() => {
  const remaining =
    100 - fileTreeDefault.value - (reviewPanelOpen.value ? rightPanelDefault.value : 0);

  return Math.max(25, remaining);
});
const repositories = ref<Repository[]>([]);
const activeRepoPath = ref<string>("");
const activeRepo = computed(
  () => repositories.value.find((repo) => repo.path === activeRepoPath.value) ?? null,
);
const reviews = ref<ReviewSession[]>([]);
const activeReview = ref<ReviewDetail | null>(null);
const selectedPath = ref("");
const searchQuery = ref("");
const loading = ref(false);
const error = ref("");
const viewed = ref<Record<string, string>>({});
const collapsed = ref<Record<string, boolean>>({});
const reviewPanelOpen = ref(false);
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
const repoViewMode = ref<"changed" | "all">("changed");

const files = computed(() => state.value?.files ?? []);
const changedByPath = computed(() => {
  const map = new Map<string, ChangedFile>();
  for (const file of files.value) {
    map.set(file.path, file);
  }
  return map;
});
const changedPathsSet = computed(() => new Set(changedByPath.value.keys()));
const filteredFiles = computed(() => {
  const query = searchQuery.value.trim().toLowerCase();
  if (!query) {
    return files.value;
  }

  return files.value.filter((file) => file.path.toLowerCase().includes(query));
});
const trackedFiles = computed(() => state.value?.trackedFiles ?? []);
const repoPaths = computed(() => {
  const set = new Set<string>(trackedFiles.value);
  for (const file of files.value) {
    set.add(file.path);
  }
  return Array.from(set).sort();
});
const filteredRepoPaths = computed(() => {
  const query = searchQuery.value.trim().toLowerCase();
  if (!query) {
    return repoPaths.value;
  }
  return repoPaths.value.filter((path) => path.toLowerCase().includes(query));
});
const changedPathsList = computed(() => filteredFiles.value.map((file) => file.path));
const visibleTreePaths = computed(() =>
  repoViewMode.value === "all" ? filteredRepoPaths.value : changedPathsList.value,
);
const selectedIsChanged = computed(
  () => !!selectedPath.value && changedByPath.value.has(selectedPath.value),
);
const reviewComments = computed(() => activeReview.value?.comments ?? []);

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

function highlightedLine(text: string, file: ChangedFile): string {
  // Read the rev so Vue re-runs this when a language finishes loading.
  void highlighterRev.value;
  return highlightLine(text, languageFor(file.path));
}

onMounted(async () => {
  await bootstrapAuth();
});

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
  repositories.value = await window.diffApp.listRepositories();
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
    viewed.value = readViewed(state.value.root);
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

async function addRepository() {
  const chosen = await window.diffApp.chooseRepository(activeRepoPath.value || lastRepoRoot.value);
  if (chosen) {
    await openRepo(chosen);
  }
}

async function refreshRepositoryList() {
  repositories.value = await window.diffApp.listRepositories();
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

function openReviewPanel() {
  reviewPanelOpen.value = true;
}

function closeReviewPanel() {
  reviewPanelOpen.value = false;
  commentTarget.value = null;
}

async function startReview() {
  if (!state.value) {
    return;
  }
  reviewPanelOpen.value = true;
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
}

async function selectReview(review: ReviewSession) {
  activeReview.value = await window.diffApp.reviewDetail(review.id);
}

async function toggleViewed(file: ChangedFile) {
  if (!state.value) {
    return;
  }
  const next = { ...viewed.value };
  if (isViewed(file)) {
    delete next[file.path];
  } else {
    next[file.path] = file.fingerprint;
  }
  viewed.value = next;
  localStorage.setItem(viewedKey(state.value.root), JSON.stringify(next));

  if (activeReview.value) {
    await window.diffApp.addReviewEvent({
      reviewId: activeReview.value.review.id,
      type: isViewed(file) ? "file_viewed" : "file_unviewed",
      filePath: file.path,
      message: isViewed(file) ? "Marked viewed" : "Marked unviewed",
    });
    activeReview.value = await window.diffApp.reviewDetail(activeReview.value.review.id);
  }
}

function isViewed(file: ChangedFile) {
  return viewed.value[file.path] === file.fingerprint;
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

function openComment(file: ChangedFile, section: DiffSection, line: PatchLine) {
  const lineNumber = line.newLine ?? line.oldLine;
  if (!lineNumber || !activeReview.value) {
    return;
  }
  commentTarget.value = {
    file,
    section,
    line: lineNumber,
    side: line.newLine ? "right" : "left",
  };
  commentDraft.value = "";
  reviewPanelOpen.value = true;
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
    authorLabel: "You",
    bodyHtml: sanitizeHtml(commentDraft.value),
  });
  activeReview.value = await window.diffApp.reviewDetail(activeReview.value.review.id);
  commentTarget.value = null;
  commentDraft.value = "";
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

async function setViewMode(mode: DiffViewMode) {
  await savePreferences({ [PREF_KEYS.diffViewMode]: mode });
}

async function toggleWhitespace() {
  await savePreferences({ [PREF_KEYS.diffHideWhitespace]: hideWhitespace.value ? "" : "1" });
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

const panelSaveTimers = new Map<string, ReturnType<typeof setTimeout>>();

function schedulePanelSave(key: string, value: number) {
  const existing = panelSaveTimers.get(key);

  if (existing) {
    clearTimeout(existing);
  }

  const timer = setTimeout(() => {
    panelSaveTimers.delete(key);
    void savePreferences({ [key]: value.toFixed(2) });
  }, 300);

  panelSaveTimers.set(key, timer);
}

function handleOuterLayout(sizes: number[]) {
  if (sizes.length === 0) {
    return;
  }

  if (activeRepoPath.value && !loading.value && !error.value) {
    schedulePanelSave(PREF_KEYS.panelFileTreeWidth, sizes[0]!);
  }

  if (sizes.length >= 2 && reviewPanelOpen.value) {
    schedulePanelSave(PREF_KEYS.panelRightWidth, sizes[sizes.length - 1]!);
  }
}

function parsePatch(section: DiffSection): PatchLine[] {
  if (section.binary) {
    return [{ id: `${section.id}:binary`, type: "meta", text: "Binary file changed" }];
  }
  try {
    parsePatchFiles(section.patch);
  } catch {
    // Rendering below is deliberately tolerant of partial git output.
  }

  const lines: PatchLine[] = [];
  let oldLine = 0;
  let newLine = 0;
  for (const [index, raw] of section.patch.split("\n").entries()) {
    const hunk = /^@@ -(\d+)(?:,\d+)? \+(\d+)(?:,\d+)? @@/.exec(raw);
    if (hunk) {
      oldLine = Number(hunk[1]);
      newLine = Number(hunk[2]);
      lines.push({ id: `${section.id}:${index}`, type: "meta", text: raw });
      continue;
    }
    if (
      raw.startsWith("diff --git") ||
      raw.startsWith("index ") ||
      raw.startsWith("---") ||
      raw.startsWith("+++")
    ) {
      lines.push({ id: `${section.id}:${index}`, type: "meta", text: raw });
      continue;
    }
    if (raw.startsWith("+")) {
      lines.push({ id: `${section.id}:${index}`, type: "add", text: raw.slice(1), newLine });
      newLine++;
      continue;
    }
    if (raw.startsWith("-")) {
      lines.push({ id: `${section.id}:${index}`, type: "del", text: raw.slice(1), oldLine });
      oldLine++;
      continue;
    }
    lines.push({
      id: `${section.id}:${index}`,
      type: "context",
      text: raw.startsWith(" ") ? raw.slice(1) : raw,
      oldLine,
      newLine,
    });
    oldLine++;
    newLine++;
  }

  return hideWhitespace.value
    ? lines.filter((line) => line.type === "meta" || line.text.trim() !== "")
    : lines;
}

function splitPatch(section: DiffSection): SplitRow[] {
  const lines = parsePatch(section);
  const rows: SplitRow[] = [];
  let i = 0;

  while (i < lines.length) {
    const line = lines[i]!;

    if (line.type === "meta") {
      rows.push({ id: `${line.id}:split-meta`, kind: "meta", line });
      i++;
      continue;
    }

    if (line.type === "context") {
      rows.push({ id: `${line.id}:split-ctx`, kind: "context", line });
      i++;
      continue;
    }

    const dels: PatchLine[] = [];
    while (i < lines.length && lines[i]!.type === "del") {
      dels.push(lines[i]!);
      i++;
    }

    const adds: PatchLine[] = [];
    while (i < lines.length && lines[i]!.type === "add") {
      adds.push(lines[i]!);
      i++;
    }

    const pairCount = Math.max(dels.length, adds.length);
    for (let j = 0; j < pairCount; j++) {
      const left = dels[j];
      const right = adds[j];
      const id = `${(left ?? right)!.id}:split-pair:${j}`;
      rows.push({ id, kind: "pair", left, right });
    }
  }

  return rows;
}

function splitRowCommentLines(row: SplitRow): PatchLine[] {
  if (row.kind === "pair") {
    return [row.left, row.right].filter((line): line is PatchLine => Boolean(line));
  }
  if (row.kind === "context") {
    return [row.line];
  }
  return [];
}

function commentsFor(file: ChangedFile, section: DiffSection, line?: PatchLine) {
  if (!line) {
    return reviewComments.value.filter(
      (comment) => comment.filePath === file.path && comment.diffSection === section.kind,
    );
  }
  const lineNumber = line.newLine ?? line.oldLine;
  return reviewComments.value.filter(
    (comment) =>
      comment.filePath === file.path &&
      comment.diffSection === section.kind &&
      comment.lineNumber === lineNumber,
  );
}

function viewedKey(root: string) {
  return `git-diff:viewed:${root}`;
}

function readViewed(root: string) {
  try {
    return JSON.parse(localStorage.getItem(viewedKey(root)) ?? "{}") as Record<string, string>;
  } catch {
    return {};
  }
}

function fileElementID(path: string) {
  return `file-${path.replace(/[^a-z0-9_-]/gi, "-")}`;
}

function statusLabel(file: ChangedFile) {
  return file.status[0]?.toUpperCase() ?? "M";
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat(undefined, { dateStyle: "medium", timeStyle: "short" }).format(
    new Date(value),
  );
}
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
    class="flex h-screen flex-col overflow-hidden bg-background pt-[var(--app-header-height)] text-foreground"
  >
    <header
      class="fixed inset-x-0 top-0 z-50 flex h-[var(--app-header-height)] items-center gap-3 border-b border-border bg-section-muted px-4"
    >
      <div class="relative w-72 max-w-[40%]">
        <Search
          class="pointer-events-none absolute left-2 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground"
        />
        <input
          v-model="searchQuery"
          class="h-8 w-full rounded-md border border-input bg-background pl-8 pr-3 text-sm outline-none focus:ring-2 focus:ring-ring"
          placeholder="Filter files"
        />
      </div>
      <div class="flex-1" />
      <div v-if="state" class="truncate text-xs text-muted-foreground" :title="state.root">
        {{ state.branch || "detached" }} · {{ state.headSha || "no HEAD" }} ·
        <span class="text-success">+{{ state.additions }}</span>
        <span class="text-destructive">-{{ state.deletions }}</span>
      </div>
      <DropdownMenu>
        <DropdownMenuTrigger as-child>
          <button
            class="toolbar-btn min-w-0 max-w-[260px]"
            type="button"
            :title="activeRepo?.path ?? ''"
          >
            <FolderOpen class="h-4 w-4 shrink-0 text-muted-foreground" />
            <span class="min-w-0 truncate text-sm font-semibold">
              {{ activeRepo?.name ?? state?.root ?? "Select repository" }}
            </span>
            <ChevronDown class="h-3.5 w-3.5 shrink-0 text-muted-foreground" />
          </button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end" class="w-[320px]">
          <DropdownMenuLabel>Repositories</DropdownMenuLabel>
          <DropdownMenuSeparator />
          <div v-if="repositories.length === 0" class="px-2 py-3 text-xs text-muted-foreground">
            No repositories yet.
          </div>
          <DropdownMenuItem
            v-for="repo in repositories"
            :key="repo.path"
            class="group flex items-center gap-2"
            @select="openRepo(repo.path)"
          >
            <Check
              :class="
                cn(
                  'h-3.5 w-3.5 shrink-0',
                  activeRepoPath === repo.path ? 'opacity-100' : 'opacity-0',
                )
              "
            />
            <span class="min-w-0 flex-1 truncate" :title="repo.path">{{ repo.name }}</span>
            <button
              class="icon-btn opacity-0 group-hover:opacity-100"
              type="button"
              title="Remove from list"
              @click.stop="removeRepository(repo.path)"
            >
              <Trash2 class="h-3.5 w-3.5" />
            </button>
          </DropdownMenuItem>
          <DropdownMenuSeparator />
          <DropdownMenuItem @select="addRepository">
            <Plus class="h-4 w-4" />Add repository…
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
      <button class="toolbar-btn" type="button" @click="refresh" :disabled="!state">
        <RefreshCw class="h-4 w-4" />Refresh
      </button>
      <button class="toolbar-btn" type="button" :disabled="!state" @click="openReviewPanel">
        <MessageSquare class="h-4 w-4" />Reviews
      </button>
      <div class="flex rounded-md border border-border p-0.5">
        <button
          :class="cn('seg-btn', diffViewMode === 'split' && 'seg-active')"
          type="button"
          @click="setViewMode('split')"
        >
          Split
        </button>
        <button
          :class="cn('seg-btn', diffViewMode === 'unified' && 'seg-active')"
          type="button"
          @click="setViewMode('unified')"
        >
          Unified
        </button>
      </div>
      <button class="toolbar-btn" type="button" @click="toggleWhitespace">Whitespace</button>
      <button
        v-if="currentUser"
        class="toolbar-btn"
        type="button"
        :title="`Log out ${currentUser.osUsername}`"
        @click="logOut"
      >
        <LogOut class="h-4 w-4" />Log out
      </button>
    </header>

    <ResizablePanelGroup
      direction="horizontal"
      class="min-h-0 flex-1 overflow-hidden"
      @layout="handleOuterLayout"
    >
      <template v-if="!activeRepoPath || loading || error">
        <ResizablePanel
          :default-size="reviewPanelOpen ? 100 - rightPanelDefault : 100"
          :min-size="25"
          class="flex min-h-0 min-w-0 flex-col overflow-hidden bg-panel"
        >
          <div v-if="!activeRepoPath" class="flex flex-1 flex-col overflow-auto bg-background">
            <div class="mx-auto w-full max-w-5xl px-8 pt-16 pb-10">
              <div class="hero">
                <div
                  class="flex h-10 w-10 items-center justify-center rounded-md border border-border bg-section"
                >
                  <GitPullRequest class="h-5 w-5 text-muted-foreground" />
                </div>
                <h1 class="hero-title">A local git diff viewer</h1>
                <p class="hero-subtitle">
                  Inspect changes across your repositories with a fast file tree, side-by-side
                  diffs, and lightweight local reviews. Made for the terminal-native developer.
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

            <div class="mx-auto w-full max-w-5xl px-8 pb-16">
              <div class="window-frame">
                <div class="window-titlebar">
                  <span class="window-dots"><span /><span /><span /></span>
                  <span class="window-title">git-diff</span>
                </div>
                <div class="grid grid-cols-[220px_1fr] min-h-[280px]">
                  <div class="border-r border-border bg-panel p-3">
                    <div class="text-xs text-muted-foreground font-mono">
                      Add a repository to see your file tree here.
                    </div>
                  </div>
                  <div class="p-6 font-mono text-xs text-muted-foreground">
                    <div>// Your changed files appear here.</div>
                    <div class="mt-1">// Click "Add repository" to begin.</div>
                  </div>
                </div>
              </div>
            </div>
          </div>
          <div
            v-else-if="loading"
            class="grid flex-1 place-items-center text-sm text-muted-foreground"
          >
            Loading repository…
          </div>
          <div v-else class="grid flex-1 place-items-center p-8">
            <div class="max-w-xl rounded-md border border-border bg-section p-5">
              <div class="font-semibold">Unable to read repository</div>
              <p class="mt-2 text-sm text-muted-foreground">{{ error }}</p>
              <button class="mt-4 toolbar-btn" type="button" @click="addRepository">
                Choose another repository
              </button>
            </div>
          </div>
        </ResizablePanel>
      </template>
      <template v-else>
        <ResizablePanel
          :default-size="fileTreeDefault"
          :min-size="15"
          :max-size="60"
          class="flex min-h-0 flex-col overflow-hidden border-r border-border bg-sidebar"
          style="min-width: 240px"
        >
          <aside class="flex min-h-0 flex-1 flex-col overflow-hidden">
            <div class="border-b border-border p-3">
              <div class="flex items-center justify-between gap-3">
                <span class="text-xs text-muted-foreground">
                  {{ visibleTreePaths.length }}
                  {{ repoViewMode === "all" ? "files" : "changed" }}
                </span>
                <div
                  class="inline-flex items-center rounded-md border border-input bg-background p-0.5 text-xs"
                  role="tablist"
                  aria-label="File view mode"
                >
                  <button
                    :class="
                      cn(
                        'rounded px-2 py-1 transition-colors',
                        repoViewMode === 'changed'
                          ? 'bg-muted text-foreground'
                          : 'text-muted-foreground hover:text-foreground',
                      )
                    "
                    type="button"
                    role="tab"
                    :aria-selected="repoViewMode === 'changed'"
                    @click="repoViewMode = 'changed'"
                  >
                    Changed
                  </button>
                  <button
                    :class="
                      cn(
                        'rounded px-2 py-1 transition-colors',
                        repoViewMode === 'all'
                          ? 'bg-muted text-foreground'
                          : 'text-muted-foreground hover:text-foreground',
                      )
                    "
                    type="button"
                    role="tab"
                    :aria-selected="repoViewMode === 'all'"
                    @click="repoViewMode = 'all'"
                  >
                    All
                  </button>
                </div>
              </div>
            </div>
            <div
              v-if="visibleTreePaths.length === 0"
              class="px-3 py-2 text-xs text-muted-foreground"
            >
              {{ repoViewMode === "all" ? "No files match the filter." : "No changed files." }}
            </div>
            <RepoFileTree
              v-else
              :key="repoViewMode"
              :paths="visibleTreePaths"
              :selected-path="selectedPath"
              :changed-paths="changedPathsSet"
              :initial-expansion="repoViewMode === 'all' ? 'closed' : 'open'"
              class="min-h-0 flex-1 overflow-auto p-2"
              @select="selectFile"
            />
          </aside>
        </ResizablePanel>

        <ResizableHandle with-handle />

        <ResizablePanel
          :default-size="diffPanelDefault"
          :min-size="30"
          class="flex min-h-0 min-w-0 flex-col overflow-hidden bg-panel"
          style="min-width: 320px"
        >
          <section
            :class="
              cn(
                'min-h-0 flex-1 bg-panel',
                selectedPath && !selectedIsChanged
                  ? 'flex flex-col overflow-hidden'
                  : 'overflow-auto p-4',
              )
            "
          >
            <FileContentViewer
              v-if="selectedPath && !selectedIsChanged"
              :file="selectedRepoFile"
              :path="selectedPath"
              :loading="selectedFileLoading"
              :error="selectedFileError"
            />
            <div
              v-else-if="files.length === 0"
              class="grid h-full place-items-center text-sm text-muted-foreground"
            >
              Select a file from the tree to preview its contents.
            </div>
            <template v-else>
              <article
                v-for="file in files"
                :id="fileElementID(file.path)"
                :key="file.path"
                class="mb-4 overflow-hidden rounded-md border border-border bg-background"
              >
                <header
                  class="flex min-h-12 items-center gap-3 border-b border-border bg-section-muted px-3"
                >
                  <button
                    class="icon-btn"
                    type="button"
                    @click="collapsed[file.path] = !collapsed[file.path]"
                  >
                    <ChevronRight v-if="collapsed[file.path]" class="h-4 w-4" />
                    <ChevronDown v-else class="h-4 w-4" />
                  </button>
                  <span :class="cn('status-dot', `status-${file.status}`)">{{
                    statusLabel(file)
                  }}</span>
                  <div class="min-w-0 flex-1">
                    <div class="truncate text-sm font-semibold">{{ file.path }}</div>
                    <div v-if="file.oldPath" class="truncate text-xs text-muted-foreground">
                      {{ file.oldPath }}
                    </div>
                  </div>
                  <div class="text-xs">
                    <span class="text-success">+{{ file.additions }}</span>
                    <span class="ml-2 text-destructive">-{{ file.deletions }}</span>
                  </div>
                  <button class="toolbar-btn" type="button" @click="toggleViewed(file)">
                    <Check class="h-4 w-4" />{{ isViewed(file) ? "Viewed" : "Mark viewed" }}
                  </button>
                </header>

                <div v-if="!collapsed[file.path]">
                  <section
                    v-for="section in file.sections"
                    :key="section.id"
                    class="border-b border-border last:border-b-0"
                  >
                    <div
                      class="border-b border-border bg-muted px-3 py-2 text-xs font-medium uppercase text-muted-foreground"
                    >
                      {{ section.kind }}
                    </div>
                    <div class="diff-table" :data-view-mode="diffViewMode">
                      <template v-if="diffViewMode === 'unified'">
                        <template v-for="line in parsePatch(section)" :key="line.id">
                          <button
                            :class="cn('diff-line', `diff-${line.type}`)"
                            type="button"
                            @click="openComment(file, section, line)"
                          >
                            <span class="line-no">{{ line.oldLine ?? "" }}</span>
                            <span class="line-no">{{ line.newLine ?? "" }}</span>
                            <code
                              v-if="line.type !== 'meta'"
                              v-html="highlightedLine(line.text || ' ', file)"
                            />
                            <code v-else>{{ line.text || " " }}</code>
                            <MessageSquare
                              v-if="line.type !== 'meta'"
                              class="comment-icon h-3.5 w-3.5"
                            />
                          </button>
                          <div
                            v-for="comment in commentsFor(file, section, line)"
                            :key="comment.id"
                            class="comment-row"
                          >
                            <div class="text-xs font-semibold">
                              {{ comment.authorLabel }} commented on line
                              {{ comment.lineNumber }}
                            </div>
                            <SafeHtml :html="comment.bodyHtml" class="mt-2 text-sm" />
                            <button
                              class="mt-2 text-xs text-destructive"
                              type="button"
                              @click="deleteComment(comment)"
                            >
                              Delete
                            </button>
                          </div>
                        </template>
                      </template>
                      <template v-else>
                        <template v-for="row in splitPatch(section)" :key="row.id">
                          <button
                            v-if="row.kind === 'meta'"
                            class="diff-line diff-meta diff-split-meta"
                            type="button"
                          >
                            <code>{{ row.line.text || " " }}</code>
                          </button>
                          <div
                            v-else
                            class="diff-line diff-split-row"
                            :class="row.kind === 'context' ? 'diff-context' : 'diff-pair'"
                          >
                            <button
                              type="button"
                              class="split-side"
                              :class="
                                row.kind === 'context'
                                  ? 'side-context'
                                  : row.left
                                    ? 'side-del'
                                    : 'side-empty'
                              "
                              :disabled="row.kind === 'pair' && !row.left"
                              @click="
                                row.kind === 'context'
                                  ? openComment(file, section, row.line)
                                  : row.left && openComment(file, section, row.left)
                              "
                            >
                              <span class="line-no">{{
                                (row.kind === "context" ? row.line.oldLine : row.left?.oldLine) ??
                                ""
                              }}</span>
                              <code
                                v-if="row.kind === 'context'"
                                v-html="highlightedLine(row.line.text || ' ', file)"
                              />
                              <code
                                v-else-if="row.left"
                                v-html="highlightedLine(row.left.text || ' ', file)"
                              />
                              <code v-else>&nbsp;</code>
                            </button>
                            <button
                              type="button"
                              class="split-side"
                              :class="
                                row.kind === 'context'
                                  ? 'side-context'
                                  : row.right
                                    ? 'side-add'
                                    : 'side-empty'
                              "
                              :disabled="row.kind === 'pair' && !row.right"
                              @click="
                                row.kind === 'context'
                                  ? openComment(file, section, row.line)
                                  : row.right && openComment(file, section, row.right)
                              "
                            >
                              <span class="line-no">{{
                                (row.kind === "context" ? row.line.newLine : row.right?.newLine) ??
                                ""
                              }}</span>
                              <code
                                v-if="row.kind === 'context'"
                                v-html="highlightedLine(row.line.text || ' ', file)"
                              />
                              <code
                                v-else-if="row.right"
                                v-html="highlightedLine(row.right.text || ' ', file)"
                              />
                              <code v-else>&nbsp;</code>
                            </button>
                            <MessageSquare class="comment-icon h-3.5 w-3.5" />
                          </div>
                          <template
                            v-for="line in splitRowCommentLines(row)"
                            :key="`${row.id}:${line.id}:comments`"
                          >
                            <div
                              v-for="comment in commentsFor(file, section, line)"
                              :key="comment.id"
                              class="comment-row"
                            >
                              <div class="text-xs font-semibold">
                                {{ comment.authorLabel }} commented on line
                                {{ comment.lineNumber }}
                              </div>
                              <SafeHtml :html="comment.bodyHtml" class="mt-2 text-sm" />
                              <button
                                class="mt-2 text-xs text-destructive"
                                type="button"
                                @click="deleteComment(comment)"
                              >
                                Delete
                              </button>
                            </div>
                          </template>
                        </template>
                      </template>
                    </div>
                  </section>
                </div>
              </article>
            </template>
          </section>
        </ResizablePanel>
      </template>

      <template v-if="reviewPanelOpen">
        <ResizableHandle with-handle />
        <ResizablePanel
          :default-size="rightPanelDefault"
          :min-size="15"
          :max-size="50"
          class="flex min-h-0 flex-col overflow-hidden border-l border-border bg-background"
        >
          <div class="flex items-center justify-between border-b border-border px-4 py-3">
            <div class="text-sm font-semibold">Conversation</div>
            <button class="icon-btn" type="button" title="Close" @click="closeReviewPanel">
              <X class="h-4 w-4" />
            </button>
          </div>
          <div class="min-h-0 flex-1 overflow-auto">
            <div class="border-b border-border p-4">
              <div class="text-sm font-semibold">Review Notes</div>
              <p class="mt-1 text-xs text-muted-foreground">
                Local-only review timeline and rich comments.
              </p>
              <LazyRichTextEditor
                v-model="summaryDraft"
                class="mt-3"
                min-height="8rem"
                placeholder="Write a review summary…"
                aria-label="Review summary"
                :features="commentFeatures"
              />
              <button
                class="mt-3 w-full toolbar-btn justify-center"
                type="button"
                @click="startReview"
              >
                Start review
              </button>
            </div>

            <div v-if="commentTarget" class="border-b border-border p-4">
              <div class="text-sm font-semibold">
                Comment on {{ commentTarget.file.path }}:{{ commentTarget.line }}
              </div>
              <LazyRichTextEditor
                v-model="commentDraft"
                class="mt-3"
                min-height="8rem"
                placeholder="Leave a comment…"
                aria-label="Line comment"
                autofocus
                :features="commentFeatures"
              />
              <div class="mt-3 flex gap-2">
                <button class="toolbar-btn" type="button" @click="saveComment">Save comment</button>
                <button class="toolbar-btn" type="button" @click="commentTarget = null">
                  Cancel
                </button>
              </div>
            </div>

            <div class="p-4">
              <div class="mb-2 text-sm font-semibold">Activity</div>
              <div v-if="reviews.length === 0" class="text-sm text-muted-foreground">
                No reviews yet.
              </div>
              <button
                v-for="review in reviews"
                :key="review.id"
                class="mb-2 w-full rounded-md border border-border p-3 text-left text-sm hover:bg-muted"
                type="button"
                @click="selectReview(review)"
              >
                <div class="font-medium">{{ review.title }}</div>
                <div class="text-xs text-muted-foreground">{{ formatDate(review.startedAt) }}</div>
              </button>

              <div v-if="activeReview" class="mt-4 space-y-3">
                <div class="rounded-md border border-border p-3">
                  <div class="text-sm font-semibold">{{ activeReview.review.title }}</div>
                  <SafeHtml
                    v-if="activeReview.review.summary"
                    :html="activeReview.review.summary"
                    class="mt-2 text-sm"
                  />
                </div>
                <div
                  v-for="event in activeReview.events"
                  :key="event.id"
                  class="border-l-2 border-border pl-3 text-sm"
                >
                  <div class="font-medium">{{ event.message || event.type }}</div>
                  <div class="text-xs text-muted-foreground">{{ formatDate(event.createdAt) }}</div>
                </div>
              </div>
            </div>
          </div>
        </ResizablePanel>
      </template>
    </ResizablePanelGroup>
  </div>
</template>
