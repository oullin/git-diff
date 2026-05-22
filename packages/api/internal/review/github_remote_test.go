package review

import (
	"testing"
)

func TestParseGitHubRemotesAcceptsCommonURLShapes(t *testing.T) {
	raw := "origin\tgit@github.com:anthropics/claude-code.git (fetch)\n" +
		"origin\tgit@github.com:anthropics/claude-code.git (push)\n" +
		"upstream\thttps://github.com/upstream/repo (fetch)\n" +
		"upstream\thttps://github.com/upstream/repo (push)\n" +
		"weird\tssh://git@github.com/some/thing.git (fetch)\n" +
		"gitlab\thttps://gitlab.com/foo/bar.git (fetch)\n"

	got := parseGitHubRemotes(raw)

	if len(got) != 3 {
		t.Fatalf("expected 3 GitHub remotes, got %d (%#v)", len(got), got)
	}

	if got[0].Name != "origin" {
		t.Fatalf("origin should come first, got %q", got[0].Name)
	}

	if got[0].Owner != "anthropics" || got[0].Repo != "claude-code" {
		t.Fatalf("origin owner/repo mismatched: %+v", got[0])
	}

	if got[0].HTMLURL() != "https://github.com/anthropics/claude-code" {
		t.Fatalf("html url = %q", got[0].HTMLURL())
	}
}

func TestParseGitHubRemotesDeduplicatesByName(t *testing.T) {
	// `git remote -v` reports fetch+push for each remote — they share a
	// name so we keep the first occurrence only.
	raw := "origin\tgit@github.com:foo/bar.git (fetch)\n" +
		"origin\tgit@github.com:foo/bar.git (push)\n"

	got := parseGitHubRemotes(raw)

	if len(got) != 1 {
		t.Fatalf("dedupe failed, got %d entries", len(got))
	}
}

func TestParseGitHubRemotesSkipsNonGitHub(t *testing.T) {
	raw := "origin\thttps://bitbucket.org/team/repo.git (fetch)\n"

	got := parseGitHubRemotes(raw)

	if len(got) != 0 {
		t.Fatalf("expected no entries, got %#v", got)
	}
}

func TestParseGitHubRemotesPromotesOriginEvenWhenLast(t *testing.T) {
	raw := "upstream\tgit@github.com:up/repo.git (fetch)\n" +
		"origin\tgit@github.com:user/fork.git (fetch)\n"

	got := parseGitHubRemotes(raw)

	if len(got) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(got))
	}

	if got[0].Name != "origin" {
		t.Fatalf("origin should be promoted to first, got order %q,%q", got[0].Name, got[1].Name)
	}
}

func TestGitHubRemoteHTMLURLEmptyWhenIncomplete(t *testing.T) {
	if (GitHubRemote{Name: "origin"}).HTMLURL() != "" {
		t.Fatal("incomplete remote should return empty url")
	}
}
