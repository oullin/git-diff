import type { BundledLanguage } from "shiki";

/**
 * Message protocol for the highlight worker. Both the worker and the
 * highlightPool import these types so a change to the wire shape breaks
 * the compile on both sides at once.
 */

export type HighlightThemeId = "Licht" | "Dunkel";

export interface HighlightRequest {
    kind: "highlight";
    id: number;
    text: string;
    lang: BundledLanguage;
    theme: HighlightThemeId;
}

export interface PrewarmRequest {
    kind: "prewarm";
    id: number;
    lang: BundledLanguage;
}

export type WorkerRequest = HighlightRequest | PrewarmRequest;

export interface HighlightResponse {
    kind: "highlight";
    id: number;
    html: string;
}

export interface PrewarmResponse {
    kind: "prewarm";
    id: number;
}

export interface ErrorResponse {
    kind: "error";
    id: number;
    message: string;
}

export type WorkerResponse = HighlightResponse | PrewarmResponse | ErrorResponse;
