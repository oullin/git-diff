package review

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

// GitHubRemote identifies one GitHub-hosted remote configured in a repo.
// Empty when the repo has no GitHub remotes.
type GitHubRemote struct {
	Name  string // local remote name, e.g. "origin"
	Owner string // GitHub owner, e.g. "anthropics"
	Repo  string // GitHub repo, e.g. "claude-code"
	URL   string // the raw fetch URL parsed
}

// HTMLURL returns the canonical https URL for the remote, without a trailing
// `.git` suffix. Empty when Owner or Repo is unset.
func (r GitHubRemote) HTMLURL() string {
	if r.Owner == "" || r.Repo == "" {
		return ""
	}

	return fmt.Sprintf("https://github.com/%s/%s", r.Owner, r.Repo)
}

// ReadGitHubRemotes lists every GitHub remote configured in the repo at
// root, in the order `git remote -v` reports them with `origin` first when
// present. Returns an empty slice when the repo has no GitHub remotes.
//
// Extracted as a standalone helper so callers that need to resolve `#42` to
// a PR URL without shelling out to `gh` can do so. Used by the
// pull-request flow and (planned) by the CLI launch surface.
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

	// Put origin first when present so callers can `[0]` it.
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
