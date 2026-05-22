package review

import (
	"context"
	"fmt"
	"strings"
)

// ResolveCommitRef converts a user-supplied commit-ish (raw SHA, HEAD~3,
// HEAD^, @{1}, ...) into the canonical 40-char SHA reported by git. Returns
// an error when git cannot resolve the ref or when the input is empty.
//
// The function is the single source of truth for ref resolution in the
// review package so callers don't reimplement the same `rev-parse --verify`
// dance. Once the user-config layer (phase 3) lands, the resolved SHA also
// keys per-request caches owned by other services.
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

// ResolveParentRef returns the first parent SHA of the given commit-ish, or
// the empty string when the commit has no parent (root commits). Useful when
// callers need a diff base for `<parent>..<sha>` ranges.
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
