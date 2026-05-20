package review

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"
)

func ResolveRoot(ctx context.Context, launchPath string) (string, error) {
	root, err := gitOutput(ctx, launchPath, "rev-parse", "--show-toplevel")

	if err != nil {
		return "", fmt.Errorf("resolve git root: %w", err)
	}

	return strings.TrimSpace(root), nil
}

func ListRepositoryFiles(ctx context.Context, root string) ([]string, error) {
	tracked, err := gitBytes(ctx, root, "ls-files", "-z")

	if err != nil {
		return nil, err
	}

	untracked, err := gitBytes(ctx, root, "ls-files", "-z", "-o", "--exclude-standard")

	if err != nil {
		return nil, err
	}

	seen := make(map[string]struct{})
	files := make([]string, 0)

	for _, raw := range [][]byte{tracked, untracked} {
		for _, part := range bytes.Split(raw, []byte{0}) {
			if len(part) == 0 {
				continue
			}

			path := string(part)

			if _, ok := seen[path]; ok {
				continue
			}

			seen[path] = struct{}{}
			files = append(files, path)
		}
	}

	sort.Strings(files)

	return files, nil
}

func ReadRepositoryFile(ctx context.Context, launchPath, relPath string) (RepositoryFile, error) {
	if relPath == "" {
		return RepositoryFile{}, errors.New("path is required")
	}

	rootRaw, err := gitOutput(ctx, launchPath, "rev-parse", "--show-toplevel")

	if err != nil {
		return RepositoryFile{}, fmt.Errorf("resolve git root: %w", err)
	}

	root, err := filepath.EvalSymlinks(strings.TrimSpace(rootRaw))

	if err != nil {
		return RepositoryFile{}, fmt.Errorf("resolve git root: %w", err)
	}

	clean := filepath.Clean(filepath.FromSlash(relPath))

	if filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return RepositoryFile{}, fmt.Errorf("invalid path: %s", relPath)
	}

	fullPath := filepath.Join(root, clean)
	resolved, err := filepath.EvalSymlinks(fullPath)

	if err != nil {
		return RepositoryFile{}, fmt.Errorf("read file: %w", err)
	}

	rel, err := filepath.Rel(root, resolved)

	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return RepositoryFile{}, fmt.Errorf("path escapes repository root: %s", relPath)
	}

	info, err := os.Stat(resolved)

	if err != nil {
		return RepositoryFile{}, fmt.Errorf("stat file: %w", err)
	}

	if info.IsDir() {
		return RepositoryFile{}, fmt.Errorf("path is a directory: %s", relPath)
	}

	size := info.Size()

	if size > maxFileReadBytes {
		return RepositoryFile{
			Path:      filepath.ToSlash(rel),
			Binary:    false,
			Truncated: true,
			Size:      size,
		}, nil
	}

	content, err := os.ReadFile(resolved)

	if err != nil {
		return RepositoryFile{}, fmt.Errorf("read file: %w", err)
	}

	sniff := content

	if len(sniff) > 8192 {
		sniff = sniff[:8192]
	}

	if bytes.IndexByte(sniff, 0) >= 0 || !utf8.Valid(content) {
		return RepositoryFile{
			Path:   filepath.ToSlash(rel),
			Binary: true,
			Size:   size,
		}, nil
	}

	return RepositoryFile{
		Path:    filepath.ToSlash(rel),
		Content: string(content),
		Size:    size,
	}, nil
}
