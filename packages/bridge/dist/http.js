import { request as httpRequest } from "node:http";
const DEFAULT_REQUEST_TIMEOUT_MS = 30000;
export function isBridgeError(value) {
    return value instanceof Error && typeof value.kind === "string";
}
export class SocketHttpTransport {
    socketPath;
    constructor(socketPath) {
        this.socketPath = socketPath;
    }
    request(method, path, body, timeoutMs = DEFAULT_REQUEST_TIMEOUT_MS) {
        return requestJson(this.socketPath, method, path, body, timeoutMs);
    }
    requestBytes(path, timeoutMs = DEFAULT_REQUEST_TIMEOUT_MS) {
        return requestBytes(this.socketPath, path, timeoutMs);
    }
}
export function requestJson(
    socketPath,
    method,
    path,
    body,
    timeoutMs = DEFAULT_REQUEST_TIMEOUT_MS,
) {
    return new Promise((resolve, reject) => {
        let settled = false;
        const headers = { Accept: "application/json" };
        const payload = body === undefined ? null : JSON.stringify(body);
        if (payload !== null) {
            headers["Content-Type"] = "application/json";
            headers["Content-Length"] = Buffer.byteLength(payload);
        }
        const fail = (error) => {
            if (settled) {
                return;
            }
            settled = true;
            reject(error);
        };
        const succeed = (value) => {
            if (settled) {
                return;
            }
            settled = true;
            resolve(value);
        };
        const req = httpRequest({ socketPath, method, path, headers }, (res) => {
            consumeBody(res)
                .then((raw) => {
                    if ((res.statusCode === 200 || res.statusCode === 201) && isJsonResponse(res)) {
                        try {
                            succeed(JSON.parse(raw));
                        } catch (error) {
                            fail(
                                transportError(`${method} ${path} returned malformed JSON`, error),
                            );
                        }
                        return;
                    }
                    if (
                        res.statusCode === 200 ||
                        res.statusCode === 201 ||
                        res.statusCode === 204
                    ) {
                        succeed(undefined);
                        return;
                    }
                    fail(buildHttpError(method, path, res, raw));
                })
                .catch((error) =>
                    fail(transportError(`${method} ${path} response read failed`, error)),
                );
        });
        req.setTimeout(timeoutMs, () => {
            req.destroy(transportError(`${method} ${path} timed out after ${timeoutMs}ms`));
        });
        req.on("error", (error) =>
            fail(transportError(`${method} ${path} transport error`, error)),
        );
        if (payload !== null) {
            req.write(payload);
        }
        req.end();
    });
}
export function requestBytes(socketPath, path, timeoutMs = DEFAULT_REQUEST_TIMEOUT_MS) {
    return new Promise((resolve, reject) => {
        let settled = false;
        const fail = (error) => {
            if (settled) {
                return;
            }
            settled = true;
            reject(error);
        };
        const succeed = (value) => {
            if (settled) {
                return;
            }
            settled = true;
            resolve(value);
        };
        const req = httpRequest(
            { socketPath, method: "GET", path, headers: { Accept: "*/*" } },
            (res) => {
                consumeBytes(res)
                    .then((data) => {
                        const status = res.statusCode ?? 0;
                        if (status === 200 || status === 201) {
                            succeed({
                                data,
                                mime: res.headers["content-type"] ?? "application/octet-stream",
                            });
                            return;
                        }
                        // Error bodies come back as JSON via writeError on the
                        // Go side; reuse the JSON error path by decoding the
                        // body as UTF-8.
                        const raw = Buffer.from(data).toString("utf8");
                        fail(buildHttpError("GET", path, res, raw));
                    })
                    .catch((error) =>
                        fail(transportError(`GET ${path} response read failed`, error)),
                    );
            },
        );
        req.setTimeout(timeoutMs, () => {
            req.destroy(transportError(`GET ${path} timed out after ${timeoutMs}ms`));
        });
        req.on("error", (error) => fail(transportError(`GET ${path} transport error`, error)));
        req.end();
    });
}
export function consumeBytes(res) {
    return new Promise((resolve, reject) => {
        const chunks = [];
        res.on("data", (chunk) => {
            chunks.push(chunk);
        });
        res.on("end", () => resolve(new Uint8Array(Buffer.concat(chunks))));
        res.on("error", reject);
    });
}
export function consumeBody(res) {
    return new Promise((resolve, reject) => {
        res.setEncoding("utf8");
        let body = "";
        res.on("data", (chunk) => {
            body += chunk;
        });
        res.on("end", () => resolve(body));
        res.on("error", reject);
    });
}
function isJsonResponse(res) {
    return (res.headers["content-type"] ?? "").includes("application/json");
}
function buildHttpError(method, path, res, raw) {
    const status = res.statusCode ?? 0;
    const kind = kindForStatus(status);
    const error = new Error(`${method} ${path} failed (${status}): ${raw}`);
    error.kind = kind;
    error.statusCode = status;
    if (isJsonResponse(res)) {
        applyJsonErrorPayload(error, raw);
    }
    return error;
}
function kindForStatus(status) {
    if (status === 401) return "auth";
    if (status === 400) return "validation";
    if (status === 404) return "notFound";
    if (status === 409) return "conflict";
    return "server";
}
function transportError(message, cause) {
    const error = new Error(message);
    error.kind = "transport";
    if (cause instanceof Error) {
        error.message = `${message}: ${cause.message}`;
    }
    return error;
}
function applyJsonErrorPayload(error, raw) {
    try {
        const parsed = JSON.parse(raw);
        if (parsed && typeof parsed === "object") {
            if (typeof parsed.error === "string") {
                error.message = parsed.error;
            }
            if (typeof parsed.code === "string") {
                error.code = parsed.code;
            }
            if (
                Array.isArray(parsed.files) &&
                parsed.files.every((entry) => typeof entry === "string")
            ) {
                error.files = parsed.files;
            }
        }
    } catch {
        // Generic error message remains in place.
    }
}
