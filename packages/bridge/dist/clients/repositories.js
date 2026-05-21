export class RepositoriesClient {
    transport;
    constructor(transport) {
        this.transport = transport;
    }
    list() {
        return this.transport.request("GET", "/v1/repositories");
    }
    searchFiles(request) {
        const params = new URLSearchParams({ q: request.query });
        if (typeof request.limit === "number") {
            params.set("limit", String(request.limit));
        }
        return this.transport.request("GET", `/v1/repositories/search-files?${params.toString()}`);
    }
    upsert(request) {
        return this.transport.request("POST", "/v1/repositories", {
            path: request.path,
            name: request.name ?? "",
        });
    }
    remove(request) {
        return this.transport.request(
            "DELETE",
            `/v1/repositories?path=${encodeURIComponent(request.path)}`,
        );
    }
    listCollaborators(request) {
        return this.transport.request(
            "GET",
            `/v1/repositories/collaborators?path=${encodeURIComponent(request.path)}`,
        );
    }
    addCollaborator(request) {
        return this.transport.request("POST", "/v1/repositories/collaborators", {
            path: request.path,
            userId: request.userId,
            role: request.role,
        });
    }
    removeCollaborator(request) {
        const params = new URLSearchParams({
            path: request.path,
            userId: String(request.userId),
        });
        return this.transport.request(
            "DELETE",
            `/v1/repositories/collaborators?${params.toString()}`,
        );
    }
}
