import type { BundledLanguage } from "shiki";

import HighlightWorker from "../workers/highlight.worker.ts?worker";
import type { WorkerResponse } from "../workers/highlight.worker";

export type ThemeId = "Licht" | "Dunkel";

// Past this length the per-line tokenizer can stall the worker for hundreds of
// ms on pathological inputs (minified bundles, generated JSON). Above the cap
// we serve plaintext.
export const MAX_LINE_TOKENIZE_LENGTH = 20000;

type PendingHighlight = {
    resolve: (html: string) => void;
    reject: (error: Error) => void;
};

type WorkerSlot = {
    worker: Worker;
    pending: Map<number, PendingHighlight>;
};

let nextRequestId = 1;
let slots: WorkerSlot[] | null = null;
let nextSlot = 0;
let poolDisabled = false;

function workerCount(): number {
    const cores = typeof navigator !== "undefined" ? navigator.hardwareConcurrency : 0;

    return Math.max(1, Math.min(4, (cores || 2) - 1));
}

function createSlot(): WorkerSlot | null {
    if (typeof Worker === "undefined") {
        return null;
    }

    try {
        const worker = new HighlightWorker();
        const pending = new Map<number, PendingHighlight>();

        worker.addEventListener("message", (event: MessageEvent<WorkerResponse>) => {
            const msg = event.data;
            const entry = pending.get(msg.id);

            if (!entry) {
                return;
            }

            pending.delete(msg.id);

            if (msg.kind === "highlight") {
                entry.resolve(msg.html);

                return;
            }

            if (msg.kind === "prewarm") {
                entry.resolve("");

                return;
            }

            entry.reject(new Error(msg.message));
        });

        worker.addEventListener("error", (event) => {
            for (const entry of pending.values()) {
                entry.reject(new Error(event.message || "highlight worker error"));
            }

            pending.clear();
        });

        return { worker, pending };
    } catch {
        return null;
    }
}

function ensurePool(): WorkerSlot[] | null {
    if (poolDisabled) {
        return null;
    }

    if (slots) {
        return slots;
    }

    const desired = workerCount();
    const created: WorkerSlot[] = [];

    for (let i = 0; i < desired; i += 1) {
        const slot = createSlot();

        if (slot) {
            created.push(slot);
        }
    }

    if (!created.length) {
        poolDisabled = true;

        return null;
    }

    slots = created;

    return slots;
}

function nextRequestSlot(pool: WorkerSlot[]): WorkerSlot {
    const slot = pool[nextSlot % pool.length];

    nextSlot = (nextSlot + 1) % pool.length;

    return slot;
}

export function isHighlightPoolAvailable(): boolean {
    return ensurePool() !== null;
}

export function requestHighlight(
    text: string,
    lang: BundledLanguage,
    theme: ThemeId,
): Promise<string> {
    const pool = ensurePool();

    if (!pool) {
        return Promise.reject(new Error("highlight pool unavailable"));
    }

    const slot = nextRequestSlot(pool);
    const id = nextRequestId++;

    return new Promise<string>((resolve, reject) => {
        slot.pending.set(id, { resolve, reject });
        slot.worker.postMessage({ kind: "highlight", id, text, lang, theme });
    });
}

export function prewarmLanguage(lang: BundledLanguage): Promise<void> {
    const pool = ensurePool();

    if (!pool) {
        return Promise.resolve();
    }

    // Prewarm every slot so the next per-line request hits a worker that already
    // knows the grammar. Independent failures shouldn't abort the others.
    return Promise.all(
        pool.map(
            (slot) =>
                new Promise<void>((resolve, reject) => {
                    const id = nextRequestId++;

                    slot.pending.set(id, {
                        resolve: () => resolve(),
                        reject,
                    });
                    slot.worker.postMessage({ kind: "prewarm", id, lang });
                }),
        ),
    ).then(() => undefined);
}
