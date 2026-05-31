import { app, dialog, Menu, type MenuItemConstructorOptions, shell } from "electron";
import { client } from "#electron/bridge.js";
import { recordDiagnostic } from "#electron/diagnostics.js";
import { installTerminalHelper } from "#electron/terminal-helper.js";
import { createWindow, focusMainWindow, getMainWindow } from "#electron/windows.js";

async function openUserConfigInEditor(): Promise<void> {
  try {
    const cfg = await (await client()).userConfig.get();

    if (!cfg.path) {
      throw new Error("user config path is unavailable");
    }

    const err = await shell.openPath(cfg.path);

    if (err) {
      throw new Error(err);
    }
  } catch (error) {
    recordDiagnostic({
      level: "error",
      source: "User Config",
      message: "Failed to open user config in editor",
      details: error instanceof Error ? error.stack || error.message : String(error),
    });
  }
}

function newWindowItem(): MenuItemConstructorOptions {
  return {
    label: "New Window…",
    accelerator: "CmdOrCtrl+N",
    async click() {
      const parent = getMainWindow() ?? undefined;

      const result = await dialog.showOpenDialog(parent!, {
        title: "Open Repository",
        properties: ["openDirectory"],
      });

      if (result.canceled || result.filePaths.length === 0) {
        return;
      }

      createWindow({
        kind: "working",
        repoPath: result.filePaths[0]!,
        walkthrough: false,
      });
    },
  };
}

export function installApplicationMenu(): void {
  const isMac = process.platform === "darwin";

  const template: MenuItemConstructorOptions[] = [
    ...(isMac
      ? ([
          {
            label: app.name,
            submenu: [
              { role: "about" },
              { type: "separator" },
              {
                label: "Install Terminal Helper…",
                async click() {
                  await installTerminalHelper();
                },
              },
              {
                label: "Open Config in Editor…",
                accelerator: "CmdOrCtrl+,",
                click: () => {
                  void openUserConfigInEditor();
                },
              },
              { type: "separator" },
              { role: "services" },
              { type: "separator" },
              { role: "hide" },
              { role: "hideOthers" },
              { role: "unhide" },
              { type: "separator" },
              { role: "quit" },
            ],
          },
        ] satisfies MenuItemConstructorOptions[])
      : []),
    {
      label: "File",
      submenu: [
        newWindowItem(),
        { type: "separator" },
        isMac ? { role: "close" } : { role: "quit" },
      ],
    },
    { role: "editMenu" },
    {
      label: "View",
      submenu: [
        { role: "reload" },
        { role: "forceReload" },
        { type: "separator" },
        { role: "resetZoom" },
        { role: "zoomIn" },
        { role: "zoomOut" },
        { type: "separator" },
        { role: "togglefullscreen" },
      ],
    },
    {
      label: "Window",
      submenu: [
        { role: "minimize" },
        { role: "zoom" },
        ...(isMac
          ? ([
              { type: "separator" },
              { role: "front" },
              { type: "separator" },
              { role: "window" },
            ] satisfies MenuItemConstructorOptions[])
          : ([{ role: "close" }] satisfies MenuItemConstructorOptions[])),
        { type: "separator" },
        {
          label: "Bring Diff Window To Front",
          click: () => focusMainWindow(),
        },
      ],
    },
  ];

  Menu.setApplicationMenu(Menu.buildFromTemplate(template));
}
