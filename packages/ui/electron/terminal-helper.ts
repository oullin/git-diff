import { app, dialog } from "electron";
import { chmodSync, existsSync, mkdirSync, writeFileSync } from "node:fs";
import { homedir } from "node:os";
import { dirname, join } from "node:path";

const LAUNCHER_SCRIPT = `#!/bin/sh
# Installed by Git Diff Review > Install Terminal Helper.
set -e
if [ "$#" -eq 0 ]; then
  exec open -a "Git Diff Review"
fi
resolved=""
for arg in "$@"; do
  if [ -e "$arg" ]; then
    abs=$(cd "$(dirname "$arg")" 2>/dev/null && pwd)/$(basename "$arg")
    if [ -n "$abs" ]; then
      resolved="$resolved $abs"
      continue
    fi
  fi
  resolved="$resolved $arg"
done
# shellcheck disable=SC2086
exec open -a "Git Diff Review" --args $resolved
`;

const CANDIDATE_DIRS = ["/usr/local/bin", "/opt/homebrew/bin", join(homedir(), ".local", "bin")];

function firstWritableDir(): string | null {
  for (const dir of CANDIDATE_DIRS) {
    try {
      mkdirSync(dir, { recursive: true });
      // Sentinel write/delete to verify writability without requiring sudo.
      const probe = join(dir, `.git-diff-write-check-${process.pid}`);
      writeFileSync(probe, "ok");
      try {
        require("node:fs").unlinkSync(probe);
      } catch {
        // ignore
      }
      return dir;
    } catch {
      // Not writable without sudo — keep looking.
    }
  }
  return null;
}

export async function installTerminalHelper(): Promise<void> {
  const dir = firstWritableDir();

  if (!dir) {
    const fallback = join(homedir(), ".local", "bin");
    mkdirSync(fallback, { recursive: true });
    const target = join(fallback, "git-diff");
    writeFileSync(target, LAUNCHER_SCRIPT, { mode: 0o755 });
    chmodSync(target, 0o755);
    await dialog.showMessageBox({
      type: "info",
      message: "Terminal Helper installed",
      detail: `Installed git-diff at ${target}.\n\nNo writable directory was found on $PATH. Add ${fallback} to your shell rc (e.g. .zshrc):\n\nexport PATH="${fallback}:$PATH"`,
    });
    return;
  }

  const target = join(dir, "git-diff");

  try {
    writeFileSync(target, LAUNCHER_SCRIPT, { mode: 0o755 });
    chmodSync(target, 0o755);
  } catch (error) {
    await dialog.showMessageBox({
      type: "error",
      message: "Could not install Terminal Helper",
      detail: `Failed to write ${target}: ${(error as Error).message}`,
    });
    return;
  }

  await dialog.showMessageBox({
    type: "info",
    message: "Terminal Helper installed",
    detail: `You can now run \`git-diff\`, \`git-diff <path>\`, or \`git-diff <commit-sha>\` from any terminal.\n\nLauncher installed at: ${target}`,
  });
}

export function terminalHelperPath(): string | null {
  for (const dir of CANDIDATE_DIRS) {
    const candidate = join(dir, "git-diff");
    if (existsSync(candidate)) return candidate;
  }
  return null;
}

// Touch app + dirname so the unused-import warning never fires when the file
// is included via a side-effect import. Both stay because they're useful for
// future log-to-userdata fallbacks.
void app;
void dirname;
