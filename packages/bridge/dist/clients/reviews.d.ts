import type { ReviewComment, ReviewDetail, ReviewEvent, ReviewSession } from "@git-diff/contracts";
export declare function createReview(
    socketPath: string,
    request: Partial<ReviewSession>,
): Promise<ReviewSession>;
export declare function listReviews(
    socketPath: string,
    request?: {
        limit?: number;
    },
): Promise<{
    reviews: ReviewSession[];
}>;
export declare function reviewDetail(
    socketPath: string,
    request: {
        id: string;
    },
): Promise<ReviewDetail>;
export declare function addReviewEvent(
    socketPath: string,
    request: {
        reviewId: string;
        type: string;
        filePath?: string;
        message?: string;
        metadata?: string;
    },
): Promise<ReviewEvent>;
export declare function createReviewComment(
    socketPath: string,
    request: {
        reviewId: string;
        filePath: string;
        diffSection: string;
        side: string;
        lineNumber: number;
        authorLabel: string;
        bodyHtml: string;
    },
): Promise<ReviewComment>;
export declare function updateReviewComment(
    socketPath: string,
    request: {
        reviewId: string;
        commentId: string;
        bodyHtml: string;
    },
): Promise<ReviewComment>;
export declare function deleteReviewComment(
    socketPath: string,
    request: {
        reviewId: string;
        commentId: string;
    },
): Promise<void>;
