package app

import (
	"fmt"
	"io"
	"os"
	"runtime"

	apphttpx "github.com/gocanto/git-diff/internal/app/httpx"
	"github.com/gocanto/git-diff/internal/app/migratex"
	"github.com/gocanto/git-diff/internal/app/setting"
	"github.com/gocanto/git-diff/internal/command"
)

type app struct {
	home     string
	repo     string
	settings setting.RuntimeSettings
	goos     string
	goarch   string
	stdout   io.Writer
	stderr   io.Writer
	stdin    io.Reader
	runner   command.Runner
}

func Run(args []string) int {
	home, err := os.UserHomeDir()

	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot find home directory: %v\n", err)

		return 1
	}

	repo, err := os.Getwd()

	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot find working directory: %v\n", err)

		return 1
	}

	command.EnsureSystemPath()

	a := newApp(home, findRepoRoot(repo), os.Stdin, os.Stdout, os.Stderr, command.RealRunner{})

	return a.run(args)
}

func newApp(home, repo string, stdin io.Reader, stdout, stderr io.Writer, runner command.Runner) app {
	return app{
		home:     home,
		repo:     repo,
		settings: setting.DefaultRuntimeSettings(home, repo),
		goos:     runtime.GOOS,
		goarch:   runtime.GOARCH,
		stdout:   stdout,
		stderr:   stderr,
		stdin:    stdin,
		runner:   runner,
	}
}

func (a app) run(args []string) int {
	if len(args) == 0 {
		a.usage()

		return 0
	}

	switch args[0] {
	case "help", "-h", "--help":
		a.usage()

		return 0
	case "serve-http":
		return apphttpx.Serve(args[1:], apphttpx.ServeConfig{
			Home:   a.home,
			Repo:   a.repo,
			Stderr: a.stderr,
		})
	case "migrate":
		return migratex.Run(args[1:], migratex.Config{
			Home:   a.home,
			Stdout: a.stdout,
			Stderr: a.stderr,
		})
	default:
		fmt.Fprintf(a.stderr, "unknown command %q\n\n", args[0])
		a.usage()

		return 2
	}
}
