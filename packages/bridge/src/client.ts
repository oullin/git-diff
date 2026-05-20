import type {
  AuthLoginResponse,
  AuthStateResponse,
  AuthUser,
  Branch,
  CommitSummary,
  FileSearchResult,
  OpItem,
  OpVault,
  PendingComment,
  PullRequestSummary,
  Repository,
  RepositoryCollaborator,
  RepositoryFile,
  RepositoryState,
  ReviewComment,
  ReviewDetail,
  ReviewEvent,
  ReviewSession,
  RunLog,
  RunSummary,
  RunWorkflowRequest,
  RuntimeSettings,
  SettingsResponse,
  SystemStats,
  TemplateFileContent,
  TemplateFileSummary,
  UIPreferencesResponse,
  WalkthroughRecord,
  Workflow,
} from "@git-diff/contracts";
import { requestJson } from "#bridge/http.js";
import { runWorkflowStream } from "#bridge/sse.js";
import type {
  UnixTarget,
  WorkflowBridgeClient,
  WorkflowRunStream,
} from "#bridge/client-types.js";

type ListRunsRequest = { limit?: number };

export function unixTarget(socketPath: string): UnixTarget {
  return { socketPath };
}

export function createWorkflowBridgeClient(target: UnixTarget): WorkflowBridgeClient {
  return new HttpWorkflowBridgeClient(target);
}

export function waitForReady(client: WorkflowBridgeClient, timeoutMs = 10000): Promise<void> {
  const deadline = Date.now() + timeoutMs;

  return new Promise<void>((resolve, reject) => {
    const attempt = (): void => {
      client
        .healthz()
        .then(resolve)
        .catch((error: unknown) => {
          if (Date.now() >= deadline) {
            reject(error);

            return;
          }

          setTimeout(attempt, 100);
        });
    };

    attempt();
  });
}

class HttpWorkflowBridgeClient implements WorkflowBridgeClient {
  private readonly socketPath: string;

  constructor(target: UnixTarget) {
    this.socketPath = target.socketPath;
  }

  close(): void {}

  healthz(): Promise<void> {
    return this.request<void>("GET", "/v1/healthz");
  }

  listWorkflows(): Promise<{ workflows: Workflow[] }> {
    return this.request<{ workflows: Workflow[] }>("GET", "/v1/workflows");
  }

  listRuns(request: ListRunsRequest = {}): Promise<{ runs: RunSummary[] }> {
    const query = typeof request.limit === "number" ? `?limit=${request.limit}` : "";

    return this.request<{ runs: RunSummary[] }>("GET", `/v1/runs${query}`);
  }

  runLog(request: { runId: string }): Promise<RunLog> {
    return this.request<RunLog>("GET", `/v1/runs/${encodeURIComponent(request.runId)}/log`);
  }

  listTemplateFiles(): Promise<{ files: TemplateFileSummary[] }> {
    return this.request<{ files: TemplateFileSummary[] }>("GET", "/v1/template-files");
  }

  readTemplateFile(request: { path: string }): Promise<TemplateFileContent> {
    return this.request<TemplateFileContent>(
      "GET",
      `/v1/template-files/content?path=${encodeURIComponent(request.path)}`,
    );
  }

  saveTemplateFile(request: { path: string; content: string }): Promise<TemplateFileContent> {
    return this.request<TemplateFileContent>("PUT", "/v1/template-files/content", {
      path: request.path,
      content: request.content,
    });
  }

  getSettings(): Promise<SettingsResponse> {
    return this.request<SettingsResponse>("GET", "/v1/settings");
  }

  validateSettings(request: { settings: RuntimeSettings }): Promise<SettingsResponse> {
    return this.request<SettingsResponse>("POST", "/v1/settings/validate", {
      settings: request.settings,
    });
  }

  getUIPreferences(): Promise<UIPreferencesResponse> {
    return this.request<UIPreferencesResponse>("GET", "/v1/preferences");
  }

  saveUIPreferences(values: Record<string, string>): Promise<UIPreferencesResponse> {
    return this.request<UIPreferencesResponse>("POST", "/v1/preferences", { values });
  }

