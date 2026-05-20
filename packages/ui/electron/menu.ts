import { app, dialog, Menu, type MenuItemConstructorOptions } from "electron";
import { createWindow, focusMainWindow, getMainWindow } from "#electron/windows.js";

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
