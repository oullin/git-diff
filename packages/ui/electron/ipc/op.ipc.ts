import { ipcMain } from "electron";
import { client } from "#electron/bridge.js";
import { openTerminalCommand } from "#electron/terminal.js";

export function register(): void {
  ipcMain.handle("op:list-vaults", async () => {
    try {
      const response = await (await client()).listOpVaults();

      return { ok: true as const, vaults: response.vaults ?? [] };
    } catch (error) {
      return opErrorEnvelope(error);
    }
  });

  ipcMain.handle("op:list-items", async (_event, vault: string) => {
    try {
      const response = await (await client()).listOpItems({ vault });

      return { ok: true as const, items: response.items ?? [] };
    } catch (error) {
      return opErrorEnvelope(error);
    }
  });

  ipcMain.handle("op:signin", async () => {
    return openTerminalCommand(
      'op signin && echo "\\n[Signed in. You can close this window and return to git-diff.]"',
    );
  });

  ipcMain.handle("op:install-dependencies", async () => {
    return openTerminalCommand(
      [
        'if ! command -v brew >/dev/null 2>&1; then echo "Homebrew is required. Run ./setup.sh first, then retry."; exit 1; fi',
        "brew install --cask 1password 1password-cli",
        'echo "\\n[1Password and 1Password CLI install finished. Open 1Password, enable CLI integration if needed, then return to git-diff.]"',
      ].join("; "),
    );
  });
}

function opErrorEnvelope(error: unknown) {
  const message = error instanceof Error ? error.message : String(error);
  const code =
    error &&
    typeof error === "object" &&
    "code" in error &&
    typeof (error as { code: unknown }).code === "string"
      ? (error as { code: string }).code
      : "op_failed";

  return { ok: false as const, code, message };
}
