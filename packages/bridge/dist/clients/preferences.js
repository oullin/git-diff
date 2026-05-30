export class PreferenceClient {
    transport;
    constructor(transport) {
        this.transport = transport;
    }
    get() {
        return this.transport.request("GET", "/v1/preferences");
    }
    save(values) {
        return this.transport.request("POST", "/v1/preferences", { values });
    }
}
