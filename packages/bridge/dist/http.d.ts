import { type IncomingMessage } from "node:http";
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
export declare function isBridgeError(value: unknown): value is BridgeError;
/**
 * HttpTransport is the seam between bridge clients and the unix-socket
 * transport. Per-domain clients depend on this interface so tests can
 * substitute an in-memory transport, and the implementation can grow
 * features (retries, tracing) without leaking into the client surface.
 */
export interface HttpTransport {
    request<Response>(
        method: HttpMethod,
        path: string,
        body?: JsonBody,
        timeoutMs?: number,
    ): Promise<Response>;
}
export declare class SocketHttpTransport implements HttpTransport {
    private readonly socketPath;
    constructor(socketPath: string);
    request<Response>(
        method: HttpMethod,
        path: string,
        body?: JsonBody,
        timeoutMs?: number,
    ): Promise<Response>;
}
export declare function requestJson<Response>(
    socketPath: string,
    method: HttpMethod,
    path: string,
    body?: JsonBody,
    timeoutMs?: number,
): Promise<Response>;
export declare function consumeBody(res: IncomingMessage): Promise<string>;
