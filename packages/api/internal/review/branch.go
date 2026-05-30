package review

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

func ListBranches(ctx context.Context, launchPath string) ([]string, error) {
	root, err := RootFor(ctx, launchPath)

	if err != nil {
		return nil, err
	}

	raw, err := gitOutput(ctx, root, "for-each-ref", "--format=%(refname:short)", "refs/heads")

	if err != nil {
		return nil, fmt.Errorf("list branches: %w", err)
	}

	lines := strings.Split(raw, "\n")
	branches := make([]string, 0, len(lines))

	for _, line := range lines {
		name := strings.TrimSpace(line)

		if name != "" {
			branches = append(branches, name)
		}
	}

	return branches, nil
}

func CheckoutBranch(ctx context.Context, launchPath string, branch string) error {
	if err := validateBranchName(branch); err != nil {
		return err
	}

	root, err := RootFor(ctx, launchPath)

	if err != nil {
		return err
	}

	statusRaw, err := gitBytes(ctx, root, "status", "--porcelain=v1", "-z")

	if err != nil {
		return fmt.Errorf("read git status: %w", err)
	}

	if dirty := dirtyPaths(parseStatus(statusRaw)); len(dirty) > 0 {
		return &WorkingTreeDirtyError{Files: dirty}
	}

	if _, err := gitOutput(ctx, root, "checkout", branch); err != nil {
		return fmt.Errorf("checkout branch: %w", err)
	}

	return nil
}

func CreateBranch(ctx context.Context, launchPath string, name string) error {
	if err := validateBranchName(name); err != nil {
		return err
	}

	root, err := RootFor(ctx, launchPath)

	if err != nil {
		return err
	}

	if _, err := gitOutput(ctx, root, "check-ref-format", "--branch", name); err != nil {
		return fmt.Errorf("invalid branch name: %w", err)
	}

	if _, err := gitOutput(ctx, root, "checkout", "-b", name); err != nil {
		return fmt.Errorf("create branch: %w", err)
	}

	return nil
}

func DeleteBranch(ctx context.Context, launchPath string, name string) error {
	if err := validateBranchName(name); err != nil {
		return err
	}

	root, err := RootFor(ctx, launchPath)

	if err != nil {
		return err
	}

	current, err := gitOutput(ctx, root, "branch", "--show-current")

	if err == nil && strings.TrimSpace(current) == name {
		return fmt.Errorf("cannot delete the currently checked-out branch")
	}

	if _, err := gitOutput(ctx, root, "branch", "-D", name); err != nil {
		return fmt.Errorf("delete branch: %w", err)
	}

	return nil
}

func validateBranchName(name string) error {
	name = strings.TrimSpace(name)

	if name == "" {
		return errors.New("branch name is required")
	}

	if strings.HasPrefix(name, "-") {
		return errors.New("branch name must not start with '-'")
	}

	if strings.ContainsAny(name, " \t\n\r") {
		return errors.New("branch name must not contain whitespace")
	}

	if strings.Contains(name, "..") {
		return errors.New("branch name must not contain '..'")
	}

	return nil
}
