import type { IpcMain, IpcMainInvokeEvent } from "electron";

export type IpcHandler<Args extends unknown[] = unknown[], Result = unknown> = (
    event: IpcMainInvokeEvent,
    ...args: Args
) => Result | Promise<Result>;

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

    handlerFor(channel: string): IpcHandler | undefined {
        return this.handlers.get(channel);
    }
}
