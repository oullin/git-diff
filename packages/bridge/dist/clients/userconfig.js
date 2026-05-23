/** Hot-reload SSE is separate — subscribe to GET /v1/userconfig/stream
 *  with an EventSource directly. */
export class UserConfigClient {
    transport;
    constructor(transport) {
        this.transport = transport;
    }
    get() {
        return this.transport.request("GET", "/v1/userconfig");
    }
}
