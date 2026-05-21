package review

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"strings"
)

func countPatchLines(patch string) (int, int) {
	additions := 0
	deletions := 0

	for _, line := range strings.Split(patch, "\n") {
		switch {
		case strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "---"):
			continue
		case strings.HasPrefix(line, "+"):
			additions++
		case strings.HasPrefix(line, "-"):
			deletions++
		}
	}

	return additions, deletions
}

func isBinaryPatch(patch string) bool {
	return strings.Contains(patch, "Binary files ") || strings.Contains(patch, "GIT binary patch")
}

func fingerprint(file ChangedFile) string {
	hash := sha1.New()
	hash.Write([]byte(file.Path))
	hash.Write([]byte(file.OldPath))
	hash.Write([]byte(file.Status))

	for _, section := range file.Sections {
		hash.Write([]byte(section.Kind))
		hash.Write([]byte(section.Patch))
	}

	return hex.EncodeToString(hash.Sum(nil))
}

func (file ChangedFile) pathSectionID(kind string) string {
	return fmt.Sprintf("%s:%s", kind, file.Path)
}
