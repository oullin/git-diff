package review

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// BenchmarkReadRepositoryState exercises the batched-diff path on a
// working-tree with 50 modified files. Comparing before/after the perf
// pass should show a multi-x speedup since the old per-file loop fired
// 100 separate `git diff` subprocesses.
//
// Run with: go test -bench BenchmarkReadRepositoryState -benchtime=3x ./internal/review/
func BenchmarkReadRepositoryState(b *testing.B) {
	root := benchSetup(b, 50)
	ctx := context.Background()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		state, err := ReadRepositoryState(ctx, root)

		if err != nil {
			b.Fatal(err)
		}

		if len(state.Files) == 0 {
			b.Fatal("expected files")
		}
	}
}

// TestReadRepositoryStateHandlesManyFiles is the correctness counterpart
// to the benchmark — verifies the batched splitter actually returns one
// section per dirty file with the right paths.
func TestReadRepositoryStateHandlesManyFiles(t *testing.T) {
	root := t.TempDir()
	git(t, root, "init")
	git(t, root, "config", "user.email", "bench@example.com")
	git(t, root, "config", "user.name", "Bench")
	git(t, root, "config", "commit.gpgsign", "false")

	const files = 25

	for i := 0; i < files; i++ {
		writeFile(t, root, fmt.Sprintf("file_%02d.txt", i), "seed\n")
	}

	git(t, root, "add", ".")
	git(t, root, "commit", "-m", "seed")

	for i := 0; i < files; i++ {
		writeFile(t, root, fmt.Sprintf("file_%02d.txt", i), "seed\nchanged\n")
	}

	state, err := ReadRepositoryState(context.Background(), root)

	if err != nil {
		t.Fatal(err)
	}

	if len(state.Files) != files {
		t.Fatalf("got %d files, want %d", len(state.Files), files)
	}

	for _, f := range state.Files {
		if len(f.Sections) == 0 {
			t.Fatalf("%s: no sections — splitter lookup failed", f.Path)
		}

		if f.Sections[0].Patch == "" {
			t.Fatalf("%s: empty patch", f.Path)
		}

		if f.Additions == 0 {
			t.Fatalf("%s: zero additions", f.Path)
		}
	}
}

func benchSetup(b *testing.B, files int) string {
	b.Helper()
	root := b.TempDir()

	runGit := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = root

		if out, err := cmd.CombinedOutput(); err != nil {
			b.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}

	write := func(path, content string) {
		full := filepath.Join(root, path)

		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			b.Fatal(err)
		}

		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			b.Fatal(err)
		}
	}

	runGit("init")
	runGit("config", "user.email", "bench@example.com")
	runGit("config", "user.name", "Bench")
	runGit("config", "commit.gpgsign", "false")

	for i := 0; i < files; i++ {
		write(fmt.Sprintf("file_%02d.txt", i), "seed\n")
	}

	runGit("add", ".")
	runGit("commit", "-m", "seed")

	for i := 0; i < files; i++ {
		write(fmt.Sprintf("file_%02d.txt", i), "seed\nchanged\n")
	}

	return root
}
