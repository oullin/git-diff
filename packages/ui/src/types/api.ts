export type GitFileStatus = "added" | "deleted" | "modified" | "renamed" | "untracked";
export type DiffViewMode = "split" | "unified";
export type DiffSectionKind = "staged" | "unstaged" | "untracked" | "commit";
export type RepositoryMode = "working" | "commit";

export interface DiffSection {
  id: string;
  kind: DiffSectionKind;
  patch: string;
  binary: boolean;
}

export interface ChangedFile {
  path: string;
  oldPath?: string;
  status: GitFileStatus;
  additions: number;
  deletions: number;
  binary: boolean;
  fingerprint: string;
  sections: DiffSection[];
}

export interface RepositoryState {
  root: string;
  launchPath: string;
  mode: RepositoryMode;
  branch: string;
  headSha: string;
  commitSha?: string;
  generatedAt: string;
  files: ChangedFile[];
  // Every file under the repo root that is tracked or untracked-not-ignored. Ignored files excluded.
  trackedFiles?: string[];
  additions: number;
  deletions: number;
}

export interface CommitSummary {
  sha: string;
  shortSha: string;
  author: string;
  email: string;
  date: string;
  subject: string;
}

export interface RepositoryFile {
  path: string;
  content: string;
  binary: boolean;
  truncated: boolean;
  size: number;
}

export interface SystemStats {
  cpuPercent: number;
  memoryUsedGB: number;
  memoryTotalGB: number;
  loadAvg1: number;
}

export interface ReviewSession {
  id: string;
  repoRoot: string;
  userId: number;
  branch: string;
  headSha: string;
  status: string;
  title: string;
  summary: string;
  filesChanged: number;
  additions: number;
  deletions: number;
  startedAt: string;
  completedAt?: string;
  contextKind: RepositoryMode;
  contextSha?: string;
}

export interface ReviewEvent {
  id: number;
  reviewId: string;
  type: string;
  filePath?: string;
  message?: string;
  metadata: string;
  createdAt: string;
}

export interface ReviewComment {
  id: string;
  reviewId: string;
  filePath: string;
  diffSection: string;
  side: string;
  lineNumber: number;
  authorLabel: string;
  bodyHtml: string;
  createdAt: string;
  updatedAt: string;
  deletedAt?: string;
}

export interface ReviewDetail {
  review: ReviewSession;
  events: ReviewEvent[];
  comments: ReviewComment[];
}

export type RepositoryRole = "owner" | "write" | "read";

export interface Repository {
  path: string;
  name: string;
  ownerId: number;
  role: RepositoryRole;
  addedAt: string;
  lastOpenedAt?: string;
}

export interface FileSearchResult {
  repoPath: string;
  repoName: string;
  filePath: string;
  score: number;
}

export interface RepositoryCollaborator {
  userId: number;
  osUsername: string;
  displayName: string;
  role: "write" | "read";
  grantedAt: string;
}

export interface Branch {
  name: string;
  locked: boolean;
  lockedBy?: number;
  lockedAt?: string;
  lastSeenAt: string;
}

export interface UIPreferences {
  values: Record<string, string>;
  updatedAt?: string;
}

export interface AuthUser {
  id: number;
  osUsername: string;
  displayName: string;
}

export interface AuthStateResponse {
  osUsername: string;
  needsSetup: boolean;
  isAuthenticated: boolean;
}

export interface AuthLoginResponse {
  token?: string;
  user: AuthUser;
}

export interface AuthBootstrapResponse {
  user: AuthUser | null;
  state: AuthStateResponse;
}

export const PREF_KEYS = {
  theme: "theme",
  diffViewMode: "diff.viewMode",
  diffHideWhitespace: "diff.hideWhitespace",
  diffStyle: "diff.style",
  diffDensity: "diff.density",
  diffWordHi: "diff.wordHi",
  lastRepoRoot: "repo.lastRoot",
  panelLeftWidth: "panel.left.width",
  panelFileTreeWidth: "panel.fileTree.width",
  panelRightWidth: "panel.right.width",
  uiAccent: "ui.accent",
  uiShowMinimap: "ui.minimap",
  uiShowStatusBar: "ui.statusBar",
} as const;

export type DiffHunkStyle = "soft" | "punchy" | "bar";
export type DiffDensity = "comfortable" | "compact";
export type UIAccent = "copper" | "indigo" | "emerald" | "amber" | "rose";

export const VIEWED_KEY_PREFIX = "file.viewed.";

export function viewedPrefKey(repoRoot: string, filePath: string): string {
  return `${VIEWED_KEY_PREFIX}${repoRoot}.${filePath}`;
}

export type LaunchIntentKind = "working" | "commit" | "help";

export interface LaunchIntent {
  kind: LaunchIntentKind;
  repoPath?: string;
  sha?: string;
  walkthrough: boolean;
  helpText?: string;
  raw?: string;
}

export interface DiffAppApi {
  takeLaunchIntent(): Promise<LaunchIntent | null>;
  onLaunchIntent(handler: (intent: LaunchIntent) => void): () => void;
  openNewWindow(repoPath?: string): Promise<void>;
  repositoryState(path?: string): Promise<RepositoryState>;
  openRepository(path: string): Promise<RepositoryState>;
  refreshRepository(path: string): Promise<RepositoryState>;
  readCommit(sha: string, path?: string): Promise<RepositoryState>;
  listCommits(path?: string, limit?: number): Promise<{ commits: CommitSummary[] }>;
  readRepositoryFile(root: string, path: string): Promise<RepositoryFile>;
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
  reviewDetail(id: string): Promise<ReviewDetail>;
  addReviewEvent(request: {
    reviewId: string;
    type: string;
    filePath?: string;
    message?: string;
    metadata?: string;
  }): Promise<ReviewEvent>;
  createReviewComment(request: {
    reviewId: string;
    filePath: string;
    diffSection: string;
    side: string;
    lineNumber: number;
    authorLabel: string;
    bodyHtml: string;
  }): Promise<ReviewComment>;
  updateReviewComment(request: {
    reviewId: string;
    commentId: string;
    bodyHtml: string;
  }): Promise<ReviewComment>;
  deleteReviewComment(request: { reviewId: string; commentId: string }): Promise<void>;
  getUIPreferences(): Promise<UIPreferences>;
  saveUIPreferences(patch: Record<string, string>): Promise<UIPreferences>;
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
