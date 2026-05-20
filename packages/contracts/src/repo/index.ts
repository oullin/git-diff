export type GitFileStatus = "added" | "deleted" | "modified" | "renamed" | "untracked";

export type DiffSectionKind = "staged" | "unstaged" | "untracked" | "commit";

export type RepositoryMode = "working" | "commit";

export type RepositoryRole = "owner" | "write" | "read";

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

export interface RepositoryFile {
  path: string;
  content: string;
  binary: boolean;
  truncated: boolean;
  size: number;
}

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
