// Package walkthrough orchestrates LLM-generated review walkthroughs.
// Pure orchestration — the provider abstraction in internal/ai handles
// the actual model call, and the units in this package own the prompt,
// budget enforcement, parsing, validation, and caching seams.
package walkthrough

// Action labels what the user is expected to do with a file: a full
// review, a quick scan, or a skim. The vocabulary lets the UI colour-code
// consistently.
type Action string

// Impact labels how broadly a change reaches. "wide" means many call
// sites or shared modules; "contained" means a single feature surface;
// "mechanical" means a sweeping but logically trivial change (rename,
// formatting, dep bump).
type Impact string

// FileEntry is one file's slot inside a Group. Note is the LLM's
// per-file rationale; Action/Impact let the UI render chips.
type FileEntry struct {
	Path   string `json:"path"`
	Note   string `json:"note"`
	Action Action `json:"action"`
	Impact Impact `json:"impact"`
}

// Group bundles related files with a short rationale ("why these go
// together"). The UI renders each group as a collapsible accordion.
type Group struct {
	ID        string      `json:"id"`
	Title     string      `json:"title"`
	Rationale string      `json:"rationale"`
	Files     []FileEntry `json:"files"`
}

// Walkthrough is the structured output the renderer consumes.
//
// Backwards compatibility: callers that haven't migrated yet still read
// Order/Notes — they're populated from Groups on the fly so old UI code
// keeps rendering.
type Walkthrough struct {
	Groups      []Group `json:"groups"`
	Summary     string  `json:"summary,omitempty"`
	ProviderID  string  `json:"providerId"`
	ModelID     string  `json:"modelId"`
	GeneratedAt string  `json:"generatedAt"`
	Fingerprint string  `json:"fingerprint"`
	// Legacy mirrors of Groups for one release.
	Order []string          `json:"order,omitempty"`
	Notes map[string]string `json:"notes,omitempty"`
}

// Budget controls patch truncation when building the prompt. Sourced
// from the user config's walkthrough.* keys.
type Budget struct {
	// PerFileBytes caps the patch portion of a single file. Truncated
	// patches get a "[…truncated]" marker so the LLM knows it was cut.
	PerFileBytes int

	// TotalBytes caps the sum of all patches in one prompt. When the
	// limit is reached, the remaining files are dropped (their paths
	// still appear; their patches don't).
	TotalBytes int
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

// DefaultBudget returns the default Budget used when the caller doesn't
// supply one.
func DefaultBudget() Budget {
	return Budget{
		PerFileBytes: 4 * 1024,
		TotalBytes:   160 * 1024,
	}
}
