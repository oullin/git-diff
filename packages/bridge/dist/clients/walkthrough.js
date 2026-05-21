export class WalkthroughClient {
    transport;
    constructor(transport) {
        this.transport = transport;
    }
    generate(request) {
        return this.transport.request("POST", "/v1/walkthrough", {
            path: request.path ?? "",
            kind: request.kind ?? "working",
            sha: request.sha ?? "",
            refresh: request.refresh ?? false,
        });
    }
}
