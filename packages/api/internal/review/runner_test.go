package review

import (
	"context"
	"errors"
	"testing"
)

type fakeGit struct {
	calls  [][]string
	output string
	err    error
}

// After restore the global executor is the real `git`; the fake's call
// log must not grow further from the next gitOutput call against the
// fake (we can't easily call real git here without a repo, so just
// assert restore returned a callable func).

type fakeGh struct {
	available bool
	output    []byte
	err       error
}

func (f *fakeGit) Output(ctx context.Context, dir string, args ...string) (string, error) {
	f.calls = append(f.calls, append([]string{dir}, args...))

	return f.output, f.err
}

func (f *fakeGit) Bytes(ctx context.Context, dir string, args ...string) ([]byte, error) {
	out, err := f.Output(ctx, dir, args...)

	return []byte(out), err
}

func TestSetGitSwapsAndRestores(t *testing.T) {
	fake := &fakeGit{output: "swapped"}
	restore := SetGit(fake)

	got, err := gitOutput(context.Background(), "/tmp", "rev-parse", "HEAD")

	if err != nil {
		t.Fatalf("gitOutput: %v", err)
	}

	if got != "swapped" {
		t.Fatalf("expected swapped output, got %q", got)
	}

	if len(fake.calls) != 1 {
		t.Fatalf("expected one call, got %d", len(fake.calls))
	}

	if fake.calls[0][0] != "/tmp" || fake.calls[0][1] != "rev-parse" {
		t.Fatalf("unexpected args: %v", fake.calls[0])
	}

	restore()

}

func TestSetGitPropagatesError(t *testing.T) {
	wantErr := errors.New("boom")
	restore := SetGit(&fakeGit{err: wantErr})

	defer restore()

	if _, err := gitOutput(context.Background(), "/tmp", "status"); !errors.Is(err, wantErr) {
		t.Fatalf("expected %v, got %v", wantErr, err)
	}
}

func (f *fakeGh) Available() bool { return f.available }

func (f *fakeGh) Output(ctx context.Context, dir string, args ...string) ([]byte, error) {
	return f.output, f.err
}

func TestSetGhSwapsAndRestores(t *testing.T) {
	fake := &fakeGh{available: true, output: []byte("hi")}
	restore := SetGh(fake)

	if !hasGh(context.Background()) {
		t.Fatalf("expected hasGh true")
	}

	out, err := ghOutput(context.Background(), "/tmp", "pr", "list")

	if err != nil {
		t.Fatalf("ghOutput: %v", err)
	}

	if string(out) != "hi" {
		t.Fatalf("expected 'hi', got %q", string(out))
	}

	restore()
}