  getAuthState(): Promise<AuthStateResponse> {
    return this.request<AuthStateResponse>("GET", "/v1/auth/state");
  }

  authSetup(request: { password: string }): Promise<AuthLoginResponse> {
    return this.request<AuthLoginResponse>("POST", "/v1/auth/setup", {
      password: request.password,
    });
  }

  authLogin(request: { password: string; remember: boolean }): Promise<AuthLoginResponse> {
    return this.request<AuthLoginResponse>("POST", "/v1/auth/login", {
      password: request.password,
      remember: request.remember,
    });
  }

  authResume(request: { token: string }): Promise<{ user: AuthUser }> {
    return this.request<{ user: AuthUser }>("POST", "/v1/auth/resume", { token: request.token });
  }

  authLogout(): Promise<void> {
    return this.request<void>("POST", "/v1/auth/logout");
  }

  authWipe(request: { osUsername?: string }): Promise<void> {
    return this.request<void>("POST", "/v1/auth/wipe", { osUsername: request.osUsername ?? "" });
  }

  repositoryState(request: { path?: string }): Promise<RepositoryState> {
    const query = request.path ? `?path=${encodeURIComponent(request.path)}` : "";
    return this.request<RepositoryState>("GET", `/v1/repository/state${query}`);
  }

  openRepository(request: { path: string }): Promise<RepositoryState> {
    return this.request<RepositoryState>("POST", "/v1/repository/open", { path: request.path });
  }

  refreshRepository(request: { path: string }): Promise<RepositoryState> {
    return this.request<RepositoryState>("POST", "/v1/repository/refresh", { path: request.path });
  }

  readCommit(request: { path?: string; sha: string }): Promise<RepositoryState> {
    const parts = [`sha=${encodeURIComponent(request.sha)}`];
    if (request.path) {
      parts.push(`path=${encodeURIComponent(request.path)}`);
    }
    return this.request<RepositoryState>("GET", `/v1/repository/commit?${parts.join("&")}`);
  }

  listCommits(request: { path?: string; limit?: number }): Promise<{ commits: CommitSummary[] }> {
    const parts: string[] = [];
    if (request.path) {
      parts.push(`path=${encodeURIComponent(request.path)}`);
    }
    if (request.limit) {
      parts.push(`limit=${request.limit}`);
    }
    const query = parts.length === 0 ? "" : `?${parts.join("&")}`;
    return this.request<{ commits: CommitSummary[] }>("GET", `/v1/repository/log${query}`);
  }

  generateWalkthrough(request: {
    path?: string;
    kind?: "working" | "commit";
    sha?: string;
    refresh?: boolean;
  }): Promise<WalkthroughRecord> {
    return this.request<WalkthroughRecord>("POST", "/v1/walkthrough", {
      path: request.path ?? "",
      kind: request.kind ?? "working",
      sha: request.sha ?? "",
      refresh: request.refresh ?? false,
    });
  }

  listPullRequests(request: { path?: string; limit?: number }): Promise<{
    pullRequests: PullRequestSummary[];
  }> {
    const parts: string[] = [];
    if (request.path) parts.push(`path=${encodeURIComponent(request.path)}`);
    if (request.limit) parts.push(`limit=${request.limit}`);
    const query = parts.length === 0 ? "" : `?${parts.join("&")}`;
    return this.request<{ pullRequests: PullRequestSummary[] }>(
      "GET",
      `/v1/repository/pull-requests${query}`,
    );
  }

  readPullRequest(request: { path?: string; number: number }): Promise<RepositoryState> {
    const parts = [`number=${request.number}`];
    if (request.path) parts.push(`path=${encodeURIComponent(request.path)}`);
    return this.request<RepositoryState>("GET", `/v1/repository/pull-request?${parts.join("&")}`);
  }

  listPendingComments(request: {
    path?: string;
    kind?: "working" | "commit";
    sha?: string;
  }): Promise<{ comments: PendingComment[] }> {
    const parts: string[] = [];
    if (request.path) parts.push(`path=${encodeURIComponent(request.path)}`);
    if (request.kind) parts.push(`kind=${request.kind}`);
    if (request.sha) parts.push(`sha=${encodeURIComponent(request.sha)}`);
    const query = parts.length === 0 ? "" : `?${parts.join("&")}`;
    return this.request<{ comments: PendingComment[] }>("GET", `/v1/pending-comments${query}`);
  }

