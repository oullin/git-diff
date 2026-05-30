import { type IncomingMessage, request as httpRequest } from "node:http";

const DEFAULT_REQUEST_TIMEOUT_MS = 30000;

export type HttpMethod = "GET" | "POST" | "PUT" | "PATCH" | "DELETE";
export type JsonBody = Record<string, unknown>;

/**
 * Discriminated error union returned by the bridge transport. `kind`
 * partitions the cases so consumers can branch on a single field:
 * - `transport`: socket/connect/timeout failures
 * - `auth`: HTTP 401
 * - `validation`: HTTP 400 (`files` carries the offending paths)
 * - `notFound`: HTTP 404
 * - `conflict`: HTTP 409 (`code` carries the storage error key)
 * - `server`: everything else
 */
export type BridgeErrorKind =
    | "transport"
    | "auth"
    | "validation"
    | "notFound"
    | "conflict"
    | "server";

export interface BridgeError extends Error {
    kind: BridgeErrorKind;
    statusCode?: number;
    code?: string;
    files?: string[];
}

export function isBridgeError(value: unknown): value is BridgeError {
    return value instanceof Error && typeof (value as BridgeError).kind === "string";
}

interface JsonErrorPayload {
    error?: unknown;
    code?: unknown;
    files?: unknown;
}

/** Raw byte response from a binary endpoint (e.g. /v1/repository/file/raw). */
export interface BytesResponse {
    data: Uint8Array;
    mime: string;
}

export interface HttpTransport {
    request<Response>(
        method: HttpMethod,
        path: string,
        body?: JsonBody,
        timeoutMs?: number,
    ): Promise<Response>;

    /**
     * GET path and return the raw response body plus its Content-Type.
     * Errors follow the same BridgeError shape as request(). For non-JSON
     * error responses the message is the raw body.
     */
    requestBytes(path: string, timeoutMs?: number): Promise<BytesResponse>;
}

export class SocketHttpTransport implements HttpTransport {
    constructor(private readonly socketPath: string) {}

    request<Response>(
        method: HttpMethod,
        path: string,
        body?: JsonBody,
        timeoutMs = DEFAULT_REQUEST_TIMEOUT_MS,
    ): Promise<Response> {
        return requestJson<Response>(this.socketPath, method, path, body, timeoutMs);
    }

    requestBytes(path: string, timeoutMs = DEFAULT_REQUEST_TIMEOUT_MS): Promise<BytesResponse> {
        return requestBytes(this.socketPath, path, timeoutMs);
    }
}

export function requestJson<Response>(
    socketPath: string,
    method: HttpMethod,
    path: string,
    body?: JsonBody,
    timeoutMs = DEFAULT_REQUEST_TIMEOUT_MS,
): Promise<Response> {
    return new Promise<Response>((resolve, reject) => {
        let settled = false;
        const headers: Record<string, string | number> = { Accept: "application/json" };
        const payload = body === undefined ? null : JSON.stringify(body);

        if (payload !== null) {
            headers["Content-Type"] = "application/json";
            headers["Content-Length"] = Buffer.byteLength(payload);
        }

        const fail = (error: Error): void => {
            if (settled) {
                return;
            }

            settled = true;
            reject(error);
        };

        const succeed = (value: Response): void => {
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
                            succeed(JSON.parse(raw) as Response);
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
                        succeed(undefined as Response);

                        return;
                    }

                    fail(buildHttpError(method, path, res, raw));
                })
                .catch((error: unknown) =>
                    fail(transportError(`${method} ${path} response read failed`, error)),
                );
        });

        req.setTimeout(timeoutMs, () => {
            req.destroy(transportError(`${method} ${path} timed out after ${timeoutMs}ms`));
        });

        req.on("error", (error: Error) =>
            fail(transportError(`${method} ${path} transport error`, error)),
        );

        if (payload !== null) {
            req.write(payload);
        }

        req.end();
    });
}

export function requestBytes(
    socketPath: string,
    path: string,
    timeoutMs = DEFAULT_REQUEST_TIMEOUT_MS,
): Promise<BytesResponse> {
    return new Promise<BytesResponse>((resolve, reject) => {
        let settled = false;

        const fail = (error: Error): void => {
            if (settled) {
                return;
            }

            settled = true;
            reject(error);
        };

        const succeed = (value: BytesResponse): void => {
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
                                mime: (res.headers["content-type"] ??
                                    "application/octet-stream") as string,
                            });

                            return;
                        }

                        // Error bodies come back as JSON via writeError on the
                        // Go side; reuse the JSON error path by decoding the
                        // body as UTF-8.
                        const raw = Buffer.from(data).toString("utf8");

                        fail(buildHttpError("GET", path, res, raw));
                    })
                    .catch((error: unknown) =>
                        fail(transportError(`GET ${path} response read failed`, error)),
                    );
            },
        );

        req.setTimeout(timeoutMs, () => {
            req.destroy(transportError(`GET ${path} timed out after ${timeoutMs}ms`));
        });

        req.on("error", (error: Error) =>
            fail(transportError(`GET ${path} transport error`, error)),
        );

        req.end();
    });
}

export function consumeBytes(res: IncomingMessage): Promise<Uint8Array> {
    return new Promise<Uint8Array>((resolve, reject) => {
        const chunks: Buffer[] = [];

        res.on("data", (chunk: Buffer) => {
            chunks.push(chunk);
        });

        res.on("end", () => resolve(new Uint8Array(Buffer.concat(chunks))));
        res.on("error", reject);
    });
}

export function consumeBody(res: IncomingMessage): Promise<string> {
    return new Promise<string>((resolve, reject) => {
        res.setEncoding("utf8");
        let body = "";

        res.on("data", (chunk: string) => {
            body += chunk;
        });

        res.on("end", () => resolve(body));
        res.on("error", reject);
    });
}

function isJsonResponse(res: IncomingMessage): boolean {
    return (res.headers["content-type"] ?? "").includes("application/json");
}

function buildHttpError(
    method: HttpMethod,
    path: string,
    res: IncomingMessage,
    raw: string,
): BridgeError {
    const status = res.statusCode ?? 0;
    const kind = kindForStatus(status);
    const error = new Error(`${method} ${path} failed (${status}): ${raw}`) as BridgeError;

    error.kind = kind;
    error.statusCode = status;

    if (isJsonResponse(res)) {
        applyJsonErrorPayload(error, raw);
    }

    return error;
}

function kindForStatus(status: number): BridgeErrorKind {
    if (status === 401) return "auth";

    if (status === 400) return "validation";

    if (status === 404) return "notFound";

    if (status === 409) return "conflict";

    return "server";
}

function transportError(message: string, cause?: unknown): BridgeError {
    const error = new Error(message) as BridgeError;

    error.kind = "transport";

    if (cause instanceof Error) {
        error.message = `${message}: ${cause.message}`;
    }

    return error;
}

function applyJsonErrorPayload(error: BridgeError, raw: string): void {
    try {
        const parsed = JSON.parse(raw) as JsonErrorPayload | null;

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
                error.files = parsed.files as string[];
            }
        }
    } catch {
        // Generic error message remains in place.
    }
}
