import { existsSync, statSync } from "node:fs";
import { isAbsolute, resolve } from "node:path";

const SHA_PATTERN = /^[0-9a-f]{4,40}$/i;

export type LaunchIntentKind = "working" | "commit" | "help";

export interface LaunchIntent {
  kind: LaunchIntentKind;
  /** Resolved absolute path of the repository the UI should open, when known. */
  repoPath?: string;
  /** Commit SHA (raw, as passed) when kind === "commit". */
  sha?: string;
  /** True when the -w / --walkthrough flag was set. */
  walkthrough: boolean;
  /** Set when --help / -h / help was requested. Renderer should not consume. */
  helpText?: string;
  /** Raw positional arg from argv, for diagnostics. */
  raw?: string;
}

const USAGE_TEXT = `git-diff serves the local diff review UI.

Usage:
  git-diff                 open the most recently used repository (or pick one)
  git-diff <path>          open a specific git repository
  git-diff <sha>           open a specific commit from the current repository
  git-diff -w [<sha>]      open and request an AI walkthrough (planned)
  git-diff -h | --help     print this help and exit
`;

/**
 * Parse the raw process.argv into a LaunchIntent. Electron prepends its own
 * argv entries (the binary path, and in dev the app path); strip those before
 * applying the user-facing grammar.
 *
 * The parser intentionally accepts both `-w <sha>` and `<sha> -w`; flag order
 * is irrelevant.
 *
 * @param argv      The argv array to consume (typically process.argv).
 * @param isPackaged Whether the app is running from a packaged build. Strips
 *                  one extra entry in dev (the app path passed to `electron .`).
 * @param cwd       The working directory to resolve relative paths against.
 */
export function parseLaunchArgs(argv: string[], isPackaged: boolean, cwd: string): LaunchIntent {
  const userArgs = argv.slice(isPackaged ? 1 : 2).filter((arg) => !isElectronInternal(arg));

  let walkthrough = false;
  let positional: string | undefined;

  for (const arg of userArgs) {
    if (arg === "-h" || arg === "--help" || arg === "help") {
      return { kind: "help", walkthrough: false, helpText: USAGE_TEXT };
    }

    if (arg === "-w" || arg === "--walkthrough") {
      walkthrough = true;
      continue;
    }

    if (arg.startsWith("-")) {
      // Unknown flag — fall through to help so the user gets immediate feedback.
      return {
        kind: "help",
        walkthrough: false,
        helpText: `Unknown flag: ${arg}\n\n${USAGE_TEXT}`,
      };
    }

    if (!positional) {
      positional = arg;
      continue;
    }

    // Too many positional args.
    return {
      kind: "help",
      walkthrough: false,
      helpText: `Too many arguments. Expected at most one path or commit SHA.\n\n${USAGE_TEXT}`,
    };
  }

  if (!positional) {
    return { kind: "working", walkthrough };
  }

  return classifyPositional(positional, cwd, walkthrough);
}

function classifyPositional(arg: string, cwd: string, walkthrough: boolean): LaunchIntent {
  // A real directory always wins over the SHA pattern, even if its name looks
  // hex. This matches what users expect when they `cd into-a-dir; git-diff .`.
  const resolved = isAbsolute(arg) ? arg : resolve(cwd, arg);

  if (existsSync(resolved)) {
    try {
      if (statSync(resolved).isDirectory()) {
        return { kind: "working", repoPath: resolved, walkthrough, raw: arg };
      }
    } catch {
      // fall through to SHA classification
    }
  }

  if (SHA_PATTERN.test(arg)) {
    // SHA argument — repoPath stays unset; renderer uses the currently active
    // repo or, when launched fresh, the cwd.
    return { kind: "commit", sha: arg, repoPath: cwd, walkthrough, raw: arg };
  }

  // Neither an existing path nor a SHA — surface help.
  return {
    kind: "help",
    walkthrough: false,
    helpText: `Could not resolve "${arg}" as a path or commit SHA.\n\n${USAGE_TEXT}`,
  };
}

function isElectronInternal(arg: string): boolean {
  // Electron forwards these to the renderer; they're not user arguments.
  return (
    arg.startsWith("--remote-debugging-") ||
    arg.startsWith("--inspect") ||
    arg.startsWith("--enable-") ||
    arg.startsWith("--disable-") ||
    arg.startsWith("--user-data-dir") ||
    arg.startsWith("--no-sandbox") ||
    arg === "--squirrel-firstrun"
  );
}