  createPendingComment(request: {
    repoRoot: string;
    contextKind: "working" | "commit";
    contextSha?: string;
    filePath: string;
    diffSection: string;
    side: string;
    lineNumber: number;
    authorLabel: string;
    bodyHtml: string;
  }): Promise<PendingComment> {
    return this.request<PendingComment>("POST", "/v1/pending-comments", request);
  }

  updatePendingComment(request: { id: string; bodyHtml: string }): Promise<PendingComment> {
    return this.request<PendingComment>(
      "PATCH",
      `/v1/pending-comments/${encodeURIComponent(request.id)}`,
      {
        bodyHtml: request.bodyHtml,
      },
    );
  }

  deletePendingComment(request: { id: string }): Promise<void> {
    return this.request<void>("DELETE", `/v1/pending-comments/${encodeURIComponent(request.id)}`);
  }

  promotePendingComments(request: { reviewId: string }): Promise<{ promoted: number }> {
    return this.request<{ promoted: number }>("POST", "/v1/pending-comments/promote", request);
  }

  readRepositoryFile(request: { root: string; path: string }): Promise<RepositoryFile> {
    const query = `?root=${encodeURIComponent(request.root)}&path=${encodeURIComponent(request.path)}`;
    return this.request<RepositoryFile>("GET", `/v1/repository/file${query}`);
  }

  listBranches(request: { path?: string }): Promise<{ branches: string[]; records?: Branch[] }> {
    const query = request.path ? `?path=${encodeURIComponent(request.path)}` : "";
    return this.request<{ branches: string[]; records?: Branch[] }>(
      "GET",
      `/v1/repository/branches${query}`,
    );
  }

  checkoutBranch(request: { path: string; branch: string }): Promise<RepositoryState> {
    return this.request<RepositoryState>("POST", "/v1/repository/checkout", {
      path: request.path,
      branch: request.branch,
    });
  }

  createBranch(request: { path: string; name: string }): Promise<RepositoryState> {
    return this.request<RepositoryState>("POST", "/v1/repository/branches/create", {
      path: request.path,
      name: request.name,
    });
  }

  deleteBranch(request: { path?: string; name: string }): Promise<void> {
    const params = new URLSearchParams({ name: request.name });
    if (request.path) params.set("path", request.path);
    return this.request<void>("DELETE", `/v1/repository/branches?${params.toString()}`);
  }

  lockBranch(request: { path?: string; name: string }): Promise<{ branches: Branch[] }> {
    return this.request<{ branches: Branch[] }>("POST", "/v1/repository/branches/lock", {
      path: request.path,
      name: request.name,
    });
  }

  unlockBranch(request: { path?: string; name: string }): Promise<{ branches: Branch[] }> {
    return this.request<{ branches: Branch[] }>("POST", "/v1/repository/branches/unlock", {
      path: request.path,
      name: request.name,
    });
  }

  getSystemStats(): Promise<SystemStats> {
    return this.request<SystemStats>("GET", "/v1/system/stats");
  }

  createReview(request: Partial<ReviewSession>): Promise<ReviewSession> {
    return this.request<ReviewSession>("POST", "/v1/reviews", request as Record<string, unknown>);
  }

  listReviews(request: { limit?: number } = {}): Promise<{ reviews: ReviewSession[] }> {
    const query = typeof request.limit === "number" ? `?limit=${request.limit}` : "";
    return this.request<{ reviews: ReviewSession[] }>("GET", `/v1/reviews${query}`);
  }

  reviewDetail(request: { id: string }): Promise<ReviewDetail> {
    return this.request<ReviewDetail>("GET", `/v1/reviews/${encodeURIComponent(request.id)}`);
  }

  addReviewEvent(request: {
    reviewId: string;
    type: string;
    filePath?: string;
    message?: string;
    metadata?: string;
  }): Promise<ReviewEvent> {
    return this.request<ReviewEvent>(
      "POST",
      `/v1/reviews/${encodeURIComponent(request.reviewId)}/events`,
      {
        type: request.type,
        filePath: request.filePath,
        message: request.message,
        metadata: request.metadata,
      },
    );
  }

