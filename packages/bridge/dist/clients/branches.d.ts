import type { Branch, RepositoryState } from "@git-diff/contracts";
export declare function listBranches(
    socketPath: string,
    request: {
        path?: string;
    },
): Promise<{
    branches: string[];
    records?: Branch[];
}>;
export declare function checkoutBranch(
    socketPath: string,
    request: {
        path: string;
        branch: string;
    },
): Promise<RepositoryState>;
export declare function createBranch(
    socketPath: string,
    request: {
        path: string;
        name: string;
    },
): Promise<RepositoryState>;
export declare function deleteBranch(
    socketPath: string,
    request: {
        path?: string;
        name: string;
    },
): Promise<void>;
export declare function lockBranch(
    socketPath: string,
    request: {
        path?: string;
        name: string;
    },
): Promise<{
    branches: Branch[];
}>;
export declare function unlockBranch(
    socketPath: string,
    request: {
        path?: string;
        name: string;
    },
): Promise<{
    branches: Branch[];
}>;
