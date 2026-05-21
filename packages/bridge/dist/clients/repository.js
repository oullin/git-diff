export class RepositoryClient {
    transport;
    constructor(transport) {
        this.transport = transport;
    }
    state(request) {
        const query = request.path ? `?path=${encodeURIComponent(request.path)}` : "";
        return this.transport.request("GET", `/v1/repository/state${query}`);
    }
    open(request) {
        return this.transport.request("POST", "/v1/repository/open", {
            path: request.path,
        });
    }
    refresh(request) {
        return this.transport.request("POST", "/v1/repository/refresh", {
            path: request.path,
        });
    }
    readCommit(request) {
        const parts = [`sha=${encodeURIComponent(request.sha)}`];
        if (request.path) {
            parts.push(`path=${encodeURIComponent(request.path)}`);
        }
        return this.transport.request("GET", `/v1/repository/commit?${parts.join("&")}`);
    }
    listCommits(request) {
        const parts = [];
        if (request.path) {
            parts.push(`path=${encodeURIComponent(request.path)}`);
        }
        if (request.limit) {
            parts.push(`limit=${request.limit}`);
        }
        const query = parts.length === 0 ? "" : `?${parts.join("&")}`;
        return this.transport.request("GET", `/v1/repository/log${query}`);
    }
    readFile(request) {
        const query = `?root=${encodeURIComponent(request.root)}&path=${encodeURIComponent(request.path)}`;
        return this.transport.request("GET", `/v1/repository/file${query}`);
    }
    readFileRange(request) {
        const parts = [
            `root=${encodeURIComponent(request.root)}`,
            `path=${encodeURIComponent(request.path)}`,
            `startLine=${request.startLine}`,
            `endLine=${request.endLine}`,
        ];
        if (request.ref) {
            parts.push(`ref=${encodeURIComponent(request.ref)}`);
        }
        return this.transport.request("GET", `/v1/repository/file-range?${parts.join("&")}`);
    }
}
