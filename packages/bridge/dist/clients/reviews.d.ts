import type { ReviewComment, ReviewDetail, ReviewEvent, ReviewSession } from "@git-diff/contracts";
import type { HttpTransport } from "#bridge/http.js";
export declare class ReviewClient {
    private readonly transport;
    constructor(transport: HttpTransport);
    create(request: Partial<ReviewSession>): Promise<ReviewSession>;
    list(request?: { limit?: number }): Promise<{
        reviews: ReviewSession[];
    }>;
    detail(request: { id: number }): Promise<ReviewDetail>;
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
    setCommentResolved(request: {
        reviewId: number;
        commentId: number;
        resolved: boolean;
    }): Promise<ReviewComment>;
}
