import type { OpItem, OpVault } from "@git-diff/contracts";
export declare function listOpVaults(socketPath: string): Promise<{
    vaults: OpVault[];
}>;
export declare function listOpItems(socketPath: string, request: {
    vault: string;
}): Promise<{
    items: OpItem[];
}>;
