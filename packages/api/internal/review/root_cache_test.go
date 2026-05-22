package review

import (
	"context"
	"sync/atomic"
	"testing"
)

type countingGit struct {
	calls    *atomic.Int32
	fail     bool
	delegate GitExecutor
}

func TestRootForCachesPerRequest(t *testing.T) {
	root := t.TempDir()
	seedRepo(t, root)

	var calls atomic.Int32

	restore := SetGit(&countingGit{calls: &calls, delegate: execGit{}})

	defer restore()

	ctx := WithRootCache(context.Background())

	for i := 0; i < 3; i++ {
		if _, err := RootFor(ctx, root); err != nil {
			t.Fatalf("RootFor #%d: %v", i, err)
		}
	}

	if got := calls.Load(); got != 1 {
		t.Fatalf("expected 1 git call, got %d (cache miss?)", got)
	}
}

func TestRootForWithoutCacheStillWorks(t *testing.T) {
	root := t.TempDir()
	seedRepo(t, root)

	if _, err := RootFor(context.Background(), root); err != nil {
		t.Fatalf("uncached RootFor: %v", err)
	}
}

func TestWithRootCacheIdempotent(t *testing.T) {
	ctx := WithRootCache(context.Background())
	ctx2 := WithRootCache(ctx)

	if ctx2.Value(rootCacheKey{}) != ctx.Value(rootCacheKey{}) {
		t.Fatal("WithRootCache should not replace an installed cache")
	}
}

func TestRootForCachesErrorsToAvoidRetryStorm(t *testing.T) {
	var calls atomic.Int32

	restore := SetGit(&countingGit{calls: &calls, fail: true})

	defer restore()

	ctx := WithRootCache(context.Background())

	for i := 0; i < 5; i++ {
		if _, err := RootFor(ctx, "/nonexistent"); err == nil {
			t.Fatal("expected error")
		}
	}

	if got := calls.Load(); got != 1 {
		t.Fatalf("error path should also cache; got %d calls", got)
	}
}

func (g *countingGit) Output(ctx context.Context, dir string, args ...string) (string, error) {
	g.calls.Add(1)

	if g.fail {
		return "", context.DeadlineExceeded
	}

	return g.delegate.Output(ctx, dir, args...)
}

func (g *countingGit) Bytes(ctx context.Context, dir string, args ...string) ([]byte, error) {
	g.calls.Add(1)

	if g.fail {
		return nil, context.DeadlineExceeded
	}

	return g.delegate.Bytes(ctx, dir, args...)
}
