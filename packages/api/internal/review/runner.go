package review

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

func gitOutput(ctx context.Context, dir string, args ...string) (string, error) {
	output, err := gitBytes(ctx, dir, args...)

	return string(output), err
}

func gitBytes(ctx context.Context, dir string, args ...string) ([]byte, error) {
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

func hasGh(_ context.Context) bool {
	_, err := exec.LookPath("gh")

	return err == nil
}

func ghOutput(ctx context.Context, dir string, args ...string) ([]byte, error) {
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
