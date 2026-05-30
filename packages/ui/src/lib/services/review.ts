import type {
    PendingComment,
    ReviewComment,
    ReviewDetail,
    ReviewEvent,
    ReviewSession,
    RepositoryMode,
} from "@git-diff/contracts";

export interface ReviewService {
    create(request: Partial<ReviewSession>): Promise<ReviewSession>;
    list(limit?: number): Promise<{ reviews: ReviewSession[] }>;
    detail(id: number): Promise<ReviewDetail>;
    addEvent(request: {
        reviewId: number;
        type: string;
        filePath?: string;
        message?: string;
        metadata?: string;
    }): Promise<ReviewEvent>;
    createComment(request: {
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
    updateComment(request: {
        reviewId: number;
        commentId: number;
        bodyHtml: string;
    }): Promise<ReviewComment>;
    deleteComment(request: { reviewId: number; commentId: number }): Promise<void>;
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
}

export function createReviewService(): ReviewService {
    const api = () => window.diffApp;

    return {
        create: (request) => api().createReview(request),
        list: (limit) => api().listReviews(limit),
        detail: (id) => api().reviewDetail(id),
        addEvent: (request) => api().addReviewEvent(request),
        createComment: (request) => api().createReviewComment(request),
        updateComment: (request) => api().updateReviewComment(request),
        deleteComment: (request) => api().deleteReviewComment(request),
        listPendingComments: (request) => api().listPendingComments(request),
        createPendingComment: (request) => api().createPendingComment(request),
        updatePendingComment: (request) => api().updatePendingComment(request),
        deletePendingComment: (id) => api().deletePendingComment(id),
        promotePendingComments: (reviewId) => api().promotePendingComments(reviewId),
    };
}
