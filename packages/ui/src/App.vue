<script setup lang="ts">
import { computed, ref, watch } from "vue";
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
import type { RichTextFeatures } from "@ui/rich-text-editor";
import type { DiffViewMode } from "@git-diff/contracts";
import { PREF_KEYS } from "@git-diff/contracts";
import { ensureLanguage, languageFor } from "@lib/highlight";
import { ACCENTS, resolveAccent } from "@lib/accent";
import { TWEAK_DEFAULTS, tweakPrefPatch, useTweaks, type Tweaks } from "@composables/useTweaks";
import { useToasts } from "@composables/useToasts";
import { useStyleWatchers } from "@composables/useStyleWatchers";
import { useDiffNavigation } from "@composables/useDiffNavigation";
import { resetLazyRender } from "@composables/useLazyRender";
import { useWalkthrough } from "@composables/useWalkthrough";
import { useCommits } from "@composables/useCommits";
import { usePullRequests } from "@composables/usePullRequests";
import { useRepositoryList } from "@composables/useRepositoryList";
import { useReviewSession } from "@composables/useReviewSession";
import { useRepoBrowsing } from "@composables/useRepoBrowsing";
import { useAppSession } from "@composables/useAppSession";
import { useSelectedFile } from "@composables/useSelectedFile";
import { usePreferences } from "@composables/usePreferences";
import { useDiffLayout } from "@composables/useDiffLayout";
import { useRepoSelectors } from "@composables/useRepoSelectors";
import { useAppCommands } from "@composables/useAppCommands";
import CommandPalette from "@entry/components/CommandPalette.vue";
import { storeToRefs } from "pinia";
import { useAuthStore } from "@/stores/auth.store";
import { useRepoStore } from "@/stores/repo.store";
import { useReviewsStore } from "@/stores/reviews.store";

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
    commentTarget,
    commentDraft,
    commentDialogOpen,
    loadPendingComments,
    copyReviewState,
    startReview,
    toggleViewed,
    isViewed,
    openCommentForLine,
    saveComment,
    cancelComment,
    deleteComment,
    resolveComment,
    replyToComment,
    copyReviewAsMarkdown,
} = useReviewSession({
    state,
    activeReview,
    reviews,
    summaryDraft,
    reviewPanelOpen,
    currentUser,
    prefValues,
    savePreferences,
    onError: (message) => {
        error.value = message;
    },
});
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

const {
    files,
    changedByPath,
    changedPathsSet,
    repoPaths,
    selectedFile,
    selectedIsChanged,
    reviewComments,
    userInitials,
    changedIndex,
    threadsForFile,
} = useRepoSelectors({ state, selectedPath, activeReview, currentUser });

// Reset viewport-deferred render bookkeeping whenever the diff context changes
// so a previous search/navigation "render all" doesn't defeat lazy loading.
watch(
    () => [state.value?.root, state.value?.commitSha, state.value?.mode],
    () => resetLazyRender(),
);

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

useAppCommands({
    hideWhitespace: () => hideWhitespace.value,
    selectedFile,
    savePreferences,
    copyReviewAsMarkdown: () => copyReviewAsMarkdown(),
    generateWalkthrough: () => generateWalkthrough(),
    copyPath,
});

const shortcutsEnabled = computed(() => authMode.value === "ready");
const {
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
} = useRepoBrowsing({
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
    withBusy: (fn) => repoStore.withBusy(fn),
    savePreferences,
    refreshRepositoryList,
    removeRepositoryFromList,
    loadPendingComments,
    resetSelectedFile,
    loadSelectedFile,
    showToast,
    onError: (message) => {
        error.value = message;
    },
});

const { handleAuthCompleted, handleAuthWiped, logOut } = useAppSession({
    bootstrapAuthStore: () => authStore.bootstrap(),
    authBootstrapError: () => authStore.bootstrapError,
    completeAuth: (response) => authStore.complete(response),
    markAuthWiped: () => authStore.markWiped(),
    logoutAuth: () => authStore.logout(),
    state,
    activeRepoPath,
    reviews,
    activeReview,
    selectedPath,
    repositories,
    searchOpen,
    selectedFile,
    changedByPath,
    lastRepoRoot,
    shortcutsEnabled,
    loadPreferences,
    refreshRepositoryList,
    resetPreferences,
    resetSelectedFile,
    generateWalkthrough: () => generateWalkthrough(),
    openRepo,
    openPullRequest,
    openCommit,
    selectAdjacent,
    jumpToHunk,
    toggleViewed,
    startReview,
    onError: (message) => {
        error.value = message;
    },
});

async function updateTweak<K extends keyof Tweaks>(key: K, value: Tweaks[K]) {
    await savePreferences(tweakPrefPatch(key, value));
}

function copyPath(path: string) {
    if (path) {
        void navigator.clipboard.writeText(path);
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
