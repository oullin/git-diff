import type { EventEmitter } from "node:events";

export interface Phase {
  id: string;
  name: string;
  enabled: boolean;
}

export interface ConfirmationOption {
  id: string;
  label: string;
  description: string;
  continue: boolean;
  back: boolean;
  requiresApproval: boolean;
  phases: Phase[];
}

export interface Workflow {
  id: string;
  name: string;
  description: string;
  changesMac: string;
  phases: Phase[];
  confirmation?: {
    title: string;
    message: string;
    options: ConfirmationOption[];
  };
}

export interface RunWorkflowRequest {
  workflowId: string;
  confirmationOptionId: string;
  enabledPhaseIds: string[];
}

export interface WorkflowEvent {
  id?: number;
  runId: string;
  seq: number;
  type: string;
  phaseId?: string;
  phaseName?: string;
  status?: string;
  message?: string;
  createdAt?: string;
}

export interface RunSummary {
  id: string;
  workflowId: string;
  workflowName: string;
  confirmationOptionId: string;
  confirmationOptionLabel: string;
  mode: string;
  status: string;
  startedAt: string;
  completedAt?: string;
  errorMessage?: string;
}

export interface RunLog {
  run?: RunSummary;
  events: WorkflowEvent[];
}

export interface TemplateFileSummary {
  path: string;
  relative: string;
  kind: string;
  size: number;
  modifiedAt?: string;
  exists: boolean;
}

export interface TemplateFileContent {
  file: TemplateFileSummary;
  content: string;
}

export interface RuntimeSettings {
  repoRoot: string;
  appsConfigPath: string;
  secretsConfigPath: string;
  generatedAppsPath: string;
  archiveRoot: string;
  workflowDbPath: string;
  opVault: string;
  opItem: string;
}

export interface SettingsCheck {
  key: string;
  label: string;
  path: string;
  status: string;
  message: string;
}

export interface SettingsResponse {
  settings?: RuntimeSettings;
  checks: SettingsCheck[];
  valid: boolean;
}

export interface UIPreferencesResponse {
  values: Record<string, string>;
  updatedAt: string;
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

export type GitFileStatus = "added" | "deleted" | "modified" | "renamed" | "untracked";

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

export type RepositoryRole = "owner" | "write" | "read";

export interface Repository {
  path: string;
  name: string;
  ownerId: number;
  role: RepositoryRole;
  addedAt: string;
  lastOpenedAt?: string;
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

export interface OpVault {
  id: string;
  name: string;
}

export interface OpItem {
  id: string;
  title: string;
}

export interface OpUnavailableError extends Error {
  code: "op_unavailable";
}

export interface UnixTarget {
  socketPath: string;
}

export interface WorkflowRunEndInfo {
  exitCode: number;
  status: string;
  message?: string;
}

export interface WorkflowRunStream extends EventEmitter {
  on(event: "data", listener: (event: WorkflowEvent) => void): this;
  on(event: "end", listener: () => void): this;
  on(event: "end-info", listener: (info: WorkflowRunEndInfo) => void): this;
  on(event: "error", listener: (error: Error) => void): this;
}

export interface WorkflowBridgeClient {
  close(): void;
  healthz(): Promise<void>;
  listWorkflows(): Promise<{ workflows: Workflow[] }>;
  listRuns(request: { limit: number }): Promise<{ runs: RunSummary[] }>;
  runLog(request: { runId: string }): Promise<RunLog>;
  listTemplateFiles(): Promise<{ files: TemplateFileSummary[] }>;
  readTemplateFile(request: { path: string }): Promise<TemplateFileContent>;
  saveTemplateFile(request: { path: string; content: string }): Promise<TemplateFileContent>;
  getSettings(): Promise<SettingsResponse>;
  validateSettings(request: { settings: RuntimeSettings }): Promise<SettingsResponse>;
  getUIPreferences(): Promise<UIPreferencesResponse>;
  saveUIPreferences(values: Record<string, string>): Promise<UIPreferencesResponse>;
  getAuthState(): Promise<AuthStateResponse>;
  authSetup(request: { password: string }): Promise<AuthLoginResponse>;
  authLogin(request: { password: string; remember: boolean }): Promise<AuthLoginResponse>;
  authResume(request: { token: string }): Promise<{ user: AuthUser }>;
  authLogout(): Promise<void>;
  authWipe(request: { osUsername?: string }): Promise<void>;
  repositoryState(request: { path?: string }): Promise<RepositoryState>;
  openRepository(request: { path: string }): Promise<RepositoryState>;
  refreshRepository(request: { path: string }): Promise<RepositoryState>;
  readRepositoryFile(request: { root: string; path: string }): Promise<RepositoryFile>;
  listBranches(request: { path?: string }): Promise<{ branches: string[]; records?: Branch[] }>;
  checkoutBranch(request: { path: string; branch: string }): Promise<RepositoryState>;
  createBranch(request: { path: string; name: string }): Promise<RepositoryState>;
  deleteBranch(request: { path?: string; name: string }): Promise<void>;
  lockBranch(request: { path?: string; name: string }): Promise<{ branches: Branch[] }>;
  unlockBranch(request: { path?: string; name: string }): Promise<{ branches: Branch[] }>;
  listCollaborators(request: {
    path: string;
  }): Promise<{ collaborators: RepositoryCollaborator[] }>;
  addCollaborator(request: {
    path: string;
    userId: number;
    role: "write" | "read";
  }): Promise<RepositoryCollaborator>;
  removeCollaborator(request: { path: string; userId: number }): Promise<void>;
  getSystemStats(): Promise<SystemStats>;
  createReview(request: Partial<ReviewSession>): Promise<ReviewSession>;
  listReviews(request: { limit?: number }): Promise<{ reviews: ReviewSession[] }>;
  reviewDetail(request: { id: string }): Promise<ReviewDetail>;
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
  listRepositories(): Promise<{ repositories: Repository[] }>;
  upsertRepository(request: { path: string; name?: string }): Promise<Repository>;
  removeRepository(request: { path: string }): Promise<void>;
  listOpVaults(): Promise<{ vaults: OpVault[] }>;
  listOpItems(request: { vault: string }): Promise<{ items: OpItem[] }>;
  runWorkflow(request: RunWorkflowRequest): WorkflowRunStream;
}
