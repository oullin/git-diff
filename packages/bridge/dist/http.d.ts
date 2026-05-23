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
export declare class SocketHttpTransport implements HttpTransport {
    private readonly socketPath;
    constructor(socketPath: string);
    request<Response>(
        method: HttpMethod,
        path: string,
        body?: JsonBody,
        timeoutMs?: number,
    ): Promise<Response>;
    requestBytes(path: string, timeoutMs?: number): Promise<BytesResponse>;
}
export declare function requestJson<Response>(
    socketPath: string,
    method: HttpMethod,
    path: string,
    body?: JsonBody,
    timeoutMs?: number,
): Promise<Response>;
export declare function requestBytes(
    socketPath: string,
    path: string,
    timeoutMs?: number,
): Promise<BytesResponse>;
export declare function consumeBytes(res: IncomingMessage): Promise<Uint8Array>;
export declare function consumeBody(res: IncomingMessage): Promise<string>;
