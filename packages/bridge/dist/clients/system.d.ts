import type { SystemStats } from "@git-diff/contracts";
import type { HttpTransport } from "#bridge/http.js";
export declare class SystemClient {
    private readonly transport;
    constructor(transport: HttpTransport);
    healthz(): Promise<void>;
    stats(): Promise<SystemStats>;
}
