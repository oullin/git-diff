// @vitest-environment happy-dom
import { describe, expect, test } from "vitest";
import { formatReviewAsMarkdown, htmlToMarkdown } from "../src/lib/reviewMarkdown";
import type { ReviewComment, ReviewDetail, ReviewSession } from "@git-diff/contracts";

function comment(overrides: Partial<ReviewComment> = {}): ReviewComment {
  return {
    id: "c-1",
    reviewId: "r-1",
    filePath: "src/App.vue",
    diffSection: "unstaged:src/App.vue",
    side: "add",
    lineNumber: 12,
    authorLabel: "Reviewer",
    bodyHtml: "<p>looks good</p>",
    createdAt: "2026-05-20T10:00:00Z",
    updatedAt: "2026-05-20T10:00:00Z",
    ...overrides,
  };
}

function session(overrides: Partial<ReviewSession> = {}): ReviewSession {
  return {
    id: "r-1",
    repoRoot: "/repo",
    userId: 1,
    branch: "feat/x",
    headSha: "abc1234",
    status: "open",
    title: "Review feat/x",
    summary: "",
    filesChanged: 1,
    additions: 4,
    deletions: 1,
    startedAt: "2026-05-20T09:55:00Z",
    contextKind: "working",
    ...overrides,
  };
}

describe("htmlToMarkdown", () => {
  test("paragraphs and bold/italic", () => {
    expect(htmlToMarkdown("<p>hello <strong>world</strong> <em>now</em></p>")).toBe(
      "hello **world** _now_",
    );
  });

  test("inline code and code block", () => {
    expect(htmlToMarkdown("<p>see <code>foo</code></p>")).toBe("see `foo`");
    expect(htmlToMarkdown("<pre><code>const x = 1\nconst y = 2</code></pre>")).toContain(
      "```\nconst x = 1\nconst y = 2\n```",
    );
  });

  test("ordered and unordered lists", () => {
    expect(htmlToMarkdown("<ul><li>a</li><li>b</li></ul>")).toBe("- a\n- b");
    expect(htmlToMarkdown("<ol><li>a</li><li>b</li></ol>")).toBe("1. a\n2. b");
  });

  test("links", () => {
    expect(htmlToMarkdown('<p><a href="https://x">x</a></p>')).toBe("[x](https://x)");
  });

  test("blockquote", () => {
    expect(htmlToMarkdown("<blockquote>quoted</blockquote>")).toBe("> quoted");
  });

  test("unknown tags collapse to text", () => {
    expect(htmlToMarkdown("<section>ok</section>")).toBe("ok");
  });
});

describe("formatReviewAsMarkdown", () => {
  test("includes header, branch metadata, and grouped comments", () => {
    const detail: ReviewDetail = {
      review: session(),
      events: [],
      comments: [
        comment({ lineNumber: 12, bodyHtml: "<p>looks <strong>great</strong></p>" }),
        comment({ id: "c-2", lineNumber: 20, bodyHtml: "<p>could be cleaner</p>" }),
        comment({
          id: "c-3",
          filePath: "src/main.ts",
          lineNumber: 1,
          bodyHtml: "<p>nit</p>",
        }),
      ],
    };
    const out = formatReviewAsMarkdown(detail);
    expect(out).toContain("# Review feat/x");
    expect(out).toContain("Branch: `feat/x`");
    expect(out).toContain("HEAD: `abc1234`");
    expect(out).toContain("## `src/App.vue`");
    expect(out).toContain("## `src/main.ts`");
    expect(out).toContain("**L12** (new)");
    expect(out).toContain("looks **great**");
    expect(out.indexOf("L12")).toBeLessThan(out.indexOf("L20"));
  });

  test("commit-mode review shows the commit SHA", () => {
    const detail: ReviewDetail = {
      review: session({ contextKind: "commit", contextSha: "deadbeef", branch: "deadbef" }),
      events: [],
      comments: [comment()],
    };
    expect(formatReviewAsMarkdown(detail)).toContain("Commit: `deadbeef`");
  });

  test("deleted comments are omitted", () => {
    const detail: ReviewDetail = {
      review: session(),
      events: [],
      comments: [comment({ deletedAt: "2026-05-20T11:00:00Z", bodyHtml: "<p>oops</p>" })],
    };
    expect(formatReviewAsMarkdown(detail)).toContain("_No comments yet._");
  });

  test("empty review surfaces a no-comments line", () => {
    const detail: ReviewDetail = { review: session(), events: [], comments: [] };
    expect(formatReviewAsMarkdown(detail)).toContain("_No comments yet._");
  });
});
