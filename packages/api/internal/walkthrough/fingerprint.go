package walkthrough

import (
	"crypto/sha1"
	"encoding/hex"

	"github.com/gocanto/git-diff/internal/review"
)

// schemaVersion bumps whenever the Walkthrough shape changes in a way
// that invalidates cached rows. Mixed into the fingerprint so old caches
// are forced to regenerate instead of round-tripping a stale shape.
const schemaVersion = "v2"

// FingerprintForState returns a deterministic hash of the file paths,
// per-file fingerprints, AND the current schema version. Two
// RepositoryStates with the same fingerprint should yield the same
// walkthrough; this powers the cache key.
//
// Pure function — no IO. Public so the service layer can compare against
// the cached row's fingerprint without re-implementing the rule.
func FingerprintForState(state review.RepositoryState) string {
	return FingerprintForStateAndProvider(state, "")
}

// FingerprintForStateAndProvider mixes the providerID into the hash so
// switching providers (anthropic ↔ codex) invalidates the cache. Pass an
// empty providerID to fall back to the legacy (state-only) hash.
func FingerprintForStateAndProvider(state review.RepositoryState, providerID string) string {
	hash := sha1.New()

	hash.Write([]byte(schemaVersion))
	hash.Write([]byte(providerID))
	hash.Write([]byte(state.Mode))
	hash.Write([]byte(state.CommitSHA))
	hash.Write([]byte(state.HeadSHA))

	for _, file := range state.Files {
		hash.Write([]byte(file.Path))
		hash.Write([]byte(file.Fingerprint))
	}

	return hex.EncodeToString(hash.Sum(nil))
}
