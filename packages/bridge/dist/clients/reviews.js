export class ReviewClient {
    transport;
    constructor(transport) {
        this.transport = transport;
    }
    create(request) {
        return this.transport.request("POST", "/v1/reviews", request);
    }
    list(request = {}) {
        const query = typeof request.limit === "number" ? `?limit=${request.limit}` : "";
        return this.transport.request("GET", `/v1/reviews${query}`);
    }
    detail(request) {
        return this.transport.request("GET", `/v1/reviews/${encodeURIComponent(request.id)}`);
    }
    addEvent(request) {
        return this.transport.request(
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
    createComment(request) {
        return this.transport.request(
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
    updateComment(request) {
        return this.transport.request(
            "PATCH",
            `/v1/reviews/${encodeURIComponent(request.reviewId)}/comments/${encodeURIComponent(request.commentId)}`,
            { bodyHtml: request.bodyHtml },
        );
    }
    deleteComment(request) {
        return this.transport.request(
            "DELETE",
            `/v1/reviews/${encodeURIComponent(request.reviewId)}/comments/${encodeURIComponent(request.commentId)}`,
        );
    }
}
