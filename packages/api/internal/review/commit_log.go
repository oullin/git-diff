package review

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

// ListCommitLog returns the most recent commits, newest first.
func ListCommitLog(ctx context.Context, launchPath string, limit int) ([]CommitSummary, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}

	root, err := RootFor(ctx, launchPath)

	if err != nil {
		return nil, err
	}

	const sep = "\x1f"
	const recordSep = "\x1e"
	format := "%H" + sep + "%h" + sep + "%an" + sep + "%ae" + sep + "%aI" + sep + "%s" + recordSep

	raw, err := gitOutput(ctx, root, "log", "--pretty=format:"+format, "-n", strconv.Itoa(limit))

	if err != nil {
		return nil, fmt.Errorf("read git log: %w", err)
	}

	commits := []CommitSummary{}

	for _, record := range strings.Split(raw, recordSep) {
		record = strings.TrimSpace(record)

		if record == "" {
			continue
		}

		parts := strings.SplitN(record, sep, 6)

		if len(parts) < 6 {
			continue
		}

		commits = append(commits, CommitSummary{
			SHA:      parts[0],
			ShortSHA: parts[1],
			Author:   parts[2],
			Email:    parts[3],
			Date:     parts[4],
			Subject:  parts[5],
		})
	}

	return commits, nil
}
