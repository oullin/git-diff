import type { SystemStats } from "@git-diff/contracts";
export declare function healthz(socketPath: string): Promise<void>;
export declare function getSystemStats(socketPath: string): Promise<SystemStats>;
