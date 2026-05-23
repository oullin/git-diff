package walks

import (
	"fmt"
	"strings"

	"github.com/oullin/git-diff/internal/review"
)

type PromptInput struct {
	State  review.RepositoryState
	Budget Budget
}

// BuildPrompt produces the prompt for a walkthrough call; the model is
// instructed to emit the v2 grouped JSON schema.
func BuildPrompt(input PromptInput) string {
	files := enforceBudget(input.State.Files, input.Budget)

	var b strings.Builder

	b.WriteString(instructions)
	b.WriteString("\n\nRepository: ")
	b.WriteString(input.State.Root)
	b.WriteString("\nMode: ")
	b.WriteString(input.State.Mode)

	if input.State.CommitSHA != "" {
		b.WriteString("\nCommit: ")
		b.WriteString(input.State.CommitSHA)
	}

	b.WriteString("\n\nChanged files (truncated patches):\n\n")

	for _, file := range files {
		fmt.Fprintf(&b, "----- %s (status=%s, +%d/-%d)\n",
			file.File.Path, file.File.Status, file.File.Additions, file.File.Deletions)
		b.WriteString(file.PatchBody)
		b.WriteString("\n")

		if file.OmittedAfter {
			b.WriteString("\n[…remaining files omitted to stay within the total patch budget…]\n")

			break
		}
	}

	return b.String()
}

const instructions = `You are helping a developer review a Git diff. Output ONLY a JSON object with the following shape — no prose, no code fences:

{
  "summary": "<one short sentence summarising the overall theme>",
  "groups": [
    {
      "id":        "<short kebab-case id, unique within this response>",
      "title":     "<3-6 word group title>",
      "rationale": "<one short sentence: why these files belong together>",
      "files": [
        {
          "path":   "<exact path from the diff>",
          "note":   "<one short sentence: why this file matters and what to look for>",
          "action": "review" | "scan" | "skim",
          "impact": "wide"   | "contained" | "mechanical"
        }
      ]
    }
  ]
}

Rules:
- Use ONLY paths that appear in the diff below. Never invent or rename a path.
- Every file in the diff should appear in exactly ONE group.
- Order groups by review priority (most important first). Within a group, order files by what to read first.
- "action":
    "review"  — read carefully, the change has real logic to verify.
    "scan"    — glance through; verify nothing surprising slipped in.
    "skim"    — barely look; trivial / mechanical change.
- "impact":
    "wide"        — touches many call sites or shared modules.
    "contained"   — single feature surface, low blast radius.
    "mechanical"  — sweeping but logically trivial (rename, format, dep bump).
- Output a single JSON object. No prose before or after.`
