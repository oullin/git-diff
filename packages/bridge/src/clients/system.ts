import type { SystemStats } from "@git-diff/domain";
import type { HttpTransport } from "#bridge/http.js";

export class SystemClient {
    constructor(private readonly transport: HttpTransport) {}

    healthz(): Promise<void> {
        return this.transport.request<void>("GET", "/v1/healthz");
    }

    stats(): Promise<SystemStats> {
        return this.transport.request<SystemStats>("GET", "/v1/system/stats");
    }
}
