import { existsSync, statSync } from "node:fs";
import { isAbsolute, resolve } from "node:path";

// Three patterns define the launch grammar:
//
//   commitHashPattern    : 4-64 hex chars — short or full SHAs.
//   headCommitRefPattern : HEAD / @ revision syntax — HEAD, HEAD~3, HEAD^,
//                          HEAD^{commit}, @{1}, etc. Anything `git rev-parse`
//                          would accept that starts at HEAD or @.
//   pullRequestNumberPattern : the `#42` shorthand for `pr 42`.
const commitHashPattern = /^[0-9a-f]{4,64}$/i;
const headCommitRefPattern = /^(?:HEAD|@)(?:(?:[~^]\d*)|\^\{[^}]+\}|@?\{[^}]+\})*$/;
const pullRequestNumberPattern = /^#([1-9]\d*)$/;
const pullRequestUrlPattern =
    /^https?:\/\/github\.com\/[^/\s]+\/[^/\s]+\/pull\/([1-9]\d*)(?:[/?#].*)?$/i;
const numericPattern = /^[0-9]{1,7}$/;

export type LaunchIntentKind = "working" | "commit" | "pull-request" | "help";

export interface LaunchIntent {
    kind: LaunchIntentKind;
    /** Resolved absolute path of the repository the UI should open, when known. */
    repoPath?: string;
    /**
     * The commit-ish the user asked for when kind === "commit". Accepts raw
     * SHAs and HEAD/@ revision syntax; the backend resolves with
     * `git rev-parse --verify <ref>^{commit}`.
     */
    commitRef?: string;
    /**
     * @deprecated Use {@link commitRef}. Retained for one release so existing
     * renderer code that reads `intent.sha` keeps working.
     */
    sha?: string;
    /** PR number when kind === "pull-request". */
    pullRequestNumber?: number;
    /**
     * @deprecated Use {@link pullRequestNumber}. Retained for one release.
     */
    prNumber?: number;
    /** Source GitHub URL when the user launched via `pr <url>`. */
    pullRequestUrl?: string;
    /** True when the -w / --walkthrough flag was set. */
    walkthrough: boolean;
    /** Set when --help / -h / help was requested. Renderer should not consume. */
    helpText?: string;
    /** Raw positional arg from argv, for diagnostics. */
    raw?: string;
}

const USAGE_TEXT = `git-diff serves the local diff review UI.

Usage:
  git-diff                    open the most recently used repository (or pick one)
  git-diff <path>             open a specific git repository
  git-diff <sha>              open a specific commit (4-64 hex chars)
  git-diff HEAD[~N|^N]        open a commit by revision syntax (HEAD, HEAD~3, HEAD^)
  git-diff #<number>          open a pull request (shorthand for \`pr <number>\`)
  git-diff pr <number>        open an open pull request (requires the gh CLI)
  git-diff pr <github-url>    open a pull request by its GitHub URL
  git-diff -w [<ref>]         open and request an AI walkthrough
  git-diff -h | --help        print this help and exit
`;

/**
 * Electron prepends its own argv entries (the binary path, and in dev the app
 * path); strip those before applying the user-facing grammar. Flag order is
 * intentionally irrelevant: both `-w <ref>` and `<ref> -w` are accepted.
 */
export function parseLaunchArgs(argv: string[], isPackaged: boolean, cwd: string): LaunchIntent {
    const userArgs = argv.slice(isPackaged ? 1 : 2).filter((arg) => !isElectronInternal(arg));

    let walkthrough = false;
    const positionals: string[] = [];

    for (const arg of userArgs) {
        if (arg === "-h" || arg === "--help" || arg === "help") {
            return { kind: "help", walkthrough: false, helpText: USAGE_TEXT };
        }

        if (arg === "-w" || arg === "--walkthrough") {
            walkthrough = true;
            continue;
        }

        if (arg.startsWith("-")) {
            return {
                kind: "help",
                walkthrough: false,
                helpText: `Unknown flag: ${arg}\n\n${USAGE_TEXT}`,
            };
        }

        positionals.push(arg);
    }

    if (positionals[0] === "pr") {
        return classifyPrSubcommand(positionals.slice(1), cwd, walkthrough);
    }

    if (positionals.length === 0) {
        return { kind: "working", walkthrough };
    }

    if (positionals.length === 1) {
        return classifyPositional(positionals[0]!, cwd, walkthrough);
    }

    if (positionals.length === 2) {
        const pathArg = positionals[0]!;
        const refArg = positionals[1]!;
        const resolvedPath = resolveExistingDir(pathArg, cwd);

        if (!resolvedPath) {
            return {
                kind: "help",
                walkthrough: false,
                helpText: `First argument "${pathArg}" is not a directory.\n\n${USAGE_TEXT}`,
            };
        }

        return classifyPositional(refArg, resolvedPath, walkthrough, {
            repoPathOverride: resolvedPath,
        });
    }

    return {
        kind: "help",
        walkthrough: false,
        helpText: `Too many arguments.\n\n${USAGE_TEXT}`,
    };
}

interface ClassifyOptions {
    /** Force a specific repoPath instead of using cwd. */
    repoPathOverride?: string;
}

function classifyPositional(
    arg: string,
    cwd: string,
    walkthrough: boolean,
    opts: ClassifyOptions = {},
): LaunchIntent {
    // A real directory always wins over the SHA / HEAD pattern, even if its
    // name looks hex. Matches what users expect when they `cd into-a-dir; git-diff .`.
    if (!opts.repoPathOverride) {
        const resolved = resolveExistingDir(arg, cwd);

        if (resolved) {
            return { kind: "working", repoPath: resolved, walkthrough, raw: arg };
        }
    }

    const repoPath = opts.repoPathOverride ?? cwd;

    if (isCommitRefArgument(arg, cwd)) {
        return {
            kind: "commit",
            commitRef: arg,
            sha: arg,
            repoPath,
            walkthrough,
            raw: arg,
        };
    }

    const prMatch = arg.match(pullRequestNumberPattern);

    if (prMatch) {
        const number = Number(prMatch[1]);

        return {
            kind: "pull-request",
            pullRequestNumber: number,
            prNumber: number,
            repoPath,
            walkthrough,
            raw: arg,
        };
    }

    const urlMatch = arg.match(pullRequestUrlPattern);

    if (urlMatch) {
        const number = Number(urlMatch[1]);

        return {
            kind: "pull-request",
            pullRequestNumber: number,
            prNumber: number,
            pullRequestUrl: arg,
            repoPath,
            walkthrough,
            raw: arg,
        };
    }

    return {
        kind: "help",
        walkthrough: false,
        helpText: `Could not resolve "${arg}" as a path, commit ref, or pull request.\n\n${USAGE_TEXT}`,
    };
}

function classifyPrSubcommand(args: string[], cwd: string, walkthrough: boolean): LaunchIntent {
    if (args.length !== 1) {
        return {
            kind: "help",
            walkthrough: false,
            helpText: `\`pr\` subcommand requires a number or GitHub URL.\n\n${USAGE_TEXT}`,
        };
    }

    const target = args[0]!;
    const urlMatch = target.match(pullRequestUrlPattern);

    if (urlMatch) {
        const number = Number(urlMatch[1]);

        return {
            kind: "pull-request",
            pullRequestNumber: number,
            prNumber: number,
            pullRequestUrl: target,
            repoPath: cwd,
            walkthrough,
            raw: target,
        };
    }

    if (numericPattern.test(target)) {
        const number = Number(target);

        return {
            kind: "pull-request",
            pullRequestNumber: number,
            prNumber: number,
            repoPath: cwd,
            walkthrough,
            raw: target,
        };
    }

    return {
        kind: "help",
        walkthrough: false,
        helpText: `\`pr\` expected a number or GitHub URL, got "${target}".\n\n${USAGE_TEXT}`,
    };
}

/**
 * True when `arg` is a commit-ish the backend should try to resolve. A path
 * with the same name on disk always wins, so callers must check existence
 * first when ambiguous.
 */
export function isCommitRefArgument(arg: string, cwd: string): boolean {
    if (commitHashPattern.test(arg)) {
        return true;
    }

    if (!headCommitRefPattern.test(arg)) {
        return false;
    }

    // Disambiguate `HEAD~3` from a file literally named `HEAD~3`.
    return resolveExistingDir(arg, cwd) === null && !existsSync(resolveAbsolute(arg, cwd));
}

function resolveExistingDir(arg: string, cwd: string): string | null {
    const resolved = resolveAbsolute(arg, cwd);

    if (!existsSync(resolved)) {
        return null;
    }

    try {
        return statSync(resolved).isDirectory() ? resolved : null;
    } catch {
        return null;
    }
}

function resolveAbsolute(arg: string, cwd: string): string {
    return isAbsolute(arg) ? arg : resolve(cwd, arg);
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
