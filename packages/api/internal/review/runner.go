package review

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// GitExecutor runs `git` against a working tree. The default implementation
// shells out to the on-PATH `git` binary; tests swap it via SetGit to
// avoid real subprocesses.
type GitExecutor interface {
	Output(ctx context.Context, dir string, args ...string) (string, error)
	Bytes(ctx context.Context, dir string, args ...string) ([]byte, error)
}

// GhExecutor runs the GitHub CLI. Available reports whether the binary is
// installed; Output runs it inside a working tree.
type GhExecutor interface {
	Available() bool
	Output(ctx context.Context, dir string, args ...string) ([]byte, error)
}

type execGit struct{}

type execGh struct{}

func (execGit) Output(ctx context.Context, dir string, args ...string) (string, error) {
	output, err := execGit{}.Bytes(ctx, dir, args...)

	return string(output), err
}

func (execGit) Bytes(ctx context.Context, dir string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	output, err := cmd.Output()

	if err != nil {
		var exit *exec.ExitError

		if errors.As(err, &exit) {
			return nil, fmt.Errorf("git %s: %s", strings.Join(args, " "), strings.TrimSpace(string(exit.Stderr)))
		}

		return nil, err
	}

	return output, nil
}

func (execGh) Available() bool {
	_, err := exec.LookPath("gh")

	return err == nil
}

func (execGh) Output(ctx context.Context, dir string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "gh", args...)
	cmd.Dir = dir
	output, err := cmd.Output()

	if err != nil {
		var exit *exec.ExitError

		if errors.As(err, &exit) {
			return nil, fmt.Errorf("gh %s: %s", strings.Join(args, " "), strings.TrimSpace(string(exit.Stderr)))
		}

		return nil, err
	}

	return output, nil
}

// Default executors are exposed as package-level variables so tests can
// swap them in process. Use SetGit / SetGh for scoped overrides that
// restore on the returned func.
var (
	gitExec GitExecutor = execGit{}
	ghExec  GhExecutor  = execGh{}
)

// SetGit swaps the active git executor and returns a restore func.
func SetGit(executor GitExecutor) (restore func()) {
	prev := gitExec
	gitExec = executor

	return func() { gitExec = prev }
}

// SetGh swaps the active gh executor and returns a restore func.
func SetGh(executor GhExecutor) (restore func()) {
	prev := ghExec
	ghExec = executor

	return func() { ghExec = prev }
}

func gitOutput(ctx context.Context, dir string, args ...string) (string, error) {
	return gitExec.Output(ctx, dir, args...)
}

func gitBytes(ctx context.Context, dir string, args ...string) ([]byte, error) {
	return gitExec.Bytes(ctx, dir, args...)
}

func hasGh(_ context.Context) bool {
	return ghExec.Available()
}

func ghOutput(ctx context.Context, dir string, args ...string) ([]byte, error) {
	return ghExec.Output(ctx, dir, args...)
}
