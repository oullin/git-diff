import type { WalkthroughRecord } from "@git-diff/domain";
import type { HttpTransport } from "#bridge/http.js";

export class WalkthroughClient {
    constructor(private readonly transport: HttpTransport) {}

    generate(request: {
        path?: string;
        kind?: "working" | "commit";
        sha?: string;
        refresh?: boolean;
    }): Promise<WalkthroughRecord> {
        return this.transport.request<WalkthroughRecord>("POST", "/v1/walkthrough", {
            path: request.path ?? "",
            kind: request.kind ?? "working",
            sha: request.sha ?? "",
            refresh: request.refresh ?? false,
        });
    }
}
