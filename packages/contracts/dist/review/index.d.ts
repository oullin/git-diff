import type { RepositoryMode } from "../repo/index.js";
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
export interface PendingComment {
    id: string;
    userId: number;
    repoRoot: string;
    contextKind: RepositoryMode;
    contextSha?: string;
    filePath: string;
    diffSection: string;
    side: string;
    lineNumber: number;
    authorLabel: string;
    bodyHtml: string;
    createdAt: string;
    updatedAt: string;
}
//# sourceMappingURL=index.d.ts.map
