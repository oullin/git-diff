import type { RepositoryMode } from "../repo/index.js";

export interface ReviewSession {
    id: number;
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
    createdAt: string;
    updatedAt: string;
}

export interface ReviewEvent {
    id: number;
    reviewId: number;
    type: string;
    filePath?: string;
    message?: string;
    metadata: string;
    createdAt: string;
    updatedAt: string;
}

export interface ReviewComment {
    id: number;
    reviewId: number;
    filePath: string;
    diffSection: string;
    side: string;
    lineNumber: number;
    /** First line of the comment range; absent for single-line comments. */
    startLineNumber?: number;
    /** Side of the range start; absent when the range stays on `side`. */
    startSide?: string;
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

export interface PendingComment {
    id: number;
    userId: number;
    repoRoot: string;
    contextKind: RepositoryMode;
    contextSha?: string;
    filePath: string;
    diffSection: string;
    side: string;
    lineNumber: number;
    /** First line of the comment range; absent for single-line drafts. */
    startLineNumber?: number;
    /** Side of the range start; absent when the range stays on `side`. */
    startSide?: string;
    authorLabel: string;
    bodyHtml: string;
    createdAt: string;
    updatedAt: string;
}
