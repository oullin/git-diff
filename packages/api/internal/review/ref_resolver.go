package review

import (
	"context"
	"fmt"
	"strings"
)

// ResolveCommitRef canonicalises a commit-ish (raw SHA, HEAD~3, @{1}, …)
// to its 40-char SHA via `rev-parse --verify`.
func ResolveCommitRef(ctx context.Context, root, ref string) (string, error) {
	trimmed := strings.TrimSpace(ref)

	if trimmed == "" {
		return "", fmt.Errorf("commit ref is required")
	}

	out, err := gitOutput(ctx, root, "rev-parse", "--verify", trimmed+"^{commit}")

	if err != nil {
		return "", fmt.Errorf("resolve commit ref %q: %w", trimmed, err)
	}

	return strings.TrimSpace(out), nil
}

// ResolveParentRef returns the first parent SHA, or "" for root commits.
func ResolveParentRef(ctx context.Context, root, ref string) (string, error) {
	resolved, err := ResolveCommitRef(ctx, root, ref)

	if err != nil {
		return "", err
	}

	out, err := gitOutput(ctx, root, "rev-list", "--parents", "-n", "1", resolved)

	if err != nil {
		return "", fmt.Errorf("list parents of %s: %w", resolved, err)
	}

	parts := strings.Fields(strings.TrimSpace(out))

	if len(parts) <= 1 {
		return "", nil
	}

	return parts[1], nil
}
