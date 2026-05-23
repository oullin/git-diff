export type { DiffSectionKind, GitFileStatus, RepositoryMode, RepositoryRole, } from "../common/index.js";
import type { DiffSectionKind, GitFileStatus, RepositoryMode, RepositoryRole } from "../common/index.js";
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
export interface RepositoryFileRange {
    path: string;
    startLine: number;
    endLine: number;
    lines: string[];
    eof: boolean;
}
export interface Repository {
    id: number;
    path: string;
    name: string;
    ownerId: number;
    role: RepositoryRole;
    addedAt: string;
    lastOpenedAt?: string;
    createdAt: string;
    updatedAt: string;
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
/**
 * Lowercase file extensions (with leading dot) the renderer can display
 * via the inline image-diff component. The backend serves the raw bytes
 * regardless of extension — this list is purely a UI routing decision.
 */
export declare const IMAGE_EXTENSIONS: ReadonlyArray<string>;
/** True when path's extension is in IMAGE_EXTENSIONS (case-insensitive). */
export declare function isImagePath(path: string): boolean;
//# sourceMappingURL=index.d.ts.map