package review

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

type GitHubRemote struct {
	Name  string
	Owner string
	Repo  string
	URL   string
}

// HTMLURL drops the .git suffix; empty when Owner or Repo is unset.
func (r GitHubRemote) HTMLURL() string {
	if r.Owner == "" || r.Repo == "" {
		return ""
	}

	return fmt.Sprintf("https://github.com/%s/%s", r.Owner, r.Repo)
}

// ReadGitHubRemotes lists every GitHub remote, with `origin` first when
// present so callers can `[0]` it.
func ReadGitHubRemotes(ctx context.Context, root string) ([]GitHubRemote, error) {
	raw, err := gitOutput(ctx, root, "remote", "-v")

	if err != nil {
		return nil, fmt.Errorf("list git remotes: %w", err)
	}

	return parseGitHubRemotes(raw), nil
}

var githubURLPattern = regexp.MustCompile(`(?i)^(?:https?://|git://|ssh://(?:git@)?|git@)github\.com[:/]([^/\s]+)/([^/\s]+?)(?:\.git)?$`)

func parseGitHubRemotes(raw string) []GitHubRemote {
	seen := make(map[string]struct{})
	out := make([]GitHubRemote, 0, 2)

	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		fields := strings.Fields(line)

		if len(fields) < 2 {
			continue
		}

		name := fields[0]
		url := fields[1]

		if _, dup := seen[name]; dup {
			continue
		}

		match := githubURLPattern.FindStringSubmatch(url)

		if match == nil {
			continue
		}

		seen[name] = struct{}{}
		out = append(out, GitHubRemote{
			Name:  name,
			Owner: match[1],
			Repo:  match[2],
			URL:   url,
		})
	}

	for i := range out {
		if out[i].Name == "origin" {
			if i != 0 {
				out[0], out[i] = out[i], out[0]
			}

			break
		}
	}

	return out
}
