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

export interface UserPreferences {
  theme: string;
  diffViewMode: DiffViewMode;
  hideWhitespace: boolean;
  lastRepoRoot: string;
  updatedAt?: string;
}

export interface DiffAppApi {
  repositoryState(path?: string): Promise<RepositoryState>;
  openRepository(path: string): Promise<RepositoryState>;
  refreshRepository(path: string): Promise<RepositoryState>;
  readRepositoryFile(root: string, path: string): Promise<RepositoryFile>;
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
  getUserPreferences(): Promise<UserPreferences>;
  saveUserPreferences(preferences: Partial<UserPreferences>): Promise<UserPreferences>;
  openDevTools(): Promise<void>;
}

declare global {
  interface Window {
    diffApp: DiffAppApi;
  }
}
