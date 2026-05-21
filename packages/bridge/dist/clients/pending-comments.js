export class PendingCommentClient {
    transport;
    constructor(transport) {
        this.transport = transport;
    }
    list(request) {
        const parts = [];
        if (request.path) parts.push(`path=${encodeURIComponent(request.path)}`);
        if (request.kind) parts.push(`kind=${request.kind}`);
        if (request.sha) parts.push(`sha=${encodeURIComponent(request.sha)}`);
        const query = parts.length === 0 ? "" : `?${parts.join("&")}`;
        return this.transport.request("GET", `/v1/pending-comments${query}`);
    }
    create(request) {
        return this.transport.request("POST", "/v1/pending-comments", request);
    }
    update(request) {
        return this.transport.request(
            "PATCH",
            `/v1/pending-comments/${encodeURIComponent(request.id)}`,
            { bodyHtml: request.bodyHtml },
        );
    }
    delete(request) {
        return this.transport.request(
            "DELETE",
            `/v1/pending-comments/${encodeURIComponent(request.id)}`,
        );
    }
    promote(request) {
        return this.transport.request("POST", "/v1/pending-comments/promote", request);
    }
}
