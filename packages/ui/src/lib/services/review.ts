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
    detail(id: string): Promise<ReviewDetail>;
    addEvent(request: {
        reviewId: string;
        type: string;
        filePath?: string;
        message?: string;
        metadata?: string;
    }): Promise<ReviewEvent>;
    createComment(request: {
        reviewId: string;
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
        reviewId: string;
        commentId: string;
        bodyHtml: string;
    }): Promise<ReviewComment>;
    deleteComment(request: { reviewId: string; commentId: string }): Promise<void>;
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
    updatePendingComment(request: { id: string; bodyHtml: string }): Promise<PendingComment>;
    deletePendingComment(id: string): Promise<void>;
    promotePendingComments(reviewId: string): Promise<{ promoted: number }>;
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
