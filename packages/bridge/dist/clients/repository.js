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
    /**
     * Fetch the raw bytes of a file at a given ref. Empty ref reads the
     * working tree, ":0" reads the index, otherwise treated as a git ref.
     * Returns the bytes plus the Content-Type so the renderer can build
     * an object URL for inline image display.
     */
    readFileBytes(request) {
        const parts = [`path=${encodeURIComponent(request.path)}`];
        if (request.root) {
            parts.push(`root=${encodeURIComponent(request.root)}`);
        }
        if (request.ref !== undefined) {
            parts.push(`ref=${encodeURIComponent(request.ref)}`);
        }
        return this.transport.requestBytes(`/v1/repository/file/raw?${parts.join("&")}`);
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
