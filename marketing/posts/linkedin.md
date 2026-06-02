# LinkedIn — launch announcement

**Suggested images:** `cards/hero-og.png` (lead), `cards/screenshot.png` (second image). LinkedIn shows up to ~2–4 images well; lead with the hero card.

---

Introducing **Git Diff Review** — a native macOS app that gives you a full GitHub-style pull-request review for your *local* changes, before you commit.

Here's the gap it fills:

`git diff` in the terminal is cramped and easy to skim past. Opening a real pull request just to look over your own work is heavyweight — and it pushes your code onto a server before it's ready. So we commit changes we never actually re-read, and the stray debug log, the leftover TODO, the half-finished refactor slip into history.

Git Diff Review gives you the review experience you already know — but locally:

🔍 A real review surface — changed-file list, status badges, +/− counts, jump between hunks
🪟 Split or unified diffs with syntax highlighting
✅ Viewed-file tracking, so you never lose your place in a big change (0/14 viewed)
📝 Inline comments and a rich-text review summary, stored locally
⌨️ Keyboard-driven — j/k to navigate, v to mark viewed, c to comment, ⌘↵ to submit
🔒 100% local — no cloud, no sign-up, no telemetry. Reviews live in a local SQLite file and never leave your machine.

Under the hood it's an Electron + Vue interface backed by a Go engine that reads Git directly — fast, and it never pushes your code or opens a PR for you. It's a review surface, nothing more.

It's **free, open source (MIT), and in beta** for macOS on Apple Silicon.

Grab the latest build: https://github.com/oullin/git-diff/releases
Source + details: https://github.com/oullin/git-diff

If you review your own diffs before committing, I'd love your feedback. 🙏

#git #macOS #developertools #opensource #golang #vuejs #softwareengineering #devtools

---

**Alt text**
- hero-og.png: "Git Diff Review app card — logo and tagline: review your local Git changes like a pull request, on your machine, nothing leaves it."
- screenshot.png: "Screenshot of Git Diff Review showing a split diff with a changed-file sidebar, framed in a macOS window."
