# Marketing assets — Git Diff Review

Ready-to-publish pitch posts and branded social cards for launching
[Git Diff Review](https://github.com/oullin/git-diff).

## Posts

| Platform | File | Tone | Pair with |
| --- | --- | --- | --- |
| LinkedIn | [posts/linkedin.md](posts/linkedin.md) | Launch announcement | `cards/hero-og.png` + `cards/screenshot.png` |
| X / Twitter | [posts/x.md](posts/x.md) | Builder thread (5 tweets) | one card per tweet (hero → screenshot → features → privacy) |
| Reddit | [posts/reddit.md](posts/reddit.md) | Problem/story, honest maker voice | `cards/screenshot.png` + `cards/features.png` |

## Cards (`cards/*.png`)

Rendered at 2× for retina-sharp output. Built from on-brand HTML/CSS in
[`cards/src/`](cards/src/) using the app's real colors, Fira Code font, and the
existing screenshots.

| Card | Size (2×) | Aspect | Use |
| --- | --- | --- | --- |
| `hero-og.png` | 2400×1260 | 1.91:1 | Link preview / lead image (LinkedIn, X, OG tag) |
| `screenshot.png` | 2400×1260 | 1.91:1 | Product proof — `docs/images/hero.png` in a macOS window |
| `features.png` | 2160×2160 | 1:1 | Feature rundown (square, great for X / IG) |
| `privacy.png` | 2160×2160 | 1:1 | "Your code never leaves your machine" differentiator |

Raw product screenshots also live in [`../docs/images/`](../docs/images/)
(`hero`, `file-navigation`, `split-view`, `review-notes`, `walkthrough`) if you
want an un-styled, "real" look for Reddit.

## Regenerating the cards

Edit the templates in `cards/src/` (shared tokens in `card.css`), then re-render
each `.html` to a PNG with a headless browser at the exact viewport — e.g. via
Playwright with `deviceScaleFactor: 2`, waiting for `document.fonts.ready` and
image load, then `page.screenshot({ path: ... })`. Source paths are relative, so
the cards reference the repo's logo, screenshots, and fonts directly.

## CTA links used

- Releases: https://github.com/oullin/git-diff/releases
- Source: https://github.com/oullin/git-diff
- Homebrew: `brew install --cask oullin/tap/git-diff` _(planned)_
