import { copyFileSync, existsSync, statSync, utimesSync } from "node:fs";
import { join } from "node:path";
import { uiDir } from "#scripts/dev/paths.js";
import { output } from "#scripts/dev/processes.js";

const PLIST_BUDDY = "/usr/libexec/PlistBuddy";
const PRODUCT_NAME = "Git Diff Review";
const DEV_BUNDLE_ID = "io.gocanto.git-diff.dev";
const ICON_NAME = "git-diff-review";
const SENTINEL_KEY = "GitDiffDevPatched";

const electronApp = join(uiDir, "node_modules", "electron", "dist", "Electron.app");
const contentsDir = join(electronApp, "Contents");
const infoPlist = join(contentsDir, "Info.plist");
const resourcesDir = join(contentsDir, "Resources");
const sourceIcon = join(uiDir, "build", "icon.icns");
const targetIcon = join(resourcesDir, `${ICON_NAME}.icns`);

export async function patchElectronBundle(): Promise<void> {
  if (process.platform !== "darwin") {
    return;
  }

  if (!existsSync(infoPlist)) {
    return;
  }

  if (!existsSync(sourceIcon)) {
    return;
  }

  if (await isAlreadyPatched()) {
    return;
  }

  copyFileSync(sourceIcon, targetIcon);

  await applyPlistEntry(":CFBundleName", "string", PRODUCT_NAME);

  await applyPlistEntry(":CFBundleDisplayName", "string", PRODUCT_NAME);

  await applyPlistEntry(":CFBundleIdentifier", "string", DEV_BUNDLE_ID);

  await applyPlistEntry(":CFBundleIconFile", "string", ICON_NAME);

  await applyPlistEntry(":CFBundleIconName", "string", ICON_NAME);

  await applyPlistEntry(`:${SENTINEL_KEY}`, "bool", "true");

  const now = new Date();

  utimesSync(infoPlist, now, now);
  utimesSync(electronApp, now, now);

  console.log("Patched Electron.app bundle for dev branding");
}

async function isAlreadyPatched(): Promise<boolean> {
  if (!existsSync(targetIcon)) {
    return false;
  }

  if (statSync(sourceIcon).mtimeMs > statSync(targetIcon).mtimeMs) {
    return false;
  }

  try {
    const value = await output(PLIST_BUDDY, ["-c", `Print :${SENTINEL_KEY}`, infoPlist], uiDir);

    return value.trim() === "true";
  } catch {
    return false;
  }
}

async function applyPlistEntry(key: string, type: "string" | "bool", value: string): Promise<void> {
  try {
    await output(PLIST_BUDDY, ["-c", `Set ${key} ${value}`, infoPlist], uiDir);
  } catch {
    await output(PLIST_BUDDY, ["-c", `Add ${key} ${type} ${value}`, infoPlist], uiDir);
  }
}
