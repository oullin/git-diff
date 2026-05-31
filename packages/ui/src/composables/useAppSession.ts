import { onMounted, onUnmounted, type ComputedRef, type Ref } from "vue";
import { useKeyboardShortcuts } from "@composables/useKeyboardShortcuts";

import type {
  AuthLoginResponse,
  ChangedFile,
  Repository,
  RepositoryState,
  ReviewDetail,
  ReviewSession,
} from "@git-diff/domain";

/**
 * Owns the application session lifecycle: auth bootstrap, entering the app,
 * applying the CLI launch intent, logout teardown, and the mount-time wiring of
 * keyboard shortcuts + launch-intent subscription. All collaborators are
 * injected so this composable carries no state of its own.
 */
export interface AppSessionOptions {
  // Auth operations.
  bootstrapAuthStore: () => Promise<{ entered: boolean }>;
  authBootstrapError: () => string | null;
  completeAuth: (response: AuthLoginResponse) => void;
  markAuthWiped: () => void;
  logoutAuth: () => Promise<void>;
  // Shared state.
  state: Ref<RepositoryState | null>;
  activeRepoPath: Ref<string>;
  reviews: Ref<ReviewSession[]>;
  activeReview: Ref<ReviewDetail | null>;
  selectedPath: Ref<string>;
  repositories: Ref<Repository[]>;
  searchOpen: Ref<boolean>;
  selectedFile: ComputedRef<ChangedFile | null>;
  changedByPath: ComputedRef<Map<string, ChangedFile>>;
  lastRepoRoot: ComputedRef<string>;
  shortcutsEnabled: ComputedRef<boolean>;
  // Collaborators.
  loadPreferences: () => Promise<void>;
  refreshRepositoryList: () => Promise<void>;
  resetPreferences: () => void;
  resetSelectedFile: () => void;
  generateWalkthrough: () => Promise<void>;
  openRepo: (path: string) => Promise<void>;
  openPullRequest: (number: number) => Promise<void>;
  openCommit: (sha: string) => Promise<void>;
  selectAdjacent: (delta: number) => void;
  jumpToHunk: (delta: number) => void;
  toggleViewed: (file: ChangedFile) => void | Promise<void>;
  startReview: () => void | Promise<void>;
  onError: (message: string) => void;
}

export function useAppSession(opts: AppSessionOptions) {
  let unsubscribeShortcuts: (() => void) | null = null;
  let unsubscribeLaunchIntent: (() => void) | null = null;

  async function bootstrapAuth() {
    const { entered } = await opts.bootstrapAuthStore();

    const bootstrapError = opts.authBootstrapError();

    if (bootstrapError) {
      opts.onError(bootstrapError);
    }

    if (entered) {
      await enterApp();
    }
  }

  async function enterApp() {
    await opts.loadPreferences();

    await opts.refreshRepositoryList();

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
      opts.repositories.value.find((repo) => repo.path === opts.lastRepoRoot.value)?.path ??
      opts.repositories.value[0]?.path ??
      "";

    if (!initialPath) {
      return;
    }

    await opts.openRepo(initialPath);

    const prNumber = intent?.pullRequestNumber ?? intent?.prNumber;
    const commitRef = intent?.commitRef ?? intent?.sha;

    if (intent?.kind === "pull-request" && prNumber && opts.state.value) {
      await opts.openPullRequest(prNumber);
    } else if (commitRef && opts.state.value) {
      await opts.openCommit(commitRef);
    }

    if (intent?.walkthrough) {
      await opts.generateWalkthrough();
    }
  }

  async function handleAuthCompleted(response: AuthLoginResponse) {
    opts.completeAuth(response);

    await enterApp();
  }

  async function handleAuthWiped() {
    opts.markAuthWiped();
  }

  async function logOut() {
    await opts.logoutAuth();

    opts.resetPreferences();
    opts.state.value = null;
    opts.repositories.value = [];
    opts.activeRepoPath.value = "";
    opts.reviews.value = [];
    opts.activeReview.value = null;
    opts.selectedPath.value = "";
    opts.resetSelectedFile();

    await bootstrapAuth();
  }

  onMounted(async () => {
    await bootstrapAuth();

    unsubscribeShortcuts = useKeyboardShortcuts({
      enabled: opts.shortcutsEnabled,
      onSelectAdjacent: opts.selectAdjacent,
      onJumpToHunk: opts.jumpToHunk,
      onToggleViewed: () => {
        const file = opts.selectedFile.value;

        if (file && opts.changedByPath.value.has(file.path)) {
          void opts.toggleViewed(file);
        }
      },
      onStartReview: () => void opts.startReview(),
      onOpenSearch: () => {
        opts.searchOpen.value = true;
      },
    });
    unsubscribeLaunchIntent = window.diffApp.onLaunchIntent(async (intent) => {
      if (intent.kind === "help" || !intent.repoPath) {
        return;
      }

      await opts.openRepo(intent.repoPath);

      const prNumber = intent.pullRequestNumber ?? intent.prNumber;
      const commitRef = intent.commitRef ?? intent.sha;

      if (intent.kind === "pull-request" && prNumber && opts.state.value) {
        await opts.openPullRequest(prNumber);
      } else if (commitRef && opts.state.value) {
        await opts.openCommit(commitRef);
      }
    });
  });

  onUnmounted(() => {
    unsubscribeShortcuts?.();
    unsubscribeLaunchIntent?.();
  });

  return { handleAuthCompleted, handleAuthWiped, logOut };
}
