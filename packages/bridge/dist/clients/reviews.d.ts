import type { ReviewComment, ReviewDetail, ReviewEvent, ReviewSession } from "@git-diff/contracts";
import type { HttpTransport } from "#bridge/http.js";
export declare class ReviewClient {
    private readonly transport;
    constructor(transport: HttpTransport);
    create(request: Partial<ReviewSession>): Promise<ReviewSession>;
    list(request?: { limit?: number }): Promise<{
        reviews: ReviewSession[];
    }>;
    detail(request: { id: string }): Promise<ReviewDetail>;
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
}
