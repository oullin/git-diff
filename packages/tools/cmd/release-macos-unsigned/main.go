package main

import (
	"context"
	"fmt"
	"os"

	"github.com/oullin/git-diff/tools/internal/release"
	"github.com/oullin/git-diff/tools/internal/reporoot"
)

func main() {
	config, err := release.LoadConfig(os.Args[1:])

	if err != nil {
		fail(err)
	}

	rootDir, err := reporoot.Find("")

	if err != nil {
		fail(err)
	}

	tool := release.Tool{
		RootDir: rootDir,
		Config:  config,
	}

	if err := tool.Run(context.Background()); err != nil {
		fail(err)
	}
}

func fail(err error) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", err)
	fmt.Fprintln(os.Stderr, "Usage: release-macos-unsigned --notes-file <path> [--repo owner/name] [--tag tag]")
	os.Exit(1)
}
