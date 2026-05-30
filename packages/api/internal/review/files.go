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

// ErrBlobNotFound is returned by ReadRepositoryBlob when the requested file
// is missing from the working tree or absent at the given ref.
var ErrBlobNotFound = errors.New("blob not found")

// ErrBlobTooLarge is returned by ReadRepositoryBlob when the requested file
// exceeds maxFileReadBytes; callers should map this to HTTP 413.
var ErrBlobTooLarge = errors.New("blob too large")

func ResolveRoot(ctx context.Context, launchPath string) (string, error) {
	return RootFor(ctx, launchPath)
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

	rootRaw, err := RootFor(ctx, launchPath)

	if err != nil {
		return RepositoryFile{}, err
	}

	root, err := filepath.EvalSymlinks(rootRaw)

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

// ReadRepositoryFileRange returns lines [startLine, endLine] (1-indexed).
// An empty ref reads the working tree; otherwise `git show ref:path`.
// Capped at maxFileRangeLines.
func ReadRepositoryFileRange(
	ctx context.Context,
	launchPath, relPath, ref string,
	startLine, endLine int,
) (RepositoryFileRange, error) {
	if relPath == "" {
		return RepositoryFileRange{}, errors.New("path is required")
	}

	if startLine < 1 {
		startLine = 1
	}

	if endLine < startLine {
		return RepositoryFileRange{}, fmt.Errorf("endLine must be >= startLine")
	}

	if endLine-startLine+1 > maxFileRangeLines {
		endLine = startLine + maxFileRangeLines - 1
	}

	rootRaw, err := RootFor(ctx, launchPath)

	if err != nil {
		return RepositoryFileRange{}, err
	}

	root, err := filepath.EvalSymlinks(rootRaw)

	if err != nil {
		return RepositoryFileRange{}, fmt.Errorf("resolve git root: %w", err)
	}

	clean := filepath.Clean(filepath.FromSlash(relPath))

	if filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return RepositoryFileRange{}, fmt.Errorf("invalid path: %s", relPath)
	}

	var content []byte

	if ref == "" {
		fullPath := filepath.Join(root, clean)
		resolved, err := filepath.EvalSymlinks(fullPath)

		if err != nil {
			return RepositoryFileRange{}, fmt.Errorf("read file: %w", err)
		}

		rel, err := filepath.Rel(root, resolved)

		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return RepositoryFileRange{}, fmt.Errorf("path escapes repository root: %s", relPath)
		}

		info, err := os.Stat(resolved)

		if err != nil {
			return RepositoryFileRange{}, fmt.Errorf("stat file: %w", err)
		}

		if info.IsDir() {
			return RepositoryFileRange{}, fmt.Errorf("path is a directory: %s", relPath)
		}

		content, err = os.ReadFile(resolved)

		if err != nil {
			return RepositoryFileRange{}, fmt.Errorf("read file: %w", err)
		}
	} else {
		raw, err := gitBytes(ctx, root, "show", ref+":"+filepath.ToSlash(clean))

		if err != nil {
			return RepositoryFileRange{}, fmt.Errorf("git show %s: %w", ref, err)
		}

		content = raw
	}

	if !utf8.Valid(content) {
		return RepositoryFileRange{}, fmt.Errorf("file is not valid UTF-8: %s", relPath)
	}

	allLines := strings.Split(string(content), "\n")

	// A trailing newline produces a final empty element; drop it so
	// totalLines matches what users see in an editor.
	totalLines := len(allLines)

	if totalLines > 0 && allLines[totalLines-1] == "" {
		totalLines--
	}

	if startLine > totalLines {
		return RepositoryFileRange{
			Path:      filepath.ToSlash(clean),
			StartLine: startLine,
			EndLine:   startLine - 1,
			Lines:     []string{},
			EOF:       true,
		}, nil
	}

	if endLine > totalLines {
		endLine = totalLines
	}

	lines := make([]string, 0, endLine-startLine+1)

	for i := startLine - 1; i < endLine; i++ {
		lines = append(lines, allLines[i])
	}

	return RepositoryFileRange{
		Path:      filepath.ToSlash(clean),
		StartLine: startLine,
		EndLine:   endLine,
		Lines:     lines,
		EOF:       endLine >= totalLines,
	}, nil
}

// ReadRepositoryBlob returns the raw bytes of a file at the given git ref —
// the working tree when ref is empty, the index when ref is ":0", otherwise
// `git show ref:path`. Enforces the same path-traversal guards and 2 MB cap
// as ReadRepositoryFile. Missing blobs yield ErrBlobNotFound; oversize blobs
// yield ErrBlobTooLarge.
func ReadRepositoryBlob(ctx context.Context, launchPath, relPath, ref string) ([]byte, error) {
	if relPath == "" {
		return nil, errors.New("path is required")
	}

	rootRaw, err := RootFor(ctx, launchPath)

	if err != nil {
		return nil, err
	}

	root, err := filepath.EvalSymlinks(rootRaw)

	if err != nil {
		return nil, fmt.Errorf("resolve git root: %w", err)
	}

	clean := filepath.Clean(filepath.FromSlash(relPath))

	if filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return nil, fmt.Errorf("invalid path: %s", relPath)
	}

	if ref == "" {
		fullPath := filepath.Join(root, clean)
		resolved, err := filepath.EvalSymlinks(fullPath)

		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return nil, ErrBlobNotFound
			}

			return nil, fmt.Errorf("read file: %w", err)
		}

		rel, err := filepath.Rel(root, resolved)

		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return nil, fmt.Errorf("path escapes repository root: %s", relPath)
		}

		info, err := os.Stat(resolved)

		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return nil, ErrBlobNotFound
			}

			return nil, fmt.Errorf("stat file: %w", err)
		}

		if info.IsDir() {
			return nil, fmt.Errorf("path is a directory: %s", relPath)
		}

		if info.Size() > maxFileReadBytes {
			return nil, ErrBlobTooLarge
		}

		return os.ReadFile(resolved)
	}

	raw, err := gitBytes(ctx, root, "show", ref+":"+filepath.ToSlash(clean))

	if err != nil {
		// git show returns nonzero when the blob does not exist at this ref;
		// treat every git-show failure here as a missing blob to keep the
		// surface small. Real errors (corrupt repo, bad ref) are vanishingly
		// rare in the read path and would also produce a not-found UX.
		return nil, ErrBlobNotFound
	}

	if int64(len(raw)) > maxFileReadBytes {
		return nil, ErrBlobTooLarge
	}

	return raw, nil
}
