# Development

How to set up the toolchain, run Git Diff Review locally, and work in the
monorepo. For the product overview and install instructions, see the
[README](../README.md). For packaging and releases, see
[distribution.md](distribution.md).

## Requirements

- macOS
- Git
- Go
- Node.js with Corepack
- pnpm `10.33.0`

Enable pnpm through Corepack if needed:

```sh
corepack enable
corepack prepare pnpm@10.33.0 --activate
```

## Quick start

From the repository root:

```sh
pnpm install
pnpm build
pnpm --filter ui start
```

## Active development

```sh
pnpm dev
```

The dev command starts the Vite renderer, compiles Electron, starts the Go
backend, waits for the local socket API, and launches Electron.

## Backend CLI

The Electron app normally starts the backend automatically. To inspect the
backend directly:

```sh
cd packages/api
go run ./cmd help
go run ./cmd serve-http --socket /tmp/git-diff.sock
```

To point the Electron app at an already-running backend, set `API_BRIDGE_SOCKET`
to the socket path before launching it.

## Development commands

```sh
pnpm build
pnpm test
pnpm lint
pnpm format:check
pnpm check
```

Useful package-level commands:

```sh
pnpm --filter ui build
pnpm --filter ui test
pnpm --filter ui dev
go test ./packages/api/...
```

## Formatting

Format changed sources (Go + TS/Vue):

```sh
pnpm format      # or: make format
```

To format the entire repo, use `pnpm format-all` (`make format-all`). Run the
formatter after each change so commits stay clean.

## Review data

Review sessions, comments, events, and preferences are local. By default the Go
backend stores the review database under:

```text
~/Library/Application Support/git-diff/reviews.sqlite3
```

Viewed-file state is stored by the renderer per repository root.

## Release

Releases are built with [Electron Forge](https://www.electronforge.io/). Tag a
version to cut a release in CI:

```sh
git tag v0.1.1 && git push origin v0.1.1
```

The `Release` workflow builds the unsigned macOS arm64 `.dmg` + `.zip`, publishes
a GitHub Release, and opens a Homebrew cask bump PR against `oullin/homebrew-tap`
on final (non-prerelease) tags.

To build the same artifacts locally:

```sh
pnpm release:mac:unsigned   # builds the Go API + renderer, then `electron-forge make`
```

The Electron package is configured (in `packages/ui/forge.config.ts`) with:

- app id: `io.gocanto.git-diff`
- product name: `Git Diff Review`
- GitHub release target: `oullin/git-diff`

See [distribution.md](distribution.md) for the full signing, notarization, and
Homebrew release flow.
