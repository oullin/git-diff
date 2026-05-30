export class PullRequestClient {
    transport;
    constructor(transport) {
        this.transport = transport;
    }
    list(request) {
        const parts = [];
        if (request.path) parts.push(`path=${encodeURIComponent(request.path)}`);
        if (request.limit) parts.push(`limit=${request.limit}`);
        const query = parts.length === 0 ? "" : `?${parts.join("&")}`;
        return this.transport.request("GET", `/v1/repository/pull-requests${query}`);
    }
    read(request) {
        const parts = [`number=${request.number}`];
        if (request.path) parts.push(`path=${encodeURIComponent(request.path)}`);
        return this.transport.request("GET", `/v1/repository/pull-request?${parts.join("&")}`);
    }
}
