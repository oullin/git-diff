// Package walks orchestrates LLM-generated review walkthroughs.
package walks

// Action labels what the user should do with a file: review, scan, or skim.
type Action string

// Impact labels reach: wide (many call sites), contained (single feature
// surface), or mechanical (sweeping but logically trivial).
type Impact string

type FileEntry struct {
	Path   string `json:"path"`
	Note   string `json:"note"`
	Action Action `json:"action"`
	Impact Impact `json:"impact"`
}

type Group struct {
	ID        string      `json:"id"`
	Title     string      `json:"title"`
	Rationale string      `json:"rationale"`
	Files     []FileEntry `json:"files"`
}

type Walkthrough struct {
	Groups      []Group `json:"groups"`
	Summary     string  `json:"summary,omitempty"`
	ProviderID  string  `json:"providerId"`
	ModelID     string  `json:"modelId"`
	GeneratedAt string  `json:"generatedAt"`
	Fingerprint string  `json:"fingerprint"`
}

// Budget caps patch sizes when building the prompt. PerFileBytes
// truncates a single file (appending "[…truncated]"); TotalBytes drops
// the tail of the file list once the running total crosses it.
type Budget struct {
	PerFileBytes int
	TotalBytes   int
}

const (
	ActionReview Action = "review"
	ActionScan   Action = "scan"
	ActionSkim   Action = "skim"
)

const (
	ImpactWide       Impact = "wide"
	ImpactContained  Impact = "contained"
	ImpactMechanical Impact = "mechanical"
)

func DefaultBudget() Budget {
	return Budget{
		PerFileBytes: 4 * 1024,
		TotalBytes:   160 * 1024,
	}
}
