import type { UserConfig } from "@git-diff/domain";
import type { HttpTransport } from "#bridge/http.js";
/** Hot-reload SSE is separate — subscribe to GET /v1/userconfig/stream
 *  with an EventSource directly. */
export declare class UserConfigClient {
    private readonly transport;
    constructor(transport: HttpTransport);
    get(): Promise<UserConfig>;
}
