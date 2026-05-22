import type { UserConfig } from "@git-diff/contracts";
import type { HttpTransport } from "#bridge/http.js";

/**
 * Reads the user-editable YAML config served by the Go backend.
 *
 * SSE streaming for hot-reload notifications lives separately —
 * subscribe by hitting GET /v1/userconfig/stream with an EventSource
 * directly (the unix-socket transport doesn't expose streaming).
 */
export class UserConfigClient {
    constructor(private readonly transport: HttpTransport) {}

    get(): Promise<UserConfig> {
        return this.transport.request<UserConfig>("GET", "/v1/userconfig");
    }
}
