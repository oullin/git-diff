import type {
  AuthBootstrapResponse,
  AuthLoginResponse,
  AuthStateResponse,
  Branch,
  CommitSummary,
  FileSearchResult,
  LaunchIntent,
  PendingComment,
  PullRequestSummary,
  Repository,
  RepositoryCollaborator,
  RepositoryFile,
  RepositoryFileRange,
  RepositoryMode,
  RepositoryState,
  ReviewComment,
  ReviewDetail,
  ReviewEvent,
  ReviewSession,
  SystemStats,
  UserPreferences,
  UserConfig,
  WalkthroughRecord,
} from "@git-diff/domain";

export interface DiffAppApi {
  takeLaunchIntent(): Promise<LaunchIntent | null>;
  onLaunchIntent(handler: (intent: LaunchIntent) => void): () => void;
  openNewWindow(repoPath?: string): Promise<void>;
  repositoryState(path?: string): Promise<RepositoryState>;
  openRepository(path: string): Promise<RepositoryState>;
  refreshRepository(path: string): Promise<RepositoryState>;
  readCommit(sha: string, path?: string): Promise<RepositoryState>;
  listCommits(path?: string, limit?: number): Promise<{ commits: CommitSummary[] }>;
  generateWalkthrough(request: {
    path?: string;
    kind?: RepositoryMode;
    sha?: string;
    refresh?: boolean;
  }): Promise<WalkthroughRecord>;
  listPullRequests(path?: string, limit?: number): Promise<{ pullRequests: PullRequestSummary[] }>;
  readPullRequest(number: number, path?: string): Promise<RepositoryState>;
  listPendingComments(request: {
    path?: string;
    kind?: RepositoryMode;
    sha?: string;
  }): Promise<{ comments: PendingComment[] }>;
  createPendingComment(request: {
    repoRoot: string;
    contextKind: RepositoryMode;
    contextSha?: string;
    filePath: string;
    diffSection: string;
    side: string;
    lineNumber: number;
    startLineNumber?: number;
    startSide?: string;
    authorLabel: string;
    bodyHtml: string;
  }): Promise<PendingComment>;
  updatePendingComment(request: { id: number; bodyHtml: string }): Promise<PendingComment>;
  deletePendingComment(id: number): Promise<void>;
  promotePendingComments(reviewId: number): Promise<{ promoted: number }>;
  readRepositoryFile(root: string, path: string): Promise<RepositoryFile>;
  readRepositoryFileRange(request: {
    root: string;
    path: string;
    ref?: string;
    startLine: number;
    endLine: number;
  }): Promise<RepositoryFileRange>;
  /** Fetches raw blob bytes for inline image rendering. ref="" reads the
   *  working tree; ref=":0" reads the index; otherwise it's a git ref. */
  readRepositoryFileBytes(request: {
    root?: string;
    path: string;
    ref?: string;
  }): Promise<{ data: Uint8Array; mime: string }>;
  listBranches(path?: string): Promise<{ branches: string[]; records?: Branch[] }>;
  checkoutBranch(path: string, branch: string): Promise<RepositoryState>;
  createBranch(path: string, name: string): Promise<RepositoryState>;
  deleteBranch(name: string, path?: string): Promise<void>;
  lockBranch(name: string, path?: string): Promise<{ branches: Branch[] }>;
  unlockBranch(name: string, path?: string): Promise<{ branches: Branch[] }>;
  chooseRepository(defaultPath?: string): Promise<string | null>;
  listRepositories(): Promise<Repository[]>;
  searchRepositoryFiles(query: string, limit?: number): Promise<FileSearchResult[]>;
  upsertRepository(request: { path: string; name?: string }): Promise<Repository>;
  removeRepository(path: string): Promise<void>;
  listCollaborators(path: string): Promise<RepositoryCollaborator[]>;
  addCollaborator(request: {
    path: string;
    userId: number;
    role: "write" | "read";
  }): Promise<RepositoryCollaborator>;
  removeCollaborator(request: { path: string; userId: number }): Promise<void>;
  createReview(request: Partial<ReviewSession>): Promise<ReviewSession>;
  listReviews(limit?: number): Promise<{ reviews: ReviewSession[] }>;
  reviewDetail(id: number): Promise<ReviewDetail>;
  addReviewEvent(request: {
    reviewId: number;
    type: string;
    filePath?: string;
    message?: string;
    metadata?: string;
  }): Promise<ReviewEvent>;
  createReviewComment(request: {
    reviewId: number;
    filePath: string;
    diffSection: string;
    side: string;
    lineNumber: number;
    startLineNumber?: number;
    startSide?: string;
    authorLabel: string;
    bodyHtml: string;
  }): Promise<ReviewComment>;
  updateReviewComment(request: {
    reviewId: number;
    commentId: number;
    bodyHtml: string;
  }): Promise<ReviewComment>;
  deleteReviewComment(request: { reviewId: number; commentId: number }): Promise<void>;
  setReviewCommentResolved(request: {
    reviewId: number;
    commentId: number;
    resolved: boolean;
  }): Promise<ReviewComment>;
  getUIPreferences(): Promise<UserPreferences>;
  saveUIPreferences(patch: Record<string, string>): Promise<UserPreferences>;
  getUserConfig(): Promise<UserConfig>;
  /** Opens ~/.git-diff/config.yaml in the OS default editor. The path is
   *  resolved in the main process from the backend — the renderer cannot
   *  pass an arbitrary path. */
  openUserConfigFile(): Promise<void>;
  getAuthState(): Promise<AuthStateResponse>;
  authBootstrap(): Promise<AuthBootstrapResponse>;
  authSetup(password: string): Promise<AuthLoginResponse>;
  authLogin(password: string, remember: boolean): Promise<AuthLoginResponse>;
  authLogout(): Promise<void>;
  authWipe(osUsername?: string): Promise<void>;
  openDevTools(): Promise<void>;
  getSystemStats(): Promise<SystemStats>;
}

declare global {
  interface Window {
    diffApp: DiffAppApi;
  }
}
