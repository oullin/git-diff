import type { PendingComment } from "@git-diff/contracts";
export declare function listPendingComments(
    socketPath: string,
    request: {
        path?: string;
        kind?: "working" | "commit";
        sha?: string;
    },
): Promise<{
    comments: PendingComment[];
}>;
export declare function createPendingComment(
    socketPath: string,
    request: {
        repoRoot: string;
        contextKind: "working" | "commit";
        contextSha?: string;
        filePath: string;
        diffSection: string;
        side: string;
        lineNumber: number;
        authorLabel: string;
        bodyHtml: string;
    },
): Promise<PendingComment>;
export declare function updatePendingComment(
    socketPath: string,
    request: {
        id: string;
        bodyHtml: string;
    },
): Promise<PendingComment>;
export declare function deletePendingComment(
    socketPath: string,
    request: {
        id: string;
    },
): Promise<void>;
export declare function promotePendingComments(
    socketPath: string,
    request: {
        reviewId: string;
    },
): Promise<{
    promoted: number;
}>;
