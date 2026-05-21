export class BranchClient {
    transport;
    constructor(transport) {
        this.transport = transport;
    }
    list(request) {
        const query = request.path ? `?path=${encodeURIComponent(request.path)}` : "";
        return this.transport.request("GET", `/v1/repository/branches${query}`);
    }
    checkout(request) {
        return this.transport.request("POST", "/v1/repository/checkout", {
            path: request.path,
            branch: request.branch,
        });
    }
    create(request) {
        return this.transport.request("POST", "/v1/repository/branches/create", {
            path: request.path,
            name: request.name,
        });
    }
    delete(request) {
        const params = new URLSearchParams({ name: request.name });
        if (request.path) params.set("path", request.path);
        return this.transport.request("DELETE", `/v1/repository/branches?${params.toString()}`);
    }
    lock(request) {
        return this.transport.request("POST", "/v1/repository/branches/lock", {
            path: request.path,
            name: request.name,
        });
    }
    unlock(request) {
        return this.transport.request("POST", "/v1/repository/branches/unlock", {
            path: request.path,
            name: request.name,
        });
    }
}
