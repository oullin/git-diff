# Distribution

This document covers how the macOS app is built, signed, notarized, and
distributed. Today the packaging story is mid-migration: both
**electron-builder** (legacy) and **Electron Forge** (target) configs live in
the repo so we can cut over without a flag day.

## Status

| Surface | State |
|---|---|
| `electron-builder` (current) | `packages/ui/package.json` "build" block; `pnpm dist:mac:unsigned` / `dist:mac:signed` |
| `electron-forge` (target) | `packages/ui/forge.config.cjs`; not yet wired into `pnpm` scripts |
| GitHub Releases publishing | Automated on `v*` tags via `.github/workflows/release.yml` (unsigned `electron-builder`); Forge publisher also configured |
| Auto-update | Not wired |
| Homebrew cask | Template at `packages/ui/scripts/Casks/git-diff.rb`; release workflow opens a bump PR to `oullin/homebrew-tap` (tap repo must be created once) |
| Terminal helper | Implemented — menu item "Install Terminal Helper…" writes a launcher script |

## Cutting over from electron-builder to Forge

1. Add Forge CLI + makers + publisher as devDeps:
   ```
   pnpm -C packages/ui add -D @electron-forge/cli @electron-forge/maker-dmg \
     @electron-forge/maker-zip @electron-forge/publisher-github
   ```
2. Replace the `dist:mac:*` scripts with:
   ```json
   "package": "electron-forge package",
   "make": "electron-forge make",
   "make:mac": "electron-forge make --platform=darwin --arch=arm64",
   "publish": "electron-forge publish"
   ```
3. Remove the `"build"` block from `package.json` and uninstall
   `electron-builder`.
4. Verify a signed/notarized DMG builds locally with `pnpm make:mac` and
   that `pnpm publish` uploads to GitHub Releases under `oullin/git-diff`.

## Bundled Go API binary

The macOS app ships with the Go API binary embedded as a Forge
`extraResource`. The pipeline is:

1. `go build` produces `packages/api/dist/api`.
2. `forge.config.cjs` lists that path under `packagerConfig.extraResource`,
   which copies it into `Git Diff Review.app/Contents/Resources/api`.
3. At runtime, `packages/ui/electron/bridge.ts` resolves the binary via
   `join(process.resourcesPath, "api")` when `app.isPackaged` is true,
   and falls back to `go run ./cmd` in development.

The binary must be built before `electron-forge package` runs; otherwise
the packaged app launches without a working bridge. The release workflow
in `.github/workflows/release.yml` (planned) sequences this correctly.

## Auto-update wiring

When we land Forge:

```ts
// packages/ui/electron/main.ts
import { updateElectronApp } from "update-electron-app";
if (app.isPackaged) updateElectronApp({ updateInterval: "6 hours" });
```

`update-electron-app` reads the GitHub Releases feed declared in the Forge
publisher config and triggers Squirrel.Mac restarts.

## Homebrew tap

1. Create a new repository `oullin/homebrew-tap` (the user has to do this;
   tap repos can't be created from a PR).
2. Add `Casks/git-diff.rb` to the tap. The file in
   `packages/ui/scripts/Casks/git-diff.rb` is the source of truth; the
   release pipeline substitutes `version` and `sha256` from the published
   DMG.
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

`.github/workflows/release.yml` runs on a `v*` tag and:

1. Checks out the repo and sets up pnpm, Node 23, and Go (from `packages/api/go.mod`).
2. Builds the Go API binary into `packages/api/dist/api`.
3. Builds the renderer + electron bundles (`pnpm -C packages/ui run build`).
4. Builds the **unsigned** DMG + ZIP (`pnpm -C packages/ui run dist:mac:unsigned`),
   computes `SHASUMS256.txt`, and publishes a GitHub Release on the tag with the
   default `GITHUB_TOKEN`.
5. On a final (non-prerelease) tag, opens a PR to `oullin/homebrew-tap` bumping the
   cask `version` + `sha256`. Requires a `HOMEBREW_TAP_TOKEN` secret (a PAT with
   write access to the tap — the default `GITHUB_TOKEN` cannot push cross-repo).

The build ships **arm64-only and unsigned** today. Tag/artifact convention:
tag `v{version}`, DMG `git-diff-review-{version}-arm64.dmg` (matches the
`artifactName` in `packages/ui/package.json` and the cask `url`).

### Enabling signing + notarization later

Add these secrets and switch step 4 to `dist:mac:signed`:

- `APPLE_SIGNING_IDENTITY` — Developer ID Application certificate
- `APPLE_API_KEY` / `APPLE_API_KEY_ID` / `APPLE_API_ISSUER` — App Store Connect
  API key for notarization

A `mac.notarize` block must also be added to the `electron-builder` `build`
config (the Forge config already handles notarization via `osxNotarize`).
