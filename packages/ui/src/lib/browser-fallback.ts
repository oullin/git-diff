import type {
  DiffAppApi,
  Repository,
  RepositoryState,
  ReviewComment,
  ReviewDetail,
  ReviewSession,
  UserPreferences,
} from "@api";

export function installBrowserFallback() {
  if (window.diffApp) {
    return;
  }

  let preferences: UserPreferences = {
    theme: "dark",
    diffViewMode: "split",
    hideWhitespace: false,
    lastRepoRoot: "/Users/local/project",
  };
  let reviews: ReviewDetail[] = [];
  let repositories: Repository[] = [
    {
      path: "/Users/local/project",
      name: "project",
      addedAt: new Date().toISOString(),
      lastOpenedAt: new Date().toISOString(),
    },
  ];
  const state: RepositoryState = {
    root: "/Users/local/project",
    launchPath: "/Users/local/project",
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
    repositoryState: async () => state,
    openRepository: async () => state,
    refreshRepository: async () => ({ ...state, generatedAt: new Date().toISOString() }),
    readRepositoryFile: async (_root: string, path: string) => ({
      path,
      content: `// ${path}\n// Preview not available in browser fallback.\n`,
      binary: false,
      truncated: false,
      size: 0,
    }),
    chooseRepository: async () => state.root,
    listRepositories: async () => repositories,
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
        addedAt: now,
        lastOpenedAt: now,
      };
      repositories = [repo, ...repositories];
      return repo;
    },
    removeRepository: async (path) => {
      repositories = repositories.filter((repo) => repo.path !== path);
    },
    createReview: async (request) => {
      const review: ReviewSession = {
        id: `review-${Date.now()}`,
        repoRoot: request.repoRoot ?? state.root,
        branch: request.branch ?? state.branch,
        headSha: request.headSha ?? state.headSha,
        status: "open",
        title: request.title ?? "Local review",
        summary: request.summary ?? "",
        filesChanged: request.filesChanged ?? state.files.length,
        additions: request.additions ?? state.additions,
        deletions: request.deletions ?? state.deletions,
        startedAt: new Date().toISOString(),
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
    getUserPreferences: async () => preferences,
    saveUserPreferences: async (next) => {
      preferences = { ...preferences, ...next };
      return preferences;
    },
    openDevTools: async () => {},
  };

  window.diffApp = api;
}
