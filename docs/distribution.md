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
| GitHub Releases publishing | Manual via `electron-builder` today; Forge publisher configured |
| Auto-update | Not wired |
| Homebrew cask | Template at `packages/ui/scripts/Casks/git-diff.rb`; tap repo not yet created |
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

## Release workflow (planned)

`.github/workflows/release.yml` on a `v*` tag will:

1. Check out the repo + setup Node 23 + Go.
2. Build the Go API binary into `packages/api/dist/api`.
3. Run `pnpm install` and `pnpm -C packages/ui publish` with these secrets:
   - `APPLE_SIGNING_IDENTITY` — Developer ID Application
   - `APPLE_API_KEY` / `APPLE_API_KEY_ID` / `APPLE_API_ISSUER` — App Store
     Connect API key for notarization
   - `GITHUB_TOKEN` — to push to Releases
4. Open a PR to `oullin/homebrew-tap` bumping `version` + `sha256`.
