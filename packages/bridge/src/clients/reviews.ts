import type { ReviewComment, ReviewDetail, ReviewEvent, ReviewSession } from "@git-diff/contracts";
import type { HttpTransport } from "#bridge/http.js";

export class ReviewClient {
    constructor(private readonly transport: HttpTransport) {}

    create(request: Partial<ReviewSession>): Promise<ReviewSession> {
        return this.transport.request<ReviewSession>(
            "POST",
            "/v1/reviews",
            request as Record<string, unknown>,
        );
    }

    list(request: { limit?: number } = {}): Promise<{ reviews: ReviewSession[] }> {
        const query = typeof request.limit === "number" ? `?limit=${request.limit}` : "";

        return this.transport.request<{ reviews: ReviewSession[] }>("GET", `/v1/reviews${query}`);
    }

    detail(request: { id: string }): Promise<ReviewDetail> {
        return this.transport.request<ReviewDetail>(
            "GET",
            `/v1/reviews/${encodeURIComponent(request.id)}`,
        );
    }

    addEvent(request: {
        reviewId: string;
        type: string;
        filePath?: string;
        message?: string;
        metadata?: string;
    }): Promise<ReviewEvent> {
        return this.transport.request<ReviewEvent>(
            "POST",
            `/v1/reviews/${encodeURIComponent(request.reviewId)}/events`,
            {
                type: request.type,
                filePath: request.filePath,
                message: request.message,
                metadata: request.metadata,
            },
        );
    }

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
    }): Promise<ReviewComment> {
        return this.transport.request<ReviewComment>(
            "POST",
            `/v1/reviews/${encodeURIComponent(request.reviewId)}/comments`,
            {
                filePath: request.filePath,
                diffSection: request.diffSection,
                side: request.side,
                lineNumber: request.lineNumber,
                startLineNumber: request.startLineNumber,
                startSide: request.startSide,
                authorLabel: request.authorLabel,
                bodyHtml: request.bodyHtml,
            },
        );
    }

    updateComment(request: {
        reviewId: string;
        commentId: string;
        bodyHtml: string;
    }): Promise<ReviewComment> {
        return this.transport.request<ReviewComment>(
            "PATCH",
            `/v1/reviews/${encodeURIComponent(request.reviewId)}/comments/${encodeURIComponent(request.commentId)}`,
            { bodyHtml: request.bodyHtml },
        );
    }

    deleteComment(request: { reviewId: string; commentId: string }): Promise<void> {
        return this.transport.request<void>(
            "DELETE",
            `/v1/reviews/${encodeURIComponent(request.reviewId)}/comments/${encodeURIComponent(request.commentId)}`,
        );
    }
}
