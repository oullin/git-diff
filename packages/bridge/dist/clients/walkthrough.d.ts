import type { WalkthroughRecord } from "@git-diff/domain";
import type { HttpTransport } from "#bridge/http.js";
export declare class WalkthroughClient {
    private readonly transport;
    constructor(transport: HttpTransport);
    generate(request: {
        path?: string;
        kind?: "working" | "commit";
        sha?: string;
        refresh?: boolean;
    }): Promise<WalkthroughRecord>;
}
