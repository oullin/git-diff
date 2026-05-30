import type { SystemStats } from "@git-diff/domain";
import type { HttpTransport } from "#bridge/http.js";
export declare class SystemClient {
    private readonly transport;
    constructor(transport: HttpTransport);
    healthz(): Promise<void>;
    stats(): Promise<SystemStats>;
}
