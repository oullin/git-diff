import type { PendingComment } from "@git-diff/contracts";
import type { HttpTransport } from "#bridge/http.js";
export declare class PendingCommentClient {
    private readonly transport;
    constructor(transport: HttpTransport);
    list(request: { path?: string; kind?: "working" | "commit"; sha?: string }): Promise<{
        comments: PendingComment[];
    }>;
    create(request: {
        repoRoot: string;
        contextKind: "working" | "commit";
        contextSha?: string;
        filePath: string;
        diffSection: string;
        side: string;
        lineNumber: number;
        startLineNumber?: number;
        startSide?: string;
        authorLabel: string;
        bodyHtml: string;
    }): Promise<PendingComment>;
    update(request: { id: number; bodyHtml: string }): Promise<PendingComment>;
    delete(request: { id: number }): Promise<void>;
    promote(request: { reviewId: number }): Promise<{
        promoted: number;
    }>;
}
