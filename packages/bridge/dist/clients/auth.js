export class AuthClient {
    transport;
    constructor(transport) {
        this.transport = transport;
    }
    getState() {
        return this.transport.request("GET", "/v1/auth/state");
    }
    setup(request) {
        return this.transport.request("POST", "/v1/auth/setup", {
            password: request.password,
        });
    }
    login(request) {
        return this.transport.request("POST", "/v1/auth/login", {
            password: request.password,
            remember: request.remember,
        });
    }
    resume(request) {
        return this.transport.request("POST", "/v1/auth/resume", {
            token: request.token,
        });
    }
    logout() {
        return this.transport.request("POST", "/v1/auth/logout");
    }
    wipe(request) {
        return this.transport.request("POST", "/v1/auth/wipe", {
            osUsername: request.osUsername ?? "",
        });
    }
}
