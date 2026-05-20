import type { WalkthroughRecord } from "@git-diff/contracts";
export declare function generateWalkthrough(socketPath: string, request: {
    path?: string;
    kind?: "working" | "commit";
    sha?: string;
    refresh?: boolean;
}): Promise<WalkthroughRecord>;
