import type { RuntimeSettings } from "@git-diff/contracts";
import { type ApiClient, createApiClient, waitForReady } from "@git-diff/bridge";
import {
    type BridgeProcessHandle,
    externalBridgeSocketPath,
    killBridge,
    spawnBridge,
} from "#electron/bridge-process.js";
import { recordDiagnostic } from "#electron/diagnostics.js";

let bridgeClient: ApiClient | null = null;
let bridgeHandle: BridgeProcessHandle | null = null;
let bridgeStartup: Promise<void> | null = null;
let savedSettings: Partial<RuntimeSettings> = {};

export function hasExternalBridge(): boolean {
    return externalBridgeSocketPath() !== "";
}

export function getBridgeSettings(): Partial<RuntimeSettings> {
    return savedSettings;
}

export function setBridgeSettings(settings: Partial<RuntimeSettings>): void {
    savedSettings = settings;
}

async function startWorkflowBridge(): Promise<void> {
    if (bridgeClient) {
        return;
    }

    const handle = spawnBridge(savedSettings);

    bridgeHandle = handle;

    handle.process?.on("exit", (code, signal) => {
        if (bridgeClient) {
            const status = code ?? signal ?? "unknown status";

            recordDiagnostic({
                level: "error",
                source: "Backend bridge",
                message: `api HTTP bridge exited with ${status}`,
                details: handle.captureStderr().trim() || undefined,
            });
            console.error(`api HTTP bridge exited with ${status}`);
        }

        bridgeClient?.close();
        bridgeClient = null;
        bridgeHandle = null;
        bridgeStartup = null;
    });

    const httpClient = createApiClient(handle.socketPath);

    try {
        await waitForReady(httpClient);
        bridgeClient = httpClient;
    } catch (error) {
        httpClient.close();
        handle.process?.kill();
        const stderr = handle.captureStderr();

        recordDiagnostic({
            level: "error",
            source: "Backend bridge",
            message: "Failed to start api HTTP bridge",
            details: stderr || (error instanceof Error ? error.stack : String(error)),
        });
        throw new Error(stderr || (error instanceof Error ? error.message : String(error)));
    }
}

export function stopWorkflowBridge(): void {
    bridgeClient?.close();
    bridgeClient = null;
    bridgeStartup = null;

    killBridge(bridgeHandle);
    bridgeHandle = null;
}

export function startBridgeIfNeeded(): Promise<void> {
    if (bridgeClient) {
        return Promise.resolve();
    }

    if (!bridgeStartup) {
        bridgeStartup = startWorkflowBridge().catch((error: unknown) => {
            bridgeStartup = null;
            throw error;
        });
    }

    return bridgeStartup;
}

export async function client(): Promise<ApiClient> {
    if (!bridgeClient) {
        await startBridgeIfNeeded();
    }

    if (!bridgeClient) {
        throw new Error("api HTTP bridge is not running");
    }

    return bridgeClient;
}

// Re-export settingsArgs for any callers that built CLI invocations directly.
export { settingsArgs } from "#electron/bridge-process.js";
