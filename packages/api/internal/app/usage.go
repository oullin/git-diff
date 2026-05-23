package app

import "fmt"

func (a app) usage() {
	fmt.Fprintln(a.stdout, `api serves the git-diff review backend over a Unix HTTP socket.

Usage:
  api serve-http --socket <path> [--repo-root <path>] [--db <path>]
  api migrate <up|down|version|force> [--db <path>] [args]

The Electron app spawns this binary to read repository diffs, manage review sessions, and persist comments.`)
}
