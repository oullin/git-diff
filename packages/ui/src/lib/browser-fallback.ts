import type {
  AuthBootstrapResponse,
  AuthLoginResponse,
  AuthStateResponse,
  AuthUser,
  DiffAppApi,
  Repository,
  RepositoryState,
  ReviewComment,
  ReviewDetail,
  ReviewSession,
  UIPreferences,
} from "@api";
import { PREF_KEYS } from "@api";

export function installBrowserFallback() {
  if (window.diffApp) {
    return;
  }

  let preferences: UIPreferences = {
    values: {
      [PREF_KEYS.theme]: "system",
      [PREF_KEYS.diffViewMode]: "split",
      [PREF_KEYS.lastRepoRoot]: "/Users/local/project",
    },
  };
  const fallbackUser: AuthUser = { id: 1, osUsername: "local", displayName: "local" };
  let reviews: ReviewDetail[] = [];
  let repositories: Repository[] = [
    {
      path: "/Users/local/project",
      name: "project",
      ownerId: fallbackUser.id,
      role: "owner",
      addedAt: new Date().toISOString(),
      lastOpenedAt: new Date().toISOString(),
    },
  ];
  const state: RepositoryState = {
    root: "/Users/local/project",
    launchPath: "/Users/local/project",
    mode: "working",
    branch: "feature/local-review",
    headSha: "abc1234",
    generatedAt: new Date().toISOString(),
    additions: 6,
    deletions: 2,
    files: [
      {
        path: "src/App.vue",
        status: "modified",
        additions: 6,
        deletions: 2,
        binary: false,
        fingerprint: "demo-app",
        sections: [
          {
            id: "unstaged:src/App.vue",
            kind: "unstaged",
            binary: false,
            patch:
              "diff --git a/src/App.vue b/src/App.vue\n--- a/src/App.vue\n+++ b/src/App.vue\n@@ -1,6 +1,10 @@\n <script setup lang=\"ts\">\n-const name = 'old'\n+const name = 'reviewer'\n+const ready = true\n </script>\n \n <template>\n-  <main>{{ name }}</main>\n+  <main>\n+    <strong>{{ name }}</strong>\n+    <span v-if=\"ready\">Ready</span>\n+  </main>\n </template>\n",
          },
        ],
      },
    ],
  };

  const api: DiffAppApi = {
    takeLaunchIntent: async () => null,
    onLaunchIntent: () => () => {},
    repositoryState: async () => state,
    openRepository: async () => state,
    refreshRepository: async () => ({ ...state, generatedAt: new Date().toISOString() }),
    readCommit: async (sha: string) => ({
      ...state,
      mode: "commit",
      commitSha: sha,
      generatedAt: new Date().toISOString(),
    }),
    listCommits: async () => ({
      commits: [
        {
          sha: "abc1234abc1234abc1234abc1234abc1234abc12",
          shortSha: "abc1234",
          author: fallbackUser.displayName,
          email: `${fallbackUser.osUsername}@example.com`,
          date: new Date().toISOString(),
          subject: "Demo commit",
        },
      ],
    }),
    readRepositoryFile: async (_root: string, path: string) => ({
      path,
      content: `// ${path}\n// Preview not available in browser fallback.\n`,
      binary: false,
      truncated: false,
      size: 0,
    }),
    listBranches: async () => ({ branches: [state.branch] }),
    checkoutBranch: async () => state,
    createBranch: async (_path: string, name: string) => ({ ...state, branch: name }),
    deleteBranch: async () => {},
    lockBranch: async () => ({ branches: [] }),
    unlockBranch: async () => ({ branches: [] }),
    chooseRepository: async () => state.root,
    listRepositories: async () => repositories,
    searchRepositoryFiles: async () => [],
    upsertRepository: async ({ path, name }) => {
      const existing = repositories.find((repo) => repo.path === path);
      const now = new Date().toISOString();
      if (existing) {
        existing.lastOpenedAt = now;
        if (name) existing.name = name;
        return existing;
      }
      const repo: Repository = {
        path,
        name: name || path.split("/").filter(Boolean).pop() || path,
        ownerId: fallbackUser.id,
        role: "owner",
        addedAt: now,
        lastOpenedAt: now,
      };
      repositories = [repo, ...repositories];
      return repo;
    },
    removeRepository: async (path) => {
      repositories = repositories.filter((repo) => repo.path !== path);
    },
    listCollaborators: async () => [],
    addCollaborator: async ({ userId, role }) => ({
      userId,
      osUsername: `user-${userId}`,
      displayName: `user-${userId}`,
      role,
      grantedAt: new Date().toISOString(),
    }),
    removeCollaborator: async () => {},
    createReview: async (request) => {
      const review: ReviewSession = {
        id: `review-${Date.now()}`,
        repoRoot: request.repoRoot ?? state.root,
        userId: fallbackUser.id,
        branch: request.branch ?? state.branch,
        headSha: request.headSha ?? state.headSha,
        status: "open",
        title: request.title ?? "Local review",
        summary: request.summary ?? "",
        filesChanged: request.filesChanged ?? state.files.length,
        additions: request.additions ?? state.additions,
        deletions: request.deletions ?? state.deletions,
        startedAt: new Date().toISOString(),
        contextKind: request.contextKind ?? "working",
        contextSha: request.contextSha,
      };
      reviews = [{ review, events: [], comments: [] }, ...reviews];
      return review;
    },
    listReviews: async (limit = 50) => ({
      reviews: reviews.map((item) => item.review).slice(0, limit),
    }),
    reviewDetail: async (id) => {
      const detail = reviews.find((item) => item.review.id === id);
      if (detail) {
        return detail;
      }
      throw new Error("Review not found");
    },
    addReviewEvent: async (request) => ({
      id: Date.now(),
      reviewId: request.reviewId,
      type: request.type,
      filePath: request.filePath,
      message: request.message,
      metadata: request.metadata ?? "{}",
      createdAt: new Date().toISOString(),
    }),
    createReviewComment: async (request) => {
      const comment: ReviewComment = {
        id: `comment-${Date.now()}`,
        reviewId: request.reviewId,
        filePath: request.filePath,
        diffSection: request.diffSection,
        side: request.side,
        lineNumber: request.lineNumber,
        authorLabel: request.authorLabel,
        bodyHtml: request.bodyHtml,
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      };
      const detail = reviews.find((item) => item.review.id === request.reviewId);
      detail?.comments.push(comment);
      return comment;
    },
    updateReviewComment: async (request) => {
      for (const detail of reviews) {
        const comment = detail.comments.find((item) => item.id === request.commentId);
        if (comment) {
          comment.bodyHtml = request.bodyHtml;
          comment.updatedAt = new Date().toISOString();
          return comment;
        }
      }
      throw new Error("Comment not found");
    },
    deleteReviewComment: async (request) => {
      const detail = reviews.find((item) => item.review.id === request.reviewId);
      if (detail) {
        detail.comments = detail.comments.filter((comment) => comment.id !== request.commentId);
      }
    },
    getUIPreferences: async () => preferences,
    saveUIPreferences: async (patch) => {
      const values = { ...preferences.values };
      for (const [key, value] of Object.entries(patch)) {
        if (value === "") {
          delete values[key];
        } else {
          values[key] = value;
        }
      }
      preferences = { values, updatedAt: new Date().toISOString() };
      return preferences;
    },
    getAuthState: async (): Promise<AuthStateResponse> => ({
      osUsername: fallbackUser.osUsername,
      needsSetup: false,
      isAuthenticated: true,
    }),
    authBootstrap: async (): Promise<AuthBootstrapResponse> => ({
      user: fallbackUser,
      state: {
        osUsername: fallbackUser.osUsername,
        needsSetup: false,
        isAuthenticated: true,
      },
    }),
    authSetup: async (): Promise<AuthLoginResponse> => ({ user: fallbackUser }),
    authLogin: async (): Promise<AuthLoginResponse> => ({ user: fallbackUser }),
    authLogout: async () => {},
    authWipe: async () => {},
    openDevTools: async () => {},
    getSystemStats: async () => ({
      cpuPercent: 0,
      memoryUsedGB: 0,
      memoryTotalGB: 0,
      loadAvg1: 0,
    }),
  };

  window.diffApp = api;
}
