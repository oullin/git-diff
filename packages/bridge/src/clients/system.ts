import type { SystemStats } from "@git-diff/contracts";
import { requestJson } from "#bridge/http.js";

export function healthz(socketPath: string): Promise<void> {
    return requestJson<void>(socketPath, "GET", "/v1/healthz");
}

export function getSystemStats(socketPath: string): Promise<SystemStats> {
    return requestJson<SystemStats>(socketPath, "GET", "/v1/system/stats");
}
