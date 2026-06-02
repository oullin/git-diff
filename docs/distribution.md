# Distribution

This document covers how the macOS app is built, signed, notarized, and
distributed. Packaging is handled by **Electron Forge**
(`packages/ui/forge.config.ts`); the legacy electron-builder path has been
removed.

## Status

| Surface | State |
|---|---|
| `electron-forge` (current) | `packages/ui/forge.config.ts`; `pnpm -C packages/ui run make:mac` (local) / `pnpm release:mac:unsigned` (root, builds the Go API + renderer first) |
| GitHub Releases publishing | Automated on `v*` tags via `.github/workflows/release.yml` (unsigned Forge build); a Forge GitHub publisher is also configured for `pnpm -C packages/ui run publish` |
| Auto-update | Not wired |
| Homebrew cask | Template at `packages/ui/scripts/Casks/git-diff.rb`; release workflow opens a bump PR to `oullin/homebrew-tap` |
| Terminal helper | Implemented — menu item "Install Terminal Helper…" writes a launcher script |

## Packaging with Electron Forge

The Forge scripts in `packages/ui/package.json`:

```json
"package": "electron-forge package",
"make": "electron-forge make",
"make:mac": "electron-forge make --platform=darwin --arch=arm64",
"publish": "electron-forge publish"
```

`forge.config.ts` regenerates the app icon (`generateAssets` hook), embeds the
Go API binary as an `extraResource`, and produces an unsigned `.dmg` (ULFO) +
`.zip` on darwin/arm64. Forge writes its makers' output under
`packages/ui/out/make/` (the ZIP nested under `out/make/zip/darwin/arm64/`); the
release workflow copies those to the canonical
`git-diff-review-{version}-arm64.{dmg,zip}` names the cask `url` expects.

## Bundled Go API binary

The macOS app ships with the Go API binary embedded as a Forge
`extraResource`. The pipeline is:

1. `go build` produces `packages/api/dist/api`.
2. `forge.config.ts` lists that path under `packagerConfig.extraResource`,
   which copies it into `Git Diff Review.app/Contents/Resources/api`.
3. At runtime, `packages/ui/electron/bridge.ts` resolves the binary via
   `join(process.resourcesPath, "api")` when `app.isPackaged` is true,
   and falls back to `go run ./cmd` in development.

The binary must be built before `electron-forge package` runs; otherwise
the packaged app launches without a working bridge. The release workflow
in `.github/workflows/release.yml` sequences this correctly.

## Auto-update wiring

Not yet wired. To enable it:

```ts
// packages/ui/electron/main.ts
import { updateElectronApp } from "update-electron-app";
if (app.isPackaged) updateElectronApp({ updateInterval: "6 hours" });
```

`update-electron-app` reads the GitHub Releases feed declared in the Forge
publisher config and triggers Squirrel.Mac restarts.

## Homebrew tap

The tap lives at [`oullin/homebrew-tap`](https://github.com/oullin/homebrew-tap)
and is already seeded with `Casks/git-diff.rb`.

1. The source of truth for the cask is `packages/ui/scripts/Casks/git-diff.rb`
   in this repo; the `bump-cask` release job substitutes `version` and `sha256`
   from the published DMG and opens a PR against the tap.
2. The job authenticates with a `HOMEBREW_TAP_TOKEN` repo secret — a PAT with
   write access to the tap (the default `GITHUB_TOKEN` can't push cross-repo).
   Set it once with `gh secret set HOMEBREW_TAP_TOKEN` (or via the GitHub UI)
   before the first tagged release.
3. End users install with:
   ```
   brew install --cask oullin/tap/git-diff
   ```

## Terminal helper

The app menu offers **Git Diff Review → Install Terminal Helper…**. It
writes an executable launcher to the first writable directory among:

- `/usr/local/bin`
- `/opt/homebrew/bin`
- `~/.local/bin` (fallback; adds a `$PATH` reminder if needed)

After installation:

```
git-diff                  # open most-recently-used repo
git-diff /path/to/repo    # open that repo
git-diff <commit-sha>     # open commit in cwd
git-diff -w               # open with AI walkthrough
```

The launcher forwards argv to `open -a "Git Diff Review" --args …`. Existing
filesystem paths are resolved to absolute paths first so the app sees the
user's actual cwd; commit SHAs and flags pass through unchanged.

## Release workflow

`.github/workflows/release.yml` runs on a `v*` tag: an idempotent
`create-release` job fans out to a per-platform build job.

1. **`create-release`** (ubuntu) — `gh release view "$TAG" || gh release create
   "$TAG" --generate-notes --title "$TAG" --verify-tag`. Idempotent, so re-runs
   reuse the existing release.
2. **`build-macos`** (macos-14) — sets up pnpm, Node 23, and Go (from
   `packages/api/go.mod`); builds the Go API binary into `packages/api/dist/api`,
   the shared TS libraries, and the renderer + electron bundles; runs
   `pnpm -C packages/ui run make:mac`; canonicalizes the Forge outputs to
   `git-diff-review-{version}-arm64.{dmg,zip}`, writes `SHASUMS256.txt`, and
   uploads all three to the release with `gh release upload --clobber`.
3. **`bump-cask`** (ubuntu, final tags only) — opens a PR to `oullin/homebrew-tap`
   bumping the cask `version` + `sha256`. Requires a `HOMEBREW_TAP_TOKEN` secret
   (a PAT with write access to the tap — the default `GITHUB_TOKEN` cannot push
   cross-repo).

The build ships **arm64-only and unsigned** today. Tag/artifact convention:
tag `v{version}`, DMG `git-diff-review-{version}-arm64.dmg` (matches the cask
`url`).

### Enabling signing + notarization later

Add these secrets — `forge.config.ts` wires `osxSign`/`osxNotarize` in
automatically when they are present, no workflow change needed:

- `APPLE_SIGNING_IDENTITY` — Developer ID Application certificate
- `APPLE_API_KEY` / `APPLE_API_KEY_ID` / `APPLE_API_ISSUER` — App Store Connect
  API key for notarization
