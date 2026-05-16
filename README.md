# Git Diff Review

Git Diff Review is a native macOS app for reviewing local Git changes before
they are committed. It gives a GitHub-style review surface for a working tree:
file navigation, staged/unstaged/untracked diffs, split or unified views,
viewed-file tracking, review notes, and inline comments stored locally.

The desktop app is built with Electron, Vue, and TypeScript. A Go backend reads
repository state through Git, serves a local HTTP API over a Unix socket, and
persists review sessions in SQLite.

## What It Does

- opens any local Git repository;
- shows changed files with additions, deletions, status, and rename metadata;
- renders staged, unstaged, and untracked patches;
- supports split and unified diff modes;
- can hide whitespace-only diff noise;
- tracks viewed files per repository in local storage;
- creates local review sessions with rich-text summaries;
- stores inline review comments and timeline events in SQLite;
- runs as a packaged macOS app or as a local development stack.

This is a local review tool. It does not publish comments to GitHub, create pull
requests, or modify the repository contents.

## Workspace Layout

| Path               | Role                                                    |
| ------------------ | ------------------------------------------------------- |
| `packages/ui`      | Electron/Vue desktop app and renderer UI.               |
| `packages/bridge`  | TypeScript HTTP client used by Electron.                |
| `packages/macbook` | Go backend that reads Git diffs and stores reviews.     |
| `packages/tools`   | Turbo cache wrapper and unsigned macOS release helper.  |
| `storage/`         | Local build caches and generated runtime data.          |

Some internal package names still use the older `git-diff` / `macbook` naming.
The app surface and release target are `Git Diff Review`.

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

## Quick Start

From the repository root:

```sh
pnpm install
pnpm build
pnpm --filter ui start
```

For active development:

```sh
pnpm dev
```

The dev command starts the Vite renderer, compiles Electron, starts the Go
backend, waits for the local socket API, and launches Electron.

## Backend CLI

The Electron app normally starts the backend automatically. To inspect the
backend directly:

```sh
cd packages/macbook
go run ./cmd help
go run ./cmd serve-http --socket /tmp/git-diff.sock
```

To point the Electron app at an already-running backend, set
`API_BRIDGE_SOCKET` to the socket path before launching it.

## Development Commands

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
pnpm --filter macbook test
```

## Review Data

Review sessions, comments, events, and preferences are local. By default the Go
backend stores the review database under:

```text
~/Library/Application Support/git-diff/reviews.sqlite3
```

Viewed-file state is stored by the renderer per repository root.

## Release

Build an unsigned macOS release artifact with:

```sh
pnpm release:mac:unsigned
```

The Electron package is configured with:

- app id: `io.gocanto.git-diff`
- product name: `Git Diff Review`
- GitHub release target: `gocanto/git-diff`

## License

See [LICENSE](LICENSE).
