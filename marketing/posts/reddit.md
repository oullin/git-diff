# Reddit — problem/story, honest maker voice

Reddit rewards candor and hates marketing-speak. Lead with the problem, be upfront about the beta/unsigned caveat, end by asking for feedback. Pick the title for the subreddit; the body works for all three with minor tweaks.

**Suggested images:** post `cards/screenshot.png` and `cards/features.png` in the gallery (or link them in a comment — many subs auto-filter image+link posts, so a text/self post with images in the body or first comment tends to do best). You can also use the raw `docs/images/split-view.png` and `docs/images/review-notes.png` for a more "real, un-marketed" feel.

---

## Title options

- **r/golang:** I built a local PR-review desktop app — Go engine reading Git directly, Electron/Vue front end (open source)
- **r/SideProject:** I built a macOS app to review my own diffs like a GitHub PR — before I commit
- **r/macapps:** Git Diff Review — a native macOS app for reviewing your local Git changes like a pull request (free, MIT)
- **r/programming:** Review your local Git changes like a pull request, before you commit — open source, local-only

---

## Body

I kept committing changes I never actually re-read. A debug `console.log` here, a leftover `TODO` there, a half-finished refactor I forgot to clean up — all sliding straight into history because `git diff` in the terminal is cramped and easy to skim past.

Opening a real pull request just to review my own work felt heavyweight, and it meant pushing code to a server before it was ready. So I built **Git Diff Review**: a native macOS app that gives you the GitHub-style PR review experience for your **local working tree**, before you commit.

What it does:

- A real review surface — changed-file list with status badges and +/− counts, jump between hunks
- Staged, unstaged, and untracked changes, with rename detection and a toggle to hide whitespace noise
- Split or unified diffs with syntax highlighting
- Viewed-file tracking (`0/14 viewed`) so you don't lose your place in a big change
- Inline comments + a rich-text review summary, kept across sessions
- Keyboard-driven: `j`/`k` navigate, `v` mark viewed, `c` comment, `⌘↵` submit
- Optional AI walkthroughs to get oriented on a large change fast

It's **local-only** by design: no cloud, no sign-up, no telemetry. Reviews, comments, and prefs live in a local SQLite file (`~/Library/Application Support/git-diff/reviews.sqlite3`). It reads Git over a Unix socket and never pushes your code, opens a PR, or modifies your repo — it's a review surface, nothing more.

Stack (for the curious): Electron + Vue 3 + TypeScript front end, a **Go** backend that reads Git directly, talking over a Unix domain socket. pnpm + Turbo monorepo.

**Honest caveats:** it's **beta**, **macOS on Apple Silicon only** right now, and distributed **unsigned** — so on first launch macOS Gatekeeper needs a right-click → Open (one time). Homebrew cask is planned.

- Download (.dmg): https://github.com/oullin/git-diff/releases
- Source (MIT): https://github.com/oullin/git-diff

If you review your own diffs before committing, I'd genuinely like to hear what's missing or annoying. What would make this part of your daily flow?

---

**Alt text**
- screenshot.png: "Git Diff Review showing a split diff and a changed-file sidebar in a macOS window."
- features.png: "Feature summary card: review surface, split/unified diffs, viewed tracking, inline notes, keyboard-driven, 100% local."
