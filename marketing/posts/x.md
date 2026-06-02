# X / Twitter — builder thread

Builder/indie-hacker voice. 5 tweets, one image each. Post 1/ as the hook, reply-chain the rest.

---

**1/** — attach `cards/hero-og.png`

you commit code you never actually re-read.

the debug log. the stray TODO. the half-finished refactor. straight into history.

so i built Git Diff Review: a full GitHub-style PR review for your *local* changes — before you commit. on your machine, nothing leaves it. 🧵

---

**2/** — attach `cards/screenshot.png`

`git diff` in the terminal is cramped. opening a real PR just to review your own work is overkill — and it pushes your code to a server before it's ready.

this gives you the review surface you already know: file list, status badges, split diffs, viewed checkboxes. locally.

---

**3/** — attach `cards/features.png`

what you get:

• split or unified diffs, syntax highlighted
• viewed-file tracking (0/14 viewed)
• inline comments + a review summary
• rename detection, hide whitespace noise
• keyboard-driven: j/k navigate · v viewed · c comment · ⌘↵ submit

review without touching the mouse.

---

**4/** — attach `cards/privacy.png`

the part i care about most: it's 100% local.

no cloud. no sign-up. no telemetry. reviews + comments live in a local sqlite file. it reads git over a unix socket and never pushes or opens a PR for you.

your code stays yours.

---

**5/** — (no image, or reuse hero-og)

stack: electron + vue front end, go engine reading git directly. fast + native.

free, open source (MIT), beta for macOS on apple silicon 👇

download: https://github.com/oullin/git-diff/releases
source: https://github.com/oullin/git-diff

would love your feedback.

---

**Single-tweet version** (if you don't want a thread) — attach `cards/hero-og.png`:

you commit code you never re-read — debug logs, stray TODOs, half-finished refactors slip into history.

Git Diff Review = a full GitHub-style PR review for your local changes, before you commit. 100% local, open source, macOS.

https://github.com/oullin/git-diff/releases
