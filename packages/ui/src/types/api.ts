export type GitFileStatus = "added" | "deleted" | "modified" | "renamed" | "untracked";
export type DiffViewMode = "split" | "unified";

export interface DiffSection {
  id: string;
  kind: "staged" | "unstaged" | "untracked";
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
  branch: string;
  headSha: string;
  generatedAt: string;
  files: ChangedFile[];
  // Every file under the repo root that is tracked or untracked-not-ignored. Ignored files excluded.
  trackedFiles?: string[];
  additions: number;
  deletions: number;
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
}

export interface ReviewDetail {
  review: ReviewSession;
  events: ReviewEvent[];
  comments: ReviewComment[];
}

export interface Repository {
  path: string;
  name: string;
  addedAt: string;
  lastOpenedAt?: string;
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
export type UIAccent = "indigo" | "emerald" | "amber" | "rose";

export const VIEWED_KEY_PREFIX = "file.viewed.";

export function viewedPrefKey(repoRoot: string, filePath: string): string {
  return `${VIEWED_KEY_PREFIX}${repoRoot}.${filePath}`;
}

export interface DiffAppApi {
  repositoryState(path?: string): Promise<RepositoryState>;
  openRepository(path: string): Promise<RepositoryState>;
  refreshRepository(path: string): Promise<RepositoryState>;
  readRepositoryFile(root: string, path: string): Promise<RepositoryFile>;
  listBranches(path?: string): Promise<{ branches: string[] }>;
  checkoutBranch(path: string, branch: string): Promise<RepositoryState>;
  createBranch(path: string, name: string): Promise<RepositoryState>;
  chooseRepository(defaultPath?: string): Promise<string | null>;
  listRepositories(): Promise<Repository[]>;
  upsertRepository(request: { path: string; name?: string }): Promise<Repository>;
  removeRepository(path: string): Promise<void>;
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
