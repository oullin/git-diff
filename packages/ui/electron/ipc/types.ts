import type { BrowserWindow } from "electron";

export type IpcDeps = {
  getMainWindow: () => BrowserWindow | null;
  openDevToolsPanel: (window: BrowserWindow) => void;
};
