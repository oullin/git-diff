/**
 * Reads the user-editable YAML config served by the Go backend.
 *
 * SSE streaming for hot-reload notifications lives separately —
 * subscribe by hitting GET /v1/userconfig/stream with an EventSource
 * directly (the unix-socket transport doesn't expose streaming).
 */
export class UserConfigClient {
    transport;
    constructor(transport) {
        this.transport = transport;
    }
    get() {
        return this.transport.request("GET", "/v1/userconfig");
    }
}
