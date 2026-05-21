import type { RuntimeSettings } from "@git-diff/contracts";
import { app } from "electron";
import { spawn, type ChildProcess } from "node:child_process";
import { rmSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { apiDir } from "#electron/paths.js";

/**
 * Information about a spawned (or attached-to) Go API process.
 */
export interface BridgeProcessHandle {
    process: ChildProcess | null;
    socketPath: string;
    /** True when we attached to an externally-managed socket via API_BRIDGE_SOCKET. */
    external: boolean;
    /** Accumulated stderr from the child; empty when external. */
    captureStderr(): string;
}

export function externalBridgeSocketPath(): string {
    return process.env.API_BRIDGE_SOCKET?.trim() ?? "";
}

/**
 * spawnBridge starts the Go API binary and returns a handle wrapping the
 * child process plus its socket path. If API_BRIDGE_SOCKET is set, no
 * process is spawned — the handle just points at the existing socket.
 */
export function spawnBridge(savedSettings: Partial<RuntimeSettings>): BridgeProcessHandle {
    const externalPath = externalBridgeSocketPath();

    if (externalPath) {
        return {
            process: null,
            socketPath: externalPath,
            external: true,
            captureStderr: () => "",
        };
    }

    const command = goCommand();
    const socketPath = join(tmpdir(), `api-${process.pid}-${Date.now()}.sock`);
    const child = spawn(
        command.command,
        [...command.args, "serve-http", "--socket", socketPath, ...settingsArgs(savedSettings)],
        {
            cwd: app.isPackaged ? app.getPath("userData") : apiDir,
            env: process.env,
            stdio: ["ignore", "pipe", "pipe"],
        },
    );

    let stderr = "";

    child.stderr?.on("data", (chunk: Buffer) => {
        stderr += chunk.toString("utf8");
    });

    return {
        process: child,
        socketPath,
        external: false,
        captureStderr: () => stderr,
    };
}

export function killBridge(handle: BridgeProcessHandle | null): void {
    if (!handle) {
        return;
    }

    handle.process?.kill();

    if (!handle.external && handle.socketPath) {
        rmSync(handle.socketPath, { force: true });
    }
}

function goCommand() {
    const packaged = join(process.resourcesPath || "", "api");

    if (app.isPackaged) {
        return { command: packaged, args: [] };
    }

    return { command: "go", args: ["run", "./cmd"] };
}

export function settingsArgs(settings: Partial<RuntimeSettings>): string[] {
    const args: string[] = [];
    const pairs: Array<[string, string | undefined]> = [
        ["--repo-root", settings.repoRoot],
        ["--apps-config", settings.appsConfigPath],
        ["--secrets-config", settings.secretsConfigPath],
        ["--generated-apps", settings.generatedAppsPath],
        ["--archive-root", settings.archiveRoot],
        ["--workflow-db", settings.workflowDbPath],
        ["--op-vault", settings.opVault],
        ["--op-item", settings.opItem],
    ];

    for (const [flag, value] of pairs) {
        if (value?.trim()) {
            args.push(flag, value);
        }
    }

    return args;
}
