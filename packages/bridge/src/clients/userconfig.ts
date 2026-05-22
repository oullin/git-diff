import type { UserConfig } from "@git-diff/contracts";
import type { HttpTransport } from "#bridge/http.js";

/** Hot-reload SSE is separate — subscribe to GET /v1/userconfig/stream
 *  with an EventSource directly. */
export class UserConfigClient {
    constructor(private readonly transport: HttpTransport) {}

    get(): Promise<UserConfig> {
        return this.transport.request<UserConfig>("GET", "/v1/userconfig");
    }
}
