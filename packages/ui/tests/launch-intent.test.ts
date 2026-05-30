import { describe, expect, test } from "vitest";
import { mkdtempSync, writeFileSync, mkdirSync, rmSync } from "node:fs";
import { join } from "node:path";
import { tmpdir } from "node:os";
import { parseLaunchArgs } from "@electron/launch-intent.js";

function tempRepo(): string {
    const dir = mkdtempSync(join(tmpdir(), "launch-intent-"));

    mkdirSync(join(dir, "subdir"), { recursive: true });
    writeFileSync(join(dir, "file.txt"), "hello\n");

    return dir;
}

describe("parseLaunchArgs", () => {
    test("no args -> working mode against cwd", () => {
        const intent = parseLaunchArgs(["/bin/app"], true, "/some/cwd");

        expect(intent.kind).toBe("working");
        expect(intent.repoPath).toBeUndefined();
        expect(intent.walkthrough).toBe(false);
    });

    test("path arg becomes resolved repoPath", () => {
        const repo = tempRepo();

        try {
            const intent = parseLaunchArgs(["/bin/app", "subdir"], true, repo);

            expect(intent.kind).toBe("working");
            expect(intent.repoPath).toBe(join(repo, "subdir"));
        } finally {
            rmSync(repo, { recursive: true, force: true });
        }
    });

    test("absolute path arg passes through", () => {
        const repo = tempRepo();

        try {
            const intent = parseLaunchArgs(["/bin/app", repo], true, "/elsewhere");

            expect(intent.kind).toBe("working");
            expect(intent.repoPath).toBe(repo);
        } finally {
            rmSync(repo, { recursive: true, force: true });
        }
    });

    test("SHA-like arg becomes commit mode with cwd as repoPath", () => {
        const intent = parseLaunchArgs(["/bin/app", "abc1234"], true, "/repo");

        expect(intent.kind).toBe("commit");
        expect(intent.commitRef).toBe("abc1234");
        expect(intent.sha).toBe("abc1234"); // deprecated alias still emitted
        expect(intent.repoPath).toBe("/repo");
    });

    test("full 40-char SHA is accepted", () => {
        const sha = "a".repeat(40);
        const intent = parseLaunchArgs(["/bin/app", sha], true, "/repo");

        expect(intent.kind).toBe("commit");
        expect(intent.commitRef).toBe(sha);
    });

    test("64-char SHA-256 hash is accepted", () => {
        const sha = "f".repeat(64);
        const intent = parseLaunchArgs(["/bin/app", sha], true, "/repo");

        expect(intent.kind).toBe("commit");
        expect(intent.commitRef).toBe(sha);
    });

    test("HEAD~N revision syntax becomes a commit intent", () => {
        const intent = parseLaunchArgs(["/bin/app", "HEAD~3"], true, "/repo");

        expect(intent.kind).toBe("commit");
        expect(intent.commitRef).toBe("HEAD~3");
        expect(intent.sha).toBe("HEAD~3");
    });

    test("HEAD^ revision syntax becomes a commit intent", () => {
        const intent = parseLaunchArgs(["/bin/app", "HEAD^"], true, "/repo");

        expect(intent.kind).toBe("commit");
        expect(intent.commitRef).toBe("HEAD^");
    });

    test("bare HEAD is a commit intent", () => {
        const intent = parseLaunchArgs(["/bin/app", "HEAD"], true, "/repo");

        expect(intent.kind).toBe("commit");
        expect(intent.commitRef).toBe("HEAD");
    });

    test("reflog syntax @{1} is accepted", () => {
        const intent = parseLaunchArgs(["/bin/app", "@{1}"], true, "/repo");

        expect(intent.kind).toBe("commit");
        expect(intent.commitRef).toBe("@{1}");
    });

    test("existing directory wins over SHA pattern", () => {
        const repo = tempRepo();
        const shaLike = join(repo, "abc1234");

        mkdirSync(shaLike);
        try {
            const intent = parseLaunchArgs(["/bin/app", "abc1234"], true, repo);

            expect(intent.kind).toBe("working");
            expect(intent.repoPath).toBe(shaLike);
        } finally {
            rmSync(repo, { recursive: true, force: true });
        }
    });

    test("existing path named HEAD~3 wins over revision syntax", () => {
        const repo = tempRepo();
        const headPath = join(repo, "HEAD~3");

        mkdirSync(headPath);
        try {
            const intent = parseLaunchArgs(["/bin/app", "HEAD~3"], true, repo);

            expect(intent.kind).toBe("working");
            expect(intent.repoPath).toBe(headPath);
        } finally {
            rmSync(repo, { recursive: true, force: true });
        }
    });

    test("-w sets walkthrough", () => {
        const intent = parseLaunchArgs(["/bin/app", "-w"], true, "/repo");

        expect(intent.kind).toBe("working");
        expect(intent.walkthrough).toBe(true);
    });

    test("-w combines with SHA in either order", () => {
        const a = parseLaunchArgs(["/bin/app", "-w", "abc1234"], true, "/repo");
        const b = parseLaunchArgs(["/bin/app", "abc1234", "-w"], true, "/repo");

        expect(a.kind).toBe("commit");
        expect(a.walkthrough).toBe(true);
        expect(a.commitRef).toBe("abc1234");
        expect(b).toEqual(a);
    });

    test("--help returns help intent with usage text", () => {
        const intent = parseLaunchArgs(["/bin/app", "--help"], true, "/repo");

        expect(intent.kind).toBe("help");
        expect(intent.helpText).toContain("Usage:");
    });

    test("unknown flag returns help with explanation", () => {
        const intent = parseLaunchArgs(["/bin/app", "--bogus"], true, "/repo");

        expect(intent.kind).toBe("help");
        expect(intent.helpText).toContain("Unknown flag");
    });

    test("non-existent non-SHA arg returns help with hint", () => {
        const intent = parseLaunchArgs(["/bin/app", "not-a-path-or-sha-xyz"], true, "/tmp");

        expect(intent.kind).toBe("help");
        expect(intent.helpText).toContain("not-a-path-or-sha-xyz");
    });

    test("three positionals -> help", () => {
        const intent = parseLaunchArgs(["/bin/app", "a", "b", "c"], true, "/tmp");

        expect(intent.kind).toBe("help");
        expect(intent.helpText).toContain("Too many arguments");
    });

    test("dev mode strips two argv entries (electron .)", () => {
        const repo = tempRepo();

        try {
            const intent = parseLaunchArgs(
                ["/path/to/electron", "/path/to/app", repo],
                false,
                "/elsewhere",
            );

            expect(intent.repoPath).toBe(repo);
        } finally {
            rmSync(repo, { recursive: true, force: true });
        }
    });

    test("pr <number> becomes a pull-request intent", () => {
        const intent = parseLaunchArgs(["/bin/app", "pr", "42"], true, "/repo");

        expect(intent.kind).toBe("pull-request");
        expect(intent.pullRequestNumber).toBe(42);
        expect(intent.prNumber).toBe(42); // deprecated alias still emitted
        expect(intent.repoPath).toBe("/repo");
    });

    test("pr <github-url> resolves to the same intent and carries the URL", () => {
        const url = "https://github.com/anthropics/claude-code/pull/42";
        const intent = parseLaunchArgs(["/bin/app", "pr", url], true, "/repo");

        expect(intent.kind).toBe("pull-request");
        expect(intent.pullRequestNumber).toBe(42);
        expect(intent.pullRequestUrl).toBe(url);
    });

    test("#<number> shorthand routes to pull-request", () => {
        const intent = parseLaunchArgs(["/bin/app", "#7"], true, "/repo");

        expect(intent.kind).toBe("pull-request");
        expect(intent.pullRequestNumber).toBe(7);
    });

    test("github pull URL as a single arg routes to pull-request", () => {
        const url = "https://github.com/anthropics/claude-code/pull/99/files";
        const intent = parseLaunchArgs(["/bin/app", url], true, "/repo");

        expect(intent.kind).toBe("pull-request");
        expect(intent.pullRequestNumber).toBe(99);
        expect(intent.pullRequestUrl).toBe(url);
    });

    test("pr without an argument returns help", () => {
        const intent = parseLaunchArgs(["/bin/app", "pr"], true, "/repo");

        expect(intent.kind).toBe("help");
        expect(intent.helpText).toContain("pr` subcommand requires");
    });

    test("pr with a non-numeric non-URL arg returns help", () => {
        const intent = parseLaunchArgs(["/bin/app", "pr", "abc"], true, "/repo");

        expect(intent.kind).toBe("help");
        expect(intent.helpText).toContain("expected a number or GitHub URL");
    });

    test("two positionals: <path> <ref> opens commit in that repo", () => {
        const repo = tempRepo();

        try {
            const intent = parseLaunchArgs(["/bin/app", repo, "HEAD~2"], true, "/elsewhere");

            expect(intent.kind).toBe("commit");
            expect(intent.repoPath).toBe(repo);
            expect(intent.commitRef).toBe("HEAD~2");
        } finally {
            rmSync(repo, { recursive: true, force: true });
        }
    });

    test("two positionals where first is not a directory -> help", () => {
        const intent = parseLaunchArgs(["/bin/app", "not-a-dir", "HEAD~2"], true, "/tmp");

        expect(intent.kind).toBe("help");
        expect(intent.helpText).toContain("not a directory");
    });

    test("Electron internal flags are ignored", () => {
        const intent = parseLaunchArgs(
            ["/bin/app", "--no-sandbox", "--remote-debugging-port=9222", "abc1234"],
            true,
            "/repo",
        );

        expect(intent.kind).toBe("commit");
        expect(intent.commitRef).toBe("abc1234");
    });
});
