export class SystemClient {
    transport;
    constructor(transport) {
        this.transport = transport;
    }
    healthz() {
        return this.transport.request("GET", "/v1/healthz");
    }
    stats() {
        return this.transport.request("GET", "/v1/system/stats");
    }
}
