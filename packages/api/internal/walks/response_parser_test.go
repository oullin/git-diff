package walks

import (
	"strings"
	"testing"
)

func TestParseAcceptsPlainJSON(t *testing.T) {
	parsed, err := Parse(`{"summary":"s","groups":[]}`)

	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if parsed.Summary != "s" {
		t.Fatalf("expected summary 's', got %q", parsed.Summary)
	}
}

func TestParseStripsCodeFences(t *testing.T) {
	raw := "```json\n{\"summary\":\"hi\",\"groups\":[]}\n```"

	parsed, err := Parse(raw)

	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if parsed.Summary != "hi" {
		t.Fatalf("expected summary 'hi', got %q", parsed.Summary)
	}
}

func TestParseStripsBareTripleBackticks(t *testing.T) {
	raw := "```\n{\"summary\":\"hi\",\"groups\":[]}\n```"

	parsed, err := Parse(raw)

	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if parsed.Summary != "hi" {
		t.Fatalf("expected 'hi', got %q", parsed.Summary)
	}
}

func TestParseErrorsOnInvalidJSON(t *testing.T) {
	_, err := Parse("not json")

	if err == nil {
		t.Fatalf("expected error on invalid JSON")
	}

	if !strings.Contains(err.Error(), "decode") {
		t.Fatalf("expected decode error message, got %q", err.Error())
	}
}

func TestParseDecodesGroupsAndFiles(t *testing.T) {
	raw := `{
		"summary": "ok",
		"groups": [
			{
				"id": "g1",
				"title": "Group",
				"rationale": "why",
				"files": [{"path": "a.go", "note": "n", "action": "review", "impact": "wide"}]
			}
		]
	}`

	parsed, err := Parse(raw)

	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	if len(parsed.Groups) != 1 || parsed.Groups[0].ID != "g1" || len(parsed.Groups[0].Files) != 1 {
		t.Fatalf("unexpected parse result: %#v", parsed)
	}

	if parsed.Groups[0].Files[0].Action != ActionReview {
		t.Fatalf("expected action=review, got %q", parsed.Groups[0].Files[0].Action)
	}
}