  createReviewComment(request: {
    reviewId: string;
    filePath: string;
    diffSection: string;
    side: string;
    lineNumber: number;
    authorLabel: string;
    bodyHtml: string;
  }): Promise<ReviewComment> {
    return this.request<ReviewComment>(
      "POST",
      `/v1/reviews/${encodeURIComponent(request.reviewId)}/comments`,
      {
        filePath: request.filePath,
        diffSection: request.diffSection,
        side: request.side,
        lineNumber: request.lineNumber,
        authorLabel: request.authorLabel,
        bodyHtml: request.bodyHtml,
      },
    );
  }

  updateReviewComment(request: {
    reviewId: string;
    commentId: string;
    bodyHtml: string;
  }): Promise<ReviewComment> {
    return this.request<ReviewComment>(
      "PATCH",
      `/v1/reviews/${encodeURIComponent(request.reviewId)}/comments/${encodeURIComponent(request.commentId)}`,
      { bodyHtml: request.bodyHtml },
    );
  }

  deleteReviewComment(request: { reviewId: string; commentId: string }): Promise<void> {
    return this.request<void>(
      "DELETE",
      `/v1/reviews/${encodeURIComponent(request.reviewId)}/comments/${encodeURIComponent(request.commentId)}`,
    );
  }

  listRepositories(): Promise<{ repositories: Repository[] }> {
    return this.request<{ repositories: Repository[] }>("GET", "/v1/repositories");
  }

  searchRepositoryFiles(request: {
    query: string;
    limit?: number;
  }): Promise<{ results: FileSearchResult[] }> {
    const params = new URLSearchParams({ q: request.query });

    if (typeof request.limit === "number") {
      params.set("limit", String(request.limit));
    }

    return this.request<{ results: FileSearchResult[] }>(
      "GET",
      `/v1/repositories/search-files?${params.toString()}`,
    );
  }

  upsertRepository(request: { path: string; name?: string }): Promise<Repository> {
    return this.request<Repository>("POST", "/v1/repositories", {
      path: request.path,
      name: request.name ?? "",
    });
  }

  removeRepository(request: { path: string }): Promise<void> {
    return this.request<void>(
      "DELETE",
      `/v1/repositories?path=${encodeURIComponent(request.path)}`,
    );
  }

  listCollaborators(request: {
    path: string;
  }): Promise<{ collaborators: RepositoryCollaborator[] }> {
    return this.request<{ collaborators: RepositoryCollaborator[] }>(
      "GET",
      `/v1/repositories/collaborators?path=${encodeURIComponent(request.path)}`,
    );
  }

  addCollaborator(request: {
    path: string;
    userId: number;
    role: "write" | "read";
  }): Promise<RepositoryCollaborator> {
    return this.request<RepositoryCollaborator>("POST", "/v1/repositories/collaborators", {
      path: request.path,
      userId: request.userId,
      role: request.role,
    });
  }

  removeCollaborator(request: { path: string; userId: number }): Promise<void> {
    const params = new URLSearchParams({
      path: request.path,
      userId: String(request.userId),
    });
    return this.request<void>("DELETE", `/v1/repositories/collaborators?${params.toString()}`);
  }

  listOpVaults(): Promise<{ vaults: OpVault[] }> {
    return this.request<{ vaults: OpVault[] }>("GET", "/v1/onepassword/vaults");
  }

  listOpItems(request: { vault: string }): Promise<{ items: OpItem[] }> {
    return this.request<{ items: OpItem[] }>(
      "GET",
      `/v1/onepassword/items?vault=${encodeURIComponent(request.vault)}`,
    );
  }

  runWorkflow(request: RunWorkflowRequest): WorkflowRunStream {
    return runWorkflowStream(this.socketPath, request);
  }

  private request<Response>(
    method: "GET" | "POST" | "PUT" | "PATCH" | "DELETE",
    path: string,
    body?: Record<string, unknown>,
  ): Promise<Response> {
    return requestJson<Response>(this.socketPath, method, path, body);
  }
}
