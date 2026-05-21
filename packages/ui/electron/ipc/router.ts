import type { IpcMain, IpcMainInvokeEvent } from "electron";

/**
 * Handler signature for an IPC invoke channel. Renderer args follow the
 * event, exactly like `ipcMain.handle`.
 */
export type IpcHandler<Args extends unknown[] = unknown[], Result = unknown> = (
    event: IpcMainInvokeEvent,
    ...args: Args
) => Result | Promise<Result>;

/**
 * IpcRouter accumulates channel → handler bindings and applies them to
 * `ipcMain` in a single pass. Each domain module registers its channels
 * through the router, which keeps the wiring discoverable (one place to
 * grep for a channel name) and decouples handler registration from the
 * electron runtime — tests can inspect the handler map directly.
 */
export class IpcRouter {
    private readonly handlers = new Map<string, IpcHandler>();

    on<Args extends unknown[], Result>(channel: string, handler: IpcHandler<Args, Result>): this {
        if (this.handlers.has(channel)) {
            throw new Error(`IpcRouter: channel "${channel}" already registered`);
        }

        this.handlers.set(channel, handler as IpcHandler);

        return this;
    }

    register(ipcMain: IpcMain): void {
        for (const [channel, handler] of this.handlers) {
            ipcMain.handle(channel, handler);
        }
    }

    /** Useful for tests: returns the registered handler for a channel, if any. */
    handlerFor(channel: string): IpcHandler | undefined {
        return this.handlers.get(channel);
    }
}
