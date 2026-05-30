package walks

import (
	"crypto/sha1"
	"encoding/hex"

	"github.com/oullin/git-diff/internal/review"
)

// schemaVersion is mixed into the fingerprint so cached rows from older
// Walkthrough shapes are forced to regenerate.
const schemaVersion = "v2"

// FingerprintForState falls back to the provider-agnostic hash.
func FingerprintForState(state review.RepositoryState) string {
	return FingerprintForStateAndProvider(state, "")
}

// FingerprintForStateAndProvider mixes providerID into the hash so
// switching providers invalidates the cache.
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
